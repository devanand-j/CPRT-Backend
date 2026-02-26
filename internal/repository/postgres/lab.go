package postgres

import (
	"context"
	"fmt"
	"time"

	"cprt-lis/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type LabRepository struct {
	pool *pgxpool.Pool
}

func NewLabRepository(pool *pgxpool.Pool) *LabRepository {
	return &LabRepository{pool: pool}
}

func (r *LabRepository) MarkSampleCollected(ctx context.Context, billID int64, sampleNo, collectedBy string) (domain.SampleCollectionResponse, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.SampleCollectionResponse{}, err
	}
	defer tx.Rollback(ctx)

	var patientID int64
	if err := tx.QueryRow(ctx, `SELECT patient_id FROM lab_bills WHERE id = $1`, billID).Scan(&patientID); err != nil {
		return domain.SampleCollectionResponse{}, err
	}

	var orderID int64
	err = tx.QueryRow(ctx, `SELECT id FROM lab_orders WHERE bill_id = $1 ORDER BY id DESC LIMIT 1`, billID).Scan(&orderID)
	if err != nil {
		if err := tx.QueryRow(ctx, `INSERT INTO lab_orders (order_uuid, bill_id, patient_id, order_status) VALUES (gen_random_uuid(), $1, $2, 'PENDING') RETURNING id`, billID, patientID).Scan(&orderID); err != nil {
			return domain.SampleCollectionResponse{}, err
		}
	}

	var nextWorksheetID int64
	if err := tx.QueryRow(ctx, `SELECT COALESCE(MAX(id), 0) + 1 FROM worksheets`).Scan(&nextWorksheetID); err != nil {
		nextWorksheetID = 1
	}
	worksheetNo := fmt.Sprintf("WS-%d", nextWorksheetID)

	const upsertSample = `
		INSERT INTO samples (sample_uuid, order_id, specimen_type, barcode, collected_at, status)
		VALUES (gen_random_uuid(), $1, 'Blood', $2, NOW(), 'Collected')
		ON CONFLICT (barcode)
		DO UPDATE SET collected_at = NOW(), status = 'Collected'
	`
	if _, err := tx.Exec(ctx, upsertSample, orderID, sampleNo); err != nil {
		return domain.SampleCollectionResponse{}, err
	}

	if _, err := tx.Exec(ctx, `UPDATE lab_bills SET report_status = 'Collected' WHERE id = $1`, billID); err != nil {
		return domain.SampleCollectionResponse{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.SampleCollectionResponse{}, err
	}

	return domain.SampleCollectionResponse{
		BillID:             fmt.Sprintf("%d", billID),
		SampleNo:           sampleNo,
		CollectionStatus:   "Collected",
		CollectedBy:        collectedBy,
		WorksheetNo:        worksheetNo,
		CollectionDateTime: nowISO(),
	}, nil
}

func (r *LabRepository) VerifyResults(ctx context.Context, billID int64, params []domain.ResultVerificationParam, verifiedBy string) (domain.ResultVerificationResponse, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.ResultVerificationResponse{}, err
	}
	defer tx.Rollback(ctx)

	for _, p := range params {
		abnormal := "N"
		if p.IsAbnormal {
			abnormal = "Y"
		}
		const upsertResult = `
			INSERT INTO lab_results (bill_id, param_id, param_name, result_value, abnormal_flag, result_status, verified_by_user, verified_at)
			VALUES ($1, $2, $3, $4, $5, 'Verified', $6, NOW())
			ON CONFLICT (bill_id, param_id)
			DO UPDATE SET
				param_name = EXCLUDED.param_name,
				result_value = EXCLUDED.result_value,
				abnormal_flag = EXCLUDED.abnormal_flag,
				result_status = 'Verified',
				verified_by_user = EXCLUDED.verified_by_user,
				verified_at = NOW()
		`
		if _, err := tx.Exec(ctx, upsertResult, billID, p.ParamID, p.ParamName, p.ResultValue, abnormal, verifiedBy); err != nil {
			return domain.ResultVerificationResponse{}, err
		}
	}

	if _, err := tx.Exec(ctx, `UPDATE lab_bills SET report_status = 'Verified' WHERE id = $1`, billID); err != nil {
		return domain.ResultVerificationResponse{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.ResultVerificationResponse{}, err
	}

	return domain.ResultVerificationResponse{
		BillID:               fmt.Sprintf("%d", billID),
		VerificationStatus:   "Verified",
		VerifiedBy:           verifiedBy,
		VerificationDateTime: nowISO(),
		ResultCount:          len(params),
	}, nil
}

func (r *LabRepository) CertifyResults(ctx context.Context, billID int64, certifiedBy, remarks string) (domain.ResultCertificationResponse, error) {
	const query = `
		UPDATE lab_bills
		SET
			report_status = 'Certified',
			certified_by_user = $2,
			certification_remarks = COALESCE(NULLIF($3, ''), certification_remarks),
			certified_at = NOW(),
			dispatch_ready = TRUE
		WHERE id = $1
	`
	if _, err := r.pool.Exec(ctx, query, billID, certifiedBy, remarks); err != nil {
		return domain.ResultCertificationResponse{}, err
	}

	return domain.ResultCertificationResponse{
		BillID:               fmt.Sprintf("%d", billID),
		CertificationStatus:  "Certified",
		CertifiedBy:          certifiedBy,
		CertificationRemarks: remarks,
		DispatchReady:        true,
	}, nil
}

func (r *LabRepository) GetReport(ctx context.Context, billID int64) (domain.LabReportResponse, error) {
	const headerQuery = `
		SELECT
			lb.bill_uuid::text,
			p.patient_uuid::text,
			TRIM(CONCAT(COALESCE(p.prefix, ''), ' ', COALESCE(p.first_name, ''))) AS patient_name,
			COALESCE(MAX(lr.verified_by_user), ''),
			COALESCE(lb.certified_by_user, ''),
			COALESCE(lb.report_status, 'Pending')
		FROM lab_bills lb
		JOIN patients p ON p.id = lb.patient_id
		LEFT JOIN lab_results lr ON lr.bill_id = lb.id
		WHERE lb.id = $1
		GROUP BY lb.bill_uuid, p.patient_uuid, patient_name, lb.certified_by_user, lb.report_status
	`

	var out domain.LabReportResponse
	if err := r.pool.QueryRow(ctx, headerQuery, billID).Scan(
		&out.BillID,
		&out.PatientID,
		&out.PatientName,
		&out.VerificationBy,
		&out.CertifiedBy,
		&out.ReportStatus,
	); err != nil {
		return domain.LabReportResponse{}, err
	}

	const resultQuery = `
		SELECT
			COALESCE(param_name, ''),
			COALESCE(result_value, ''),
			COALESCE(reference_range, ''),
			CASE WHEN COALESCE(abnormal_flag, 'N') = 'Y' THEN 'H' ELSE 'N' END AS flag
		FROM lab_results
		WHERE bill_id = $1
		ORDER BY id ASC
	`

	rows, err := r.pool.Query(ctx, resultQuery, billID)
	if err != nil {
		return domain.LabReportResponse{}, err
	}
	defer rows.Close()

	out.Results = make([]domain.LabReportResult, 0)
	for rows.Next() {
		var item domain.LabReportResult
		if err := rows.Scan(&item.ParamName, &item.ResultValue, &item.Reference, &item.Flag); err != nil {
			return domain.LabReportResponse{}, err
		}
		out.Results = append(out.Results, item)
	}

	return out, rows.Err()
}

func nowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}

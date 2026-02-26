package postgres

import (
	"context"

	"cprt-lis/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PatientRepository struct {
	pool *pgxpool.Pool
}

func NewPatientRepository(pool *pgxpool.Pool) *PatientRepository {
	return &PatientRepository{pool: pool}
}

func (r *PatientRepository) Create(ctx context.Context, patient domain.Patient) (domain.Patient, error) {
	const query = `
		INSERT INTO patients (
			patient_uuid, prefix, first_name, gender, age, age_unit, phone, op_ip_no, patient_type, created_by, status
		)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, COALESCE(NULLIF($10, ''), 'Active'))
		RETURNING id, patient_uuid, patient_no, status, created_at
	`
	row := r.pool.QueryRow(
		ctx,
		query,
		patient.Prefix,
		patient.FirstName,
		patient.Gender,
		patient.Age,
		patient.AgeUnit,
		patient.Phone,
		patient.OPIPNo,
		patient.PatientType,
		patient.CreatedBy,
		patient.Status,
	)
	if err := row.Scan(&patient.ID, &patient.PatientUUID, &patient.PatientNo, &patient.Status, &patient.CreatedAt); err != nil {
		return domain.Patient{}, err
	}
	return patient, nil
}

func (r *PatientRepository) SearchByQuery(ctx context.Context, queryText string) ([]domain.PatientSearchResult, error) {
	const query = `
		SELECT
			p.patient_uuid::text,
			TRIM(CONCAT(COALESCE(p.prefix, ''), ' ', COALESCE(p.first_name, ''))) AS full_name,
			COALESCE(p.op_ip_no, '') AS op_ip_no,
			COALESCE(p.phone, '') AS phone_no,
			TO_CHAR(p.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS created_at,
			COALESCE(p.created_by, '') AS created_by,
			p.patient_no,
			COALESCE(p.patient_type, '') AS patient_type,
			COALESCE(p.status, 'Active') AS status
		FROM patients p
		WHERE (
			$1 = ''
			OR p.patient_uuid::text ILIKE '%' || $1 || '%'
			OR p.first_name ILIKE '%' || $1 || '%'
			OR COALESCE(p.op_ip_no, '') ILIKE '%' || $1 || '%'
			OR COALESCE(p.phone, '') ILIKE '%' || $1 || '%'
			OR CAST(p.patient_no AS TEXT) ILIKE '%' || $1 || '%'
		)
		ORDER BY p.created_at DESC
		LIMIT 100
	`

	rows, err := r.pool.Query(ctx, query, queryText)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]domain.PatientSearchResult, 0)
	for rows.Next() {
		var item domain.PatientSearchResult
		if err := rows.Scan(
			&item.PatientID,
			&item.FullName,
			&item.OPIPNo,
			&item.PhoneNo,
			&item.CreatedAt,
			&item.CreatedBy,
			&item.PatientNo,
			&item.PatientType,
			&item.Status,
		); err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	return result, rows.Err()
}

func (r *PatientRepository) GetHistory(ctx context.Context, patientUUID string) ([]domain.PatientHistoryItem, error) {
	const query = `
		SELECT
			TO_CHAR(lb.created_at, 'YYYY-MM-DD') AS bill_date,
			COALESCE(lb.bill_no, 0) AS bill_no,
			COALESCE(ls.service_name, CONCAT('Service #', COALESCE(lbi.service_id::text, 'NA'))) AS service_name,
			COALESCE(lbi.status, 'PENDING') AS status,
			COALESCE(lbi.unit_price, 0)::float8 AS rate
		FROM patients p
		JOIN lab_bills lb ON lb.patient_id = p.id
		LEFT JOIN lab_bill_items lbi ON lbi.bill_id = lb.id
		LEFT JOIN lab_services ls ON ls.id = lbi.service_id
		WHERE p.patient_uuid::text = $1
		ORDER BY lb.created_at DESC, lbi.id ASC
	`

	rows, err := r.pool.Query(ctx, query, patientUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	history := make([]domain.PatientHistoryItem, 0)
	for rows.Next() {
		var item domain.PatientHistoryItem
		if err := rows.Scan(&item.BillDate, &item.BillNo, &item.ServiceName, &item.Status, &item.Rate); err != nil {
			return nil, err
		}
		history = append(history, item)
	}

	return history, rows.Err()
}

func (r *PatientRepository) UpdateProfile(ctx context.Context, patientUUID string, update domain.PatientProfileUpdate) (domain.PatientProfile, error) {
	const query = `
		UPDATE patients
		SET
			prefix = COALESCE(NULLIF($2, ''), prefix),
			first_name = COALESCE(NULLIF($3, ''), first_name),
			gender = COALESCE(NULLIF($4, ''), gender),
			age = CASE WHEN $5 >= 0 THEN $5 ELSE age END,
			age_unit = COALESCE(NULLIF($6, ''), age_unit),
			phone = COALESCE(NULLIF($7, ''), phone),
			patient_type = COALESCE(NULLIF($8, ''), patient_type),
			status = COALESCE(NULLIF($9, ''), status),
			created_by = COALESCE(NULLIF($10, ''), created_by)
		WHERE patient_uuid::text = $1
		RETURNING
			patient_uuid::text,
			COALESCE(prefix, ''),
			first_name,
			COALESCE(gender, ''),
			COALESCE(age, 0),
			COALESCE(age_unit, ''),
			COALESCE(phone, ''),
			COALESCE(patient_type, ''),
			COALESCE(status, 'Active'),
			COALESCE(created_by, '')
	`

	var profile domain.PatientProfile
	err := r.pool.QueryRow(
		ctx,
		query,
		patientUUID,
		update.Prefix,
		update.FirstName,
		update.Gender,
		update.Age,
		update.AgeUnit,
		update.PhoneNo,
		update.PatientType,
		update.Status,
		update.UpdatedBy,
	).Scan(
		&profile.PatientID,
		&profile.Prefix,
		&profile.FirstName,
		&profile.Gender,
		&profile.Age,
		&profile.AgeUnit,
		&profile.PhoneNo,
		&profile.PatientType,
		&profile.Status,
		&profile.UpdatedBy,
	)
	if err != nil {
		return domain.PatientProfile{}, err
	}

	return profile, nil
}

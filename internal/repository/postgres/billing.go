package postgres

import (
	"context"

	"cprt-lis/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BillingRepository struct {
	pool *pgxpool.Pool
}

func NewBillingRepository(pool *pgxpool.Pool) *BillingRepository {
	return &BillingRepository{pool: pool}
}

func (r *BillingRepository) CreateBill(ctx context.Context, bill domain.LabBill) (domain.LabBill, error) {
	const query = `
		INSERT INTO lab_bills (bill_uuid, patient_id, visit_id, doctor_id, total_amount, discount_amount, tax_amount, net_amount, status, payment_mode)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, bill_uuid, created_at
	`
	row := r.pool.QueryRow(ctx, query, bill.PatientID, bill.VisitID, bill.DoctorID, bill.TotalAmount, bill.Discount, bill.Tax, bill.NetAmount, bill.Status, bill.PaymentMode)
	if err := row.Scan(&bill.ID, &bill.BillUUID, &bill.CreatedAt); err != nil {
		return domain.LabBill{}, err
	}
	return bill, nil
}

func (r *BillingRepository) AddBillItem(ctx context.Context, billID, serviceID int64, qty int, unitPrice float64) error {
	const query = `
		INSERT INTO lab_bill_items (bill_id, service_id, qty, unit_price, discount, tax, line_total, status)
		VALUES ($1, $2, $3, $4, 0, 0, $5, 'PENDING')
	`
	lineTotal := float64(qty) * unitPrice
	_, err := r.pool.Exec(ctx, query, billID, serviceID, qty, unitPrice, lineTotal)
	return err
}

func (r *BillingRepository) GenerateBill(ctx context.Context, bill domain.LabBill, services []domain.BillService) (domain.LabBill, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.LabBill{}, err
	}
	defer tx.Rollback(ctx)

	// Resolve patient UUID to ID
	var patientID int64
	const patientQuery = `SELECT id FROM patients WHERE patient_uuid = $1`
	if err := tx.QueryRow(ctx, patientQuery, bill.PatientUUID).Scan(&patientID); err != nil {
		return domain.LabBill{}, err
	}

	// Calculate totals
	var totalBilled float64
	for _, svc := range services {
		totalBilled += svc.Rate
	}

	taxAmt := (totalBilled - bill.Discount) * bill.Tax / 100.0
	netBilled := totalBilled - bill.Discount + taxAmt
	balanceAmt := netBilled - bill.ReceivedAmount

	paymentStatus := "Paid"
	if balanceAmt > 0 {
		paymentStatus = "Pending"
	} else if balanceAmt < 0 {
		paymentStatus = "Overpaid"
	}

	// Insert bill header
	const billQuery = `
		INSERT INTO lab_bills (
			bill_uuid, patient_id, visit_id, doctor_id, referred_by, hospital_name,
			total_amount, discount_amount, tax_amount, net_amount, 
			received_amount, balance_amount, payment_status, status, payment_mode
		)
		VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, 
			$6, $7, $8, $9, 
			$10, $11, $12, COALESCE(NULLIF($13, ''), 'FINALIZED'), $14
		)
		RETURNING id, bill_uuid, bill_no, created_at
	`

	row := tx.QueryRow(
		ctx,
		billQuery,
		patientID,
		bill.VisitID,
		bill.DoctorID,
		bill.ReferredBy,
		bill.HospitalName,
		totalBilled,
		bill.Discount,
		taxAmt,
		netBilled,
		bill.ReceivedAmount,
		balanceAmt,
		paymentStatus,
		bill.Status,
		bill.PaymentMode,
	)

	var billID int64
	if err := row.Scan(&billID, &bill.BillUUID, &bill.BillNo, &bill.CreatedAt); err != nil {
		return domain.LabBill{}, err
	}

	// Insert bill items
	const itemQuery = `
		INSERT INTO lab_bill_items (bill_id, service_id, qty, unit_price, discount, tax, line_total, status)
		VALUES ($1, NULL, 1, $2, 0, 0, $2, 'PENDING')
	`

	for _, svc := range services {
		if _, err := tx.Exec(ctx, itemQuery, billID, svc.Rate); err != nil {
			return domain.LabBill{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.LabBill{}, err
	}

	bill.ID = billID
	bill.PatientID = patientID
	bill.TotalAmount = totalBilled
	bill.Tax = taxAmt
	bill.NetAmount = netBilled
	bill.BalanceAmount = balanceAmt
	bill.PaymentStatus = paymentStatus

	return bill, nil
}

func (r *BillingRepository) GetServices(ctx context.Context) ([]domain.LabService, error) {
	const query = `
		SELECT id, service_name, COALESCE(base_price, 0)::float8, COALESCE(department, ''), COALESCE(status, 'ACTIVE')
		FROM lab_services
		ORDER BY service_name ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	services := make([]domain.LabService, 0)
	for rows.Next() {
		var item domain.LabService
		if err := rows.Scan(&item.ServiceID, &item.ServiceName, &item.Rate, &item.Department, &item.Status); err != nil {
			return nil, err
		}
		services = append(services, item)
	}

	return services, rows.Err()
}

func (r *BillingRepository) UpdatePayment(ctx context.Context, billID int64, receivedAmt float64, paymentMode string) (domain.BillPaymentUpdate, error) {
	const query = `
		UPDATE lab_bills
		SET
			received_amount = COALESCE(received_amount, 0) + $2,
			balance_amount = net_amount - (COALESCE(received_amount, 0) + $2),
			payment_mode = COALESCE(NULLIF($3, ''), payment_mode),
			payment_status = CASE
				WHEN net_amount - (COALESCE(received_amount, 0) + $2) <= 0 THEN 'Paid'
				ELSE 'Pending'
			END
		WHERE id = $1
		RETURNING
			bill_uuid::text,
			COALESCE(bill_no, 0),
			COALESCE(net_amount, 0)::float8,
			COALESCE(received_amount, 0)::float8,
			COALESCE(balance_amount, 0)::float8,
			COALESCE(payment_mode, ''),
			COALESCE(payment_status, 'Pending'),
			TO_CHAR(NOW(), 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
	`

	var out domain.BillPaymentUpdate
	if err := r.pool.QueryRow(ctx, query, billID, receivedAmt, paymentMode).Scan(
		&out.BillID,
		&out.BillNo,
		&out.NetBilledAmt,
		&out.ReceivedAmt,
		&out.BalanceAmt,
		&out.PaymentMode,
		&out.PaymentStatus,
		&out.UpdatedAt,
	); err != nil {
		return domain.BillPaymentUpdate{}, err
	}

	return out, nil
}

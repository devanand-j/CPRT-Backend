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

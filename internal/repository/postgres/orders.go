package postgres

import (
	"context"

	"cprt-lis/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order domain.LabOrder) (domain.LabOrder, error) {
	const query = `
		INSERT INTO lab_orders (order_uuid, bill_id, patient_id, visit_id, order_status)
		VALUES (gen_random_uuid(), $1, $2, $3, $4)
		RETURNING id, order_uuid, created_at
	`
	row := r.pool.QueryRow(ctx, query, order.BillID, order.PatientID, order.VisitID, order.Status)
	if err := row.Scan(&order.ID, &order.OrderUUID, &order.CreatedAt); err != nil {
		return domain.LabOrder{}, err
	}
	return order, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID int64, status string) error {
	const query = `
		UPDATE lab_orders
		SET order_status = $1
		WHERE id = $2
	`
	_, err := r.pool.Exec(ctx, query, status, orderID)
	return err
}

package repository

import (
	"context"

	"cprt-lis/internal/domain"
)

type UserRepository interface {
	GetByUsername(ctx context.Context, username string) (domain.User, error)
}

type PatientRepository interface {
	Create(ctx context.Context, patient domain.Patient) (domain.Patient, error)
	GetByID(ctx context.Context, id int64) (domain.Patient, error)
	Search(ctx context.Context, mrn, phone string) ([]domain.Patient, error)
}

type BillingRepository interface {
	CreateBill(ctx context.Context, bill domain.LabBill) (domain.LabBill, error)
	AddBillItem(ctx context.Context, billID, serviceID int64, qty int, unitPrice float64) error
}

type OrderRepository interface {
	CreateOrder(ctx context.Context, order domain.LabOrder) (domain.LabOrder, error)
	UpdateStatus(ctx context.Context, orderID int64, status string) error
}

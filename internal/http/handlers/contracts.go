package handlers

import (
	"context"

	"cprt-lis/internal/domain"
)

type AuthService interface {
	Login(ctx context.Context, username, password string) (string, domain.User, error)
}

type PatientService interface {
	Create(ctx context.Context, patient domain.Patient) (domain.Patient, error)
	GetByID(ctx context.Context, id int64) (domain.Patient, error)
	Search(ctx context.Context, mrn, phone string) ([]domain.Patient, error)
}

type BillingService interface {
	CreateBill(ctx context.Context, bill domain.LabBill) (domain.LabBill, error)
	AddBillItem(ctx context.Context, billID, serviceID int64, qty int, unitPrice float64) error
}

type OrderService interface {
	CreateOrder(ctx context.Context, order domain.LabOrder) (domain.LabOrder, error)
	UpdateStatus(ctx context.Context, orderID int64, status string) error
}

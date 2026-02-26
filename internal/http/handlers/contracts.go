package handlers

import (
	"context"

	"cprt-lis/internal/domain"
)

type AuthService interface {
	Login(ctx context.Context, username, password string) (string, domain.User, error)
}

type UserService interface {
	GetAll(ctx context.Context) ([]domain.User, error)
	Update(ctx context.Context, id string, accountGroupID *int64, status, passwordHash, updatedBy *string) error
}

type PatientService interface {
	Register(ctx context.Context, patient domain.Patient) (domain.Patient, error)
	Create(ctx context.Context, patient domain.Patient) (domain.Patient, error)
	Search(ctx context.Context, query string) ([]domain.PatientSearchResult, error)
	History(ctx context.Context, patientUUID string) ([]domain.PatientHistoryItem, error)
	UpdateProfile(ctx context.Context, patientUUID string, update domain.PatientProfileUpdate) (domain.PatientProfile, error)
}

type BillingService interface {
	CreateBill(ctx context.Context, bill domain.LabBill) (domain.LabBill, error)
	AddBillItem(ctx context.Context, billID, serviceID int64, qty int, unitPrice float64) error
	GenerateBill(ctx context.Context, bill domain.LabBill, services []domain.BillService) (domain.LabBill, error)
	GetServices(ctx context.Context) ([]domain.LabService, error)
	UpdatePayment(ctx context.Context, billID int64, receivedAmt float64, paymentMode string) (domain.BillPaymentUpdate, error)
}

type OrderService interface {
	CreateOrder(ctx context.Context, order domain.LabOrder) (domain.LabOrder, error)
	UpdateStatus(ctx context.Context, orderID int64, status string) error
}

type LabService interface {
	MarkSampleCollected(ctx context.Context, billID int64, sampleNo, collectedBy string) (domain.SampleCollectionResponse, error)
	VerifyResults(ctx context.Context, billID int64, params []domain.ResultVerificationParam, verifiedBy string) (domain.ResultVerificationResponse, error)
	CertifyResults(ctx context.Context, billID int64, certifiedBy, remarks string) (domain.ResultCertificationResponse, error)
	GetReport(ctx context.Context, billID int64) (domain.LabReportResponse, error)
}

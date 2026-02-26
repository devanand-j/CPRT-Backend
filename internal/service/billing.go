package service

import (
	"context"

	"cprt-lis/internal/domain"
	"cprt-lis/internal/repository"
)

type BillingService struct {
	repo repository.BillingRepository
}

func NewBillingService(repo repository.BillingRepository) *BillingService {
	return &BillingService{repo: repo}
}

func (s *BillingService) CreateBill(ctx context.Context, bill domain.LabBill) (domain.LabBill, error) {
	return s.repo.CreateBill(ctx, bill)
}

func (s *BillingService) AddBillItem(ctx context.Context, billID, serviceID int64, qty int, unitPrice float64) error {
	return s.repo.AddBillItem(ctx, billID, serviceID, qty, unitPrice)
}

func (s *BillingService) GenerateBill(ctx context.Context, bill domain.LabBill, services []domain.BillService) (domain.LabBill, error) {
	return s.repo.GenerateBill(ctx, bill, services)
}

func (s *BillingService) GetServices(ctx context.Context) ([]domain.LabService, error) {
	return s.repo.GetServices(ctx)
}

func (s *BillingService) UpdatePayment(ctx context.Context, billID int64, receivedAmt float64, paymentMode string) (domain.BillPaymentUpdate, error) {
	return s.repo.UpdatePayment(ctx, billID, receivedAmt, paymentMode)
}

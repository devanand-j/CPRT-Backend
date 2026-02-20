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

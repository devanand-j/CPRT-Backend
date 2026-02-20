package service

import (
	"context"

	"cprt-lis/internal/domain"
	"cprt-lis/internal/repository"
)

type OrderService struct {
	repo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) CreateOrder(ctx context.Context, order domain.LabOrder) (domain.LabOrder, error) {
	return s.repo.CreateOrder(ctx, order)
}

func (s *OrderService) UpdateStatus(ctx context.Context, orderID int64, status string) error {
	return s.repo.UpdateStatus(ctx, orderID, status)
}

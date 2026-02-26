package service

import (
	"context"

	"cprt-lis/internal/domain"
	"cprt-lis/internal/repository"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetAll(ctx context.Context) ([]domain.User, error) {
	return s.repo.GetAll(ctx)
}

func (s *UserService) Update(ctx context.Context, id string, accountGroupID *int64, status, passwordHash, updatedBy *string) error {
	return s.repo.Update(ctx, id, accountGroupID, status, passwordHash, updatedBy)
}

package service

import (
	"context"

	"cprt-lis/internal/domain"
	"cprt-lis/internal/repository"
)

type PatientService struct {
	repo repository.PatientRepository
}

func NewPatientService(repo repository.PatientRepository) *PatientService {
	return &PatientService{repo: repo}
}

func (s *PatientService) Create(ctx context.Context, patient domain.Patient) (domain.Patient, error) {
	return s.repo.Create(ctx, patient)
}

func (s *PatientService) GetByID(ctx context.Context, id int64) (domain.Patient, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *PatientService) Search(ctx context.Context, mrn, phone string) ([]domain.Patient, error) {
	return s.repo.Search(ctx, mrn, phone)
}

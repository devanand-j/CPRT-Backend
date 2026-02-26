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

func (s *PatientService) Register(ctx context.Context, patient domain.Patient) (domain.Patient, error) {
	return s.repo.Create(ctx, patient)
}

func (s *PatientService) Create(ctx context.Context, patient domain.Patient) (domain.Patient, error) {
	return s.Register(ctx, patient)
}

func (s *PatientService) Search(ctx context.Context, query string) ([]domain.PatientSearchResult, error) {
	return s.repo.SearchByQuery(ctx, query)
}

func (s *PatientService) History(ctx context.Context, patientUUID string) ([]domain.PatientHistoryItem, error) {
	return s.repo.GetHistory(ctx, patientUUID)
}

func (s *PatientService) UpdateProfile(ctx context.Context, patientUUID string, update domain.PatientProfileUpdate) (domain.PatientProfile, error) {
	return s.repo.UpdateProfile(ctx, patientUUID, update)
}

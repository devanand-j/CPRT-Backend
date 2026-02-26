package service

import (
	"context"

	"cprt-lis/internal/domain"
	"cprt-lis/internal/repository"
)

type LabService struct {
	repo repository.LabRepository
}

func NewLabService(repo repository.LabRepository) *LabService {
	return &LabService{repo: repo}
}

func (s *LabService) MarkSampleCollected(ctx context.Context, billID int64, sampleNo, collectedBy string) (domain.SampleCollectionResponse, error) {
	return s.repo.MarkSampleCollected(ctx, billID, sampleNo, collectedBy)
}

func (s *LabService) VerifyResults(ctx context.Context, billID int64, params []domain.ResultVerificationParam, verifiedBy string) (domain.ResultVerificationResponse, error) {
	return s.repo.VerifyResults(ctx, billID, params, verifiedBy)
}

func (s *LabService) CertifyResults(ctx context.Context, billID int64, certifiedBy, remarks string) (domain.ResultCertificationResponse, error) {
	return s.repo.CertifyResults(ctx, billID, certifiedBy, remarks)
}

func (s *LabService) GetReport(ctx context.Context, billID int64) (domain.LabReportResponse, error) {
	return s.repo.GetReport(ctx, billID)
}

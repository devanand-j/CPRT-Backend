package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"cprt-lis/internal/domain"

	"github.com/labstack/echo/v4"
)

type LabHandler struct {
	service LabService
}

func NewLabHandler(service LabService) *LabHandler {
	return &LabHandler{service: service}
}

type sampleCollectionRequest struct {
	SampleNo    string `json:"sample_no"`
	CollectedBy string `json:"collected_by"`
	WorksheetNo string `json:"worksheet_no"`
}

type verifyResultsRequest struct {
	BillID     int64                            `json:"bill_id"`
	Params     []domain.ResultVerificationParam `json:"params"`
	VerifiedBy string                           `json:"verified_by"`
}

type certifyResultsRequest struct {
	CertifiedBy string `json:"certified_by"`
	Remarks     string `json:"remarks"`
}

func (h *LabHandler) MarkSampleCollection(c echo.Context) error {
	billID, err := strconv.ParseInt(c.Param("billId"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid bill id")
	}

	var req sampleCollectionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}
	if strings.TrimSpace(req.SampleNo) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "sample_no is required")
	}
	if strings.TrimSpace(req.CollectedBy) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "collected_by is required")
	}

	resp, err := h.service.MarkSampleCollected(c.Request().Context(), billID, req.SampleNo, req.CollectedBy)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if strings.TrimSpace(req.WorksheetNo) != "" {
		resp.WorksheetNo = strings.TrimSpace(req.WorksheetNo)
	}

	return c.JSON(http.StatusOK, []domain.SampleCollectionResponse{resp})
}

func (h *LabHandler) VerifyResults(c echo.Context) error {
	var req verifyResultsRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}
	if req.BillID <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "bill_id is required")
	}
	if len(req.Params) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "params cannot be empty")
	}
	if strings.TrimSpace(req.VerifiedBy) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "verified_by is required")
	}

	resp, err := h.service.VerifyResults(c.Request().Context(), req.BillID, req.Params, req.VerifiedBy)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, []domain.ResultVerificationResponse{resp})
}

func (h *LabHandler) CertifyResults(c echo.Context) error {
	billID, err := strconv.ParseInt(c.Param("billId"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid bill id")
	}

	var req certifyResultsRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}
	if strings.TrimSpace(req.CertifiedBy) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "certified_by is required")
	}

	resp, err := h.service.CertifyResults(c.Request().Context(), billID, req.CertifiedBy, req.Remarks)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, []domain.ResultCertificationResponse{resp})
}

func (h *LabHandler) GetReport(c echo.Context) error {
	billID, err := strconv.ParseInt(c.Param("billId"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid bill id")
	}

	report, err := h.service.GetReport(c.Request().Context(), billID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, []domain.LabReportResponse{report})
}

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

// SampleCollectionRequest is the payload for PATCH /api/lab/sample-collection/:billId.
type SampleCollectionRequest struct {
	// Unique sample bar-code / tube number — required
	SampleNo string `json:"sample_no"   `
	// Login ID or name of the staff who collected the sample — required
	CollectedBy string `json:"collected_by"`
	// Optional worksheet / register number
	WorksheetNo string `json:"worksheet_no"`
}

// VerifyResultsRequest is the payload for POST /api/lab/results/verify.
type VerifyResultsRequest struct {
	// Numeric bill ID
	BillID int64 `json:"bill_id"`
	// One entry per test parameter
	Params []domain.ResultVerificationParam `json:"params"`
	// Login ID or name of the verifying staff — required
	VerifiedBy string `json:"verified_by"`
}

// CertifyResultsRequest is the payload for PATCH /api/lab/results/certify/:billId.
type CertifyResultsRequest struct {
	// Login ID or name of the certifying doctor — required
	CertifiedBy string `json:"certified_by"`
	Remarks     string `json:"remarks"     `
}

// MarkSampleCollection marks a sample as collected for a bill.
//
//	@Summary      Mark sample collected
//	@Description  Records that the sample for the given bill has been collected. Generates a sample ID and worksheet linkage.
//	@Description  step 1 of the lab workflow: Billing → Sample Collection → Verify → Certify → Report.
//	@Tags         Lab
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        billId  path      int                              true  "Numeric bill ID"
//	@Param        body    body      SampleCollectionRequest          true  "Sample collection details"
//	@Success      200     {array}   domain.SampleCollectionResponse  "Single-element array confirming collection"
//	@Failure      400     {object}  ErrorResponse                    "invalid bill id, or sample_no / collected_by missing"
//	@Failure      401     {object}  ErrorResponse                    "Missing or invalid JWT"
//	@Failure      500     {object}  ErrorResponse                    "Unexpected server error"
//	@Router       /api/lab/sample-collection/{billId} [patch]
func (h *LabHandler) MarkSampleCollection(c echo.Context) error {
	billID, err := strconv.ParseInt(c.Param("billId"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid bill id")
	}

	var req SampleCollectionRequest
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

// VerifyResults records test parameter results and marks the bill as verified.
//
//	@Summary      Verify lab results
//	@Description  Accepts an array of test parameter results (value + abnormal flag) and marks the bill as verified.
//	@Description  step 2 of the lab workflow. Each params entry needs param_id and result_value at minimum.
//	@Tags         Lab
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      VerifyResultsRequest               true  "Result parameters and verifier identity"
//	@Success      200   {array}   domain.ResultVerificationResponse  "Single-element array confirming verification"
//	@Failure      400   {object}  ErrorResponse                      "bill_id required, params empty, or verified_by missing"
//	@Failure      401   {object}  ErrorResponse                      "Missing or invalid JWT"
//	@Failure      500   {object}  ErrorResponse                      "Unexpected server error"
//	@Router       /api/lab/results/verify [post]
func (h *LabHandler) VerifyResults(c echo.Context) error {
	var req VerifyResultsRequest
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

// CertifyResults certifies the lab report for a bill.
//
//	@Summary      Certify lab results
//	@Description  Marks the bill as certified by a doctor, making the report ready for dispatch.
//	@Description  step 3 (final) of the lab workflow. Must be called after VerifyResults.
//	@Tags         Lab
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        billId  path      int                               true  "Numeric bill ID"
//	@Param        body    body      CertifyResultsRequest             true  "Certifier identity and optional remarks"
//	@Success      200     {array}   domain.ResultCertificationResponse "Single-element array confirming certification and dispatch readiness"
//	@Failure      400     {object}  ErrorResponse                      "invalid bill id or certified_by missing"
//	@Failure      401     {object}  ErrorResponse                      "Missing or invalid JWT"
//	@Failure      500     {object}  ErrorResponse                      "Unexpected server error"
//	@Router       /api/lab/results/certify/{billId} [patch]
func (h *LabHandler) CertifyResults(c echo.Context) error {
	billID, err := strconv.ParseInt(c.Param("billId"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid bill id")
	}

	var req CertifyResultsRequest
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

// GetReport retrieves the final lab report for a bill.
//
//	@Summary      Get lab report
//	@Description  Returns the complete lab report for a certified bill, including patient info, results, flags and certifier.
//	@Description  Only available after the bill has been certified via PATCH /api/lab/results/certify/:billId.
//	@Tags         Lab
//	@Produce      json
//	@Security     BearerAuth
//	@Param        billId  path      int                         true  "Numeric bill ID"
//	@Success      200     {array}   domain.LabReportResponse    "Single-element array with the full report"
//	@Failure      400     {object}  ErrorResponse               "invalid bill id"
//	@Failure      401     {object}  ErrorResponse               "Missing or invalid JWT"
//	@Failure      500     {object}  ErrorResponse               "Unexpected server error"
//	@Router       /api/lab/reports/{billId} [get]
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


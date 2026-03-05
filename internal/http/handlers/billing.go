package handlers

import (
	"net/http"
	"strconv"

	"cprt-lis/internal/domain"

	"github.com/labstack/echo/v4"
)

type BillingHandler struct {
	service BillingService
}

func NewBillingHandler(service BillingService) *BillingHandler {
	return &BillingHandler{service: service}
}

// CreateBillRequest is the raw payload for POST /api/billing/bills.
// Use POST /api/billing/new (GenerateBill) for the high-level workflow instead.
type CreateBillRequest struct {
	// Internal numeric patient ID (from patients table)
	PatientID   int64   `json:"patient_id"  `
	VisitID     *int64  `json:"visit_id"    `
	DoctorID    *int64  `json:"doctor_id"   `
	TotalAmount float64 `json:"total_amount"`
	Discount    float64 `json:"discount"    `
	Tax         float64 `json:"tax"         `
	NetAmount   float64 `json:"net_amount"  `
	// Pending | Paid | Partial
	Status string `json:"status"      `
	// Cash | Card | UPI | Online
	PaymentMode string `json:"payment_mode"`
}

// AddBillItemRequest is the payload for POST /api/billing/bills/:id/items.
type AddBillItemRequest struct {
	ServiceID int64   `json:"service_id"`
	Qty       int     `json:"qty"       `
	UnitPrice float64 `json:"unit_price"`
}

// UpdatePaymentRequest is the payload for PATCH /api/billing/:billId/payment.
type UpdatePaymentRequest struct {
	ReceivedAmt float64 `json:"received_amt"`
	// Cash | Card | UPI | Online
	PaymentMode string `json:"payment_mode"`
}

// CreateBill creates a raw bill record (low-level).
//
//	@Summary      Create bill (low-level)
//	@Description  Creates a bill record directly. Prefer POST /api/billing/new for the full workflow with services, discount and tax.
//	@Tags         Billing
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      CreateBillRequest  true  "Raw bill data"
//	@Success      201   {object}  domain.LabBill     "Created bill record (full domain object)"
//	@Failure      400   {object}  ErrorResponse      "invalid payload"
//	@Failure      401   {object}  ErrorResponse      "Missing or invalid JWT"
//	@Failure      500   {object}  ErrorResponse      "Unexpected server error"
//	@Router       /api/billing/bills [post]
func (h *BillingHandler) CreateBill(c echo.Context) error {
	var req CreateBillRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}

	bill, err := h.service.CreateBill(c.Request().Context(), domain.LabBill{
		PatientID:   req.PatientID,
		VisitID:     req.VisitID,
		DoctorID:    req.DoctorID,
		TotalAmount: req.TotalAmount,
		Discount:    req.Discount,
		Tax:         req.Tax,
		NetAmount:   req.NetAmount,
		Status:      req.Status,
		PaymentMode: req.PaymentMode,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, bill)
}

// AddBillItem appends a service line-item to an existing bill.
//
//	@Summary      Add item to bill
//	@Description  Appends a service/test line item to the bill specified by :id.
//	@Tags         Billing
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id    path      int               true  "Numeric bill ID"
//	@Param        body  body      AddBillItemRequest true  "Service item details"
//	@Success      200   {object}  map[string]string  "{\"status\": \"added\"}"
//	@Failure      400   {object}  ErrorResponse      "invalid bill id or payload"
//	@Failure      401   {object}  ErrorResponse      "Missing or invalid JWT"
//	@Failure      500   {object}  ErrorResponse      "Unexpected server error"
//	@Router       /api/billing/bills/{id}/items [post]
func (h *BillingHandler) AddBillItem(c echo.Context) error {
	billID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid bill id")
	}

	var req AddBillItemRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}

	if err := h.service.AddBillItem(c.Request().Context(), billID, req.ServiceID, req.Qty, req.UnitPrice); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{"status": "added"})
}

// GenerateBillServiceItem describes a single service/test within GenerateBillRequest.
type GenerateBillServiceItem struct {
	// Service UUID or code (from GET /api/billing/services)
	ServiceID   string  `json:"service_id"  `
	ServiceName string  `json:"service_name"`
	Rate        float64 `json:"rate"        `
}

// GenerateBillRequest is the payload for POST /api/billing/new — the primary billing endpoint.
type GenerateBillRequest struct {
	// Patient UUID returned by POST /api/patients/register
	PatientID    string `json:"patient_id"   `
	ReferredBy   string `json:"referred_by"  `
	HospitalName string `json:"hospital_name"`
	// At least one service is required
	Services []GenerateBillServiceItem `json:"services"`
	// Flat discount amount in currency units
	DiscountAmt float64 `json:"discount_amt"`
	// Tax percentage (e.g. 5 = 5%)
	TaxPercent float64 `json:"tax_percent" `
	// Amount collected upfront
	ReceivedAmt float64 `json:"received_amt"`
}

// GenerateBillResponse is each element of the array returned after a bill is generated.
type GenerateBillResponse struct {
	// Bill UUID
	BillID         string  `json:"bill_id"         `
	BillNo         int64   `json:"bill_no"         `
	TotalBilledAmt float64 `json:"total_billed_amt"`
	DiscountAmt    float64 `json:"discount_amt"    `
	TaxAmt         float64 `json:"tax_amt"         `
	NetBilledAmt   float64 `json:"net_billed_amt"  `
	ReceivedAmt    float64 `json:"received_amt"    `
	BalanceAmt     float64 `json:"balance_amt"     `
	// Paid | Pending | Partial
	PaymentStatus string `json:"payment_status"`
	CreatedAt     string `json:"created_at"    `
}

// GenerateBill creates a complete bill with services, discount, tax and received amount in one call.
//
//	@Summary      Generate bill (primary workflow)
//	@Description  Creates a bill with one or more services, applies discount and tax, and records the received amount.
//	@Description  patient_id must be the UUID returned by POST /api/patients/register.
//	@Description  service_id values come from GET /api/billing/services.
//	@Description  The response includes bill_no, net amount, balance, and payment_status (Paid/Pending/Partial).
//	@Tags         Billing
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      GenerateBillRequest   true  "Bill generation payload"
//	@Success      201   {array}   GenerateBillResponse  "Single-element array with full bill summary"
//	@Failure      400   {object}  ErrorResponse         "patient_id is required or services cannot be empty"
//	@Failure      401   {object}  ErrorResponse         "Missing or invalid JWT"
//	@Failure      500   {object}  ErrorResponse         "Unexpected server error"
//	@Router       /api/billing/new [post]
func (h *BillingHandler) GenerateBill(c echo.Context) error {
	var req GenerateBillRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}

	if req.PatientID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "patient_id is required")
	}
	if len(req.Services) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "services cannot be empty")
	}

	services := make([]domain.BillService, len(req.Services))
	for i, svc := range req.Services {
		services[i] = domain.BillService{
			ServiceID:   svc.ServiceID,
			ServiceName: svc.ServiceName,
			Rate:        svc.Rate,
		}
	}

	bill, err := h.service.GenerateBill(c.Request().Context(), domain.LabBill{
		PatientUUID:    req.PatientID,
		ReferredBy:     req.ReferredBy,
		HospitalName:   req.HospitalName,
		Discount:       req.DiscountAmt,
		Tax:            req.TaxPercent,
		ReceivedAmount: req.ReceivedAmt,
	}, services)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	resp := []GenerateBillResponse{{
		BillID:         bill.BillUUID,
		BillNo:         bill.BillNo,
		TotalBilledAmt: bill.TotalAmount,
		DiscountAmt:    bill.Discount,
		TaxAmt:         bill.Tax,
		NetBilledAmt:   bill.NetAmount,
		ReceivedAmt:    bill.ReceivedAmount,
		BalanceAmt:     bill.BalanceAmount,
		PaymentStatus:  bill.PaymentStatus,
		CreatedAt:      bill.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}}

	return c.JSON(http.StatusCreated, resp)
}

// GetServices returns all available lab/diagnostic services.
//
//	@Summary      List lab services
//	@Description  Returns all active services with their ID, name, rate and department.
//	@Description  Use the service_id values when generating a bill via POST /api/billing/new.
//	@Tags         Billing
//	@Produce      json
//	@Security     BearerAuth
//	@Success      200  {array}   domain.LabService  "List of available services"
//	@Failure      401  {object}  ErrorResponse      "Missing or invalid JWT"
//	@Failure      500  {object}  ErrorResponse      "Unexpected server error"
//	@Router       /api/billing/services [get]
func (h *BillingHandler) GetServices(c echo.Context) error {
	services, err := h.service.GetServices(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, services)
}

// UpdatePayment records an additional payment against a bill.
//
//	@Summary      Update bill payment
//	@Description  Records a payment received against an existing bill and recalculates balance and payment_status.
//	@Description  payment_status will be: Paid (balance=0), Partial (balance>0), or Pending (received=0).
//	@Tags         Billing
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        billId  path      int                     true  "Numeric bill ID"
//	@Param        body    body      UpdatePaymentRequest    true  "Payment details"
//	@Success      200     {array}   domain.BillPaymentUpdate "Single-element array with updated payment summary"
//	@Failure      400     {object}  ErrorResponse            "invalid bill id or payload"
//	@Failure      401     {object}  ErrorResponse            "Missing or invalid JWT"
//	@Failure      500     {object}  ErrorResponse            "Unexpected server error"
//	@Router       /api/billing/{billId}/payment [patch]
func (h *BillingHandler) UpdatePayment(c echo.Context) error {
	billID, err := strconv.ParseInt(c.Param("billId"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid bill id")
	}

	var req UpdatePaymentRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}

	resp, err := h.service.UpdatePayment(c.Request().Context(), billID, req.ReceivedAmt, req.PaymentMode)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, []domain.BillPaymentUpdate{resp})
}


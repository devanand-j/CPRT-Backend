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

type createBillRequest struct {
	PatientID   int64   `json:"patient_id"`
	VisitID     *int64  `json:"visit_id"`
	DoctorID    *int64  `json:"doctor_id"`
	TotalAmount float64 `json:"total_amount"`
	Discount    float64 `json:"discount"`
	Tax         float64 `json:"tax"`
	NetAmount   float64 `json:"net_amount"`
	Status      string  `json:"status"`
	PaymentMode string  `json:"payment_mode"`
}

type addBillItemRequest struct {
	ServiceID int64   `json:"service_id"`
	Qty       int     `json:"qty"`
	UnitPrice float64 `json:"unit_price"`
}

type updatePaymentRequest struct {
	ReceivedAmt float64 `json:"received_amt"`
	PaymentMode string  `json:"payment_mode"`
}

func (h *BillingHandler) CreateBill(c echo.Context) error {
	var req createBillRequest
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

func (h *BillingHandler) AddBillItem(c echo.Context) error {
	billID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid bill id")
	}

	var req addBillItemRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}

	if err := h.service.AddBillItem(c.Request().Context(), billID, req.ServiceID, req.Qty, req.UnitPrice); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{"status": "added"})
}

type generateBillRequest struct {
	PatientID    string `json:"patient_id"`
	ReferredBy   string `json:"referred_by"`
	HospitalName string `json:"hospital_name"`
	Services     []struct {
		ServiceID   string  `json:"service_id"`
		ServiceName string  `json:"service_name"`
		Rate        float64 `json:"rate"`
	} `json:"services"`
	DiscountAmt float64 `json:"discount_amt"`
	TaxPercent  float64 `json:"tax_percent"`
	ReceivedAmt float64 `json:"received_amt"`
}

type generateBillResponse struct {
	BillID         string  `json:"bill_id"`
	BillNo         int64   `json:"bill_no"`
	TotalBilledAmt float64 `json:"total_billed_amt"`
	DiscountAmt    float64 `json:"discount_amt"`
	TaxAmt         float64 `json:"tax_amt"`
	NetBilledAmt   float64 `json:"net_billed_amt"`
	ReceivedAmt    float64 `json:"received_amt"`
	BalanceAmt     float64 `json:"balance_amt"`
	PaymentStatus  string  `json:"payment_status"`
	CreatedAt      string  `json:"created_at"`
}

func (h *BillingHandler) GenerateBill(c echo.Context) error {
	var req generateBillRequest
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

	resp := []generateBillResponse{{
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

func (h *BillingHandler) GetServices(c echo.Context) error {
	services, err := h.service.GetServices(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, services)
}

func (h *BillingHandler) UpdatePayment(c echo.Context) error {
	billID, err := strconv.ParseInt(c.Param("billId"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid bill id")
	}

	var req updatePaymentRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}

	resp, err := h.service.UpdatePayment(c.Request().Context(), billID, req.ReceivedAmt, req.PaymentMode)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, []domain.BillPaymentUpdate{resp})
}

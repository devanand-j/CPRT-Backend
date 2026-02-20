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

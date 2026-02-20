package handlers

import (
	"net/http"
	"strconv"

	"cprt-lis/internal/domain"

	"github.com/labstack/echo/v4"
)

type OrderHandler struct {
	service OrderService
}

func NewOrderHandler(service OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

type createOrderRequest struct {
	BillID    int64  `json:"bill_id"`
	PatientID int64  `json:"patient_id"`
	VisitID   *int64 `json:"visit_id"`
	Status    string `json:"status"`
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

func (h *OrderHandler) CreateOrder(c echo.Context) error {
	var req createOrderRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}

	order, err := h.service.CreateOrder(c.Request().Context(), domain.LabOrder{
		BillID:    req.BillID,
		PatientID: req.PatientID,
		VisitID:   req.VisitID,
		Status:    req.Status,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) UpdateStatus(c echo.Context) error {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid order id")
	}

	var req updateStatusRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}

	if err := h.service.UpdateStatus(c.Request().Context(), orderID, req.Status); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{"status": "updated"})
}

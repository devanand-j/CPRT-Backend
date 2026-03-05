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

// CreateOrderRequest is the payload for POST /api/orders.
type CreateOrderRequest struct {
	// Numeric bill ID returned by POST /api/billing/new (bill_no field)
	BillID int64 `json:"bill_id"   `
	// Internal numeric patient ID
	PatientID int64  `json:"patient_id"`
	VisitID   *int64 `json:"visit_id"  `
	// Pending | InProgress | Completed
	Status string `json:"status"    `
}

// UpdateStatusRequest is the payload for PATCH /api/orders/:id/status.
type UpdateStatusRequest struct {
	// Pending | InProgress | Completed
	Status string `json:"status"`
}

// CreateOrder creates a lab order linked to a bill.
//
//	@Summary      Create lab order
//	@Description  Creates a lab workflow order linked to an existing bill and patient.
//	@Description  Typically called right after a bill is generated to trigger sample collection.
//	@Tags         Orders
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      CreateOrderRequest  true  "Order details"
//	@Success      201   {object}  domain.LabOrder     "Created order record"
//	@Failure      400   {object}  ErrorResponse       "invalid payload"
//	@Failure      401   {object}  ErrorResponse       "Missing or invalid JWT"
//	@Failure      500   {object}  ErrorResponse       "Unexpected server error"
//	@Router       /api/orders [post]
func (h *OrderHandler) CreateOrder(c echo.Context) error {
	var req CreateOrderRequest
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

// UpdateStatus updates the status of a lab order.
//
//	@Summary      Update order status
//	@Description  Updates the workflow status of the specified order.
//	@Description  Valid statuses: Pending → InProgress → Completed.
//	@Tags         Orders
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id    path      int                 true  "Numeric order ID"
//	@Param        body  body      UpdateStatusRequest true  "New status value"
//	@Success      200   {object}  map[string]string   "{\"status\": \"updated\"}"
//	@Failure      400   {object}  ErrorResponse       "invalid order id or payload"
//	@Failure      401   {object}  ErrorResponse       "Missing or invalid JWT"
//	@Failure      500   {object}  ErrorResponse       "Unexpected server error"
//	@Router       /api/orders/{id}/status [patch]
func (h *OrderHandler) UpdateStatus(c echo.Context) error {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid order id")
	}

	var req UpdateStatusRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}

	if err := h.service.UpdateStatus(c.Request().Context(), orderID, req.Status); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{"status": "updated"})
}


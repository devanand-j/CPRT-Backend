package http

import (
	"net/http"

	"cprt-lis/internal/config"
	"cprt-lis/internal/http/handlers"
	"cprt-lis/internal/http/middleware"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

type Handlers struct {
	Auth     *handlers.AuthHandler
	Patients *handlers.PatientHandler
	Billing  *handlers.BillingHandler
	Orders   *handlers.OrderHandler
	Health   *handlers.HealthHandler
	Config   config.Config
}

func NewHandlers(authService handlers.AuthService, patientService handlers.PatientService, billingService handlers.BillingService, orderService handlers.OrderService, cfg config.Config) Handlers {
	return Handlers{
		Auth:     handlers.NewAuthHandler(authService),
		Patients: handlers.NewPatientHandler(patientService),
		Billing:  handlers.NewBillingHandler(billingService),
		Orders:   handlers.NewOrderHandler(orderService),
		Health:   handlers.NewHealthHandler(),
		Config:   cfg,
	}
}

func NewRouter(h Handlers) *echo.Echo {
	e := echo.New()
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.CORS())

	e.GET("/health", h.Health.Ping)

	api := e.Group("/api/v1")
	api.POST("/auth/login", h.Auth.Login)

	secured := api.Group("")
	secured.Use(middleware.JWTAuth(h.Config.JWTSecret))

	secured.POST("/patients", h.Patients.Create)
	secured.GET("/patients/:id", h.Patients.GetByID)
	secured.GET("/patients", h.Patients.Search)

	secured.POST("/billing/bills", h.Billing.CreateBill)
	secured.POST("/billing/bills/:id/items", h.Billing.AddBillItem)

	secured.POST("/orders", h.Orders.CreateOrder)
	secured.PATCH("/orders/:id/status", h.Orders.UpdateStatus)

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
		}
		_ = c.JSON(code, map[string]any{"error": err.Error()})
	}

	return e
}

package http

import (
	"net/http"

	"cprt-lis/internal/config"
	"cprt-lis/internal/http/handlers"
	"cprt-lis/internal/http/middleware"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type Handlers struct {
	Auth     *handlers.AuthHandler
	Users    *handlers.UserHandler
	Patients *handlers.PatientHandler
	Billing  *handlers.BillingHandler
	Orders   *handlers.OrderHandler
	Lab      *handlers.LabHandler
	Health   *handlers.HealthHandler
	Config   config.Config
}

func NewHandlers(authService handlers.AuthService, userService handlers.UserService, patientService handlers.PatientService, billingService handlers.BillingService, orderService handlers.OrderService, labService handlers.LabService, cfg config.Config) Handlers {
	return Handlers{
		Auth:     handlers.NewAuthHandler(authService),
		Users:    handlers.NewUserHandler(userService),
		Patients: handlers.NewPatientHandler(patientService),
		Billing:  handlers.NewBillingHandler(billingService),
		Orders:   handlers.NewOrderHandler(orderService),
		Lab:      handlers.NewLabHandler(labService),
		Health:   handlers.NewHealthHandler(),
		Config:   cfg,
	}
}

func NewRouter(h Handlers) *echo.Echo {
	e := echo.New()
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.CORS())

	e.File("/swagger.yaml", "swagger.yaml")
	e.GET("/swagger", func(c echo.Context) error {
		return c.HTML(http.StatusOK, `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>CPRT LIS API Docs</title>
	<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
	<div id="swagger-ui"></div>
	<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
	<script>
		window.onload = function () {
			window.ui = SwaggerUIBundle({
				url: '/swagger.yaml',
				dom_id: '#swagger-ui',
				deepLinking: true,
				displayRequestDuration: true,
				persistAuthorization: true,
			});
		};
	</script>
</body>
</html>`)
	})

	e.GET("/health", h.Health.Ping)
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.GET("/redoc", func(c echo.Context) error {
		html := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
  <title>CPRT LIS API – ReDoc</title>
  <style>body { margin: 0; padding: 0; }</style>
</head>
<body>
  <redoc spec-url="/swagger/doc.json"
         expand-responses="200,201"
         required-props-first="true"
         hide-hostname="false"
         path-in-middle-panel="true">
  </redoc>
  <script src="https://cdn.jsdelivr.net/npm/redoc@latest/bundles/redoc.standalone.js"></script>
</body>
</html>`
		return c.HTML(http.StatusOK, html)
	})

	api := e.Group("/api")

	registerAPIRoutes(api, h)

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
		}
		_ = c.JSON(code, map[string]any{"error": err.Error()})
	}

	return e
}

func registerAPIRoutes(api *echo.Group, h Handlers) {
	api.POST("/auth/login", h.Auth.Login)

	secured := api.Group("")
	secured.Use(middleware.JWTAuth(h.Config.JWTSecret))
	secured.Use(middleware.RolePolicyGuard())

	secured.GET("/users", h.Users.GetAll, middleware.RequireRole("SUPER_ADMIN", "ADMIN"))
	secured.PATCH("/users/:userId", h.Users.Update, middleware.RequireRole("SUPER_ADMIN", "ADMIN"))

	secured.POST("/patients/register", h.Patients.Register)
	secured.POST("/patients", h.Patients.Create)
	secured.GET("/patients/:id", h.Patients.GetByID)
	secured.GET("/patients", h.Patients.Search)
	secured.GET("/patients/search", h.Patients.Search)
	secured.GET("/patients/:patientId/history", h.Patients.GetHistory)
	secured.PATCH("/patients/:patientId", h.Patients.Update)

	secured.POST("/billing/new", h.Billing.GenerateBill)
	secured.POST("/billing/bills", h.Billing.CreateBill)
	secured.POST("/billing/bills/:id/items", h.Billing.AddBillItem)
	secured.GET("/billing/services", h.Billing.GetServices)
	secured.PATCH("/billing/:billId/payment", h.Billing.UpdatePayment)

	secured.POST("/orders", h.Orders.CreateOrder)
	secured.PATCH("/orders/:id/status", h.Orders.UpdateStatus)

	secured.PATCH("/lab/sample-collection/:billId", h.Lab.MarkSampleCollection)
	secured.POST("/lab/results/verify", h.Lab.VerifyResults)
	secured.PATCH("/lab/results/certify/:billId", h.Lab.CertifyResults)
	secured.GET("/lab/reports/:billId", h.Lab.GetReport)
}

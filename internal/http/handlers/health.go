package handlers

import "github.com/labstack/echo/v4"

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Ping checks that the API server is running.
//
//	@Summary      Health check
//	@Description  Returns {"status":"ok"} when the server is up and reachable.
//	@Tags         Health
//	@Produce      json
//	@Success      200 {object} map[string]string "Service is healthy"
//	@Router       /health [get]
func (h *HealthHandler) Ping(c echo.Context) error {
	return c.JSON(200, map[string]any{"status": "ok"})
}


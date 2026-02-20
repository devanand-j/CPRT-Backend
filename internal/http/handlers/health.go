package handlers

import "github.com/labstack/echo/v4"

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Ping(c echo.Context) error {
	return c.JSON(200, map[string]any{"status": "ok"})
}

package handlers

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	service AuthService
}

func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

type loginRequest struct {
	LoginID  string `json:"login_id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}

	identifier := req.LoginID
	if identifier == "" {
		identifier = req.Username
	}

	token, user, err := h.service.Login(c.Request().Context(), identifier, req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}

	status := "Active"
	if strings.EqualFold(user.Status, "inactive") {
		status = "Inactive"
	}

	return c.JSON(http.StatusOK, []map[string]any{
		{
			"status": "success",
			"token":  token,
			"user": map[string]any{
				"user_id":            toUserID(user.ID),
				"user_name":          user.DisplayName,
				"account_group_code": user.GroupCode,
				"account_group_name": user.GroupName,
				"role":               strings.ToUpper(strings.TrimSpace(user.GroupCode)),
				"status":             status,
			},
			"permissions": permissionsForRole(user.GroupCode),
		},
	})
}

func permissionsForRole(role string) []string {
	permissionsByRole := map[string][]string{
		"SUPER_ADMIN": {"PATIENT_READ", "PATIENT_WRITE", "BILLING_READ", "BILLING_WRITE", "LAB_VERIFY", "LAB_CERTIFY", "USER_MANAGE"},
		"ADMIN":       {"PATIENT_READ", "BILLING_WRITE", "LAB_VERIFY", "LAB_CERTIFY", "USER_MANAGE"},
		"DOCTOR":      {"PATIENT_READ", "LAB_VERIFY", "LAB_CERTIFY"},
		"TECHNICIAN":  {"PATIENT_READ", "LAB_VERIFY"},
	}

	key := strings.ToUpper(strings.TrimSpace(role))
	if permissions, ok := permissionsByRole[key]; ok {
		return permissions
	}

	return []string{"PATIENT_READ"}
}

func toUserID(id string) string {
	return "USR-" + strings.TrimSpace(id)
}

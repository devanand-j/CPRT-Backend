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

// LoginRequest is the payload for POST /api/auth/login.
type LoginRequest struct {
	// Use either login_id or username — both are accepted
	LoginID  string `json:"login_id" `
	Username string `json:"username" `
	Password string `json:"password" `
}

// LoginUserInfo is the authenticated user block returned inside LoginResponse.
type LoginUserInfo struct {
	UserID           string `json:"user_id"           `
	UserName         string `json:"user_name"         `
	AccountGroupCode string `json:"account_group_code"`
	AccountGroupName string `json:"account_group_name"`
	// Role is one of: SUPER_ADMIN, ADMIN, DOCTOR, TECHNICIAN
	Role   string `json:"role"  `
	Status string `json:"status"`
}

// LoginResponse is each element of the array returned on a successful login.
type LoginResponse struct {
	Status string        `json:"status"`
	Token  string        `json:"token" `
	User   LoginUserInfo `json:"user"`
	// Permissions granted to this role
	// Possible values: PATIENT_READ, PATIENT_WRITE, BILLING_READ, BILLING_WRITE,
	//                  LAB_VERIFY, LAB_CERTIFY, USER_MANAGE
	Permissions []string `json:"permissions"`
}

// Login authenticates a user and returns a JWT bearer token.
//
//	@Summary      User Login
//	@Description  Authenticate with login_id (or username) + password.
//	@Description  Copy the returned token and pass it as `Authorization: Bearer <token>` in every secured request.
//	@Tags         Auth
//	@Accept       json
//	@Produce      json
//	@Param        body  body     LoginRequest   true  "Login credentials — use login_id or username interchangeably"
//	@Success      200   {array}  LoginResponse  "Single-element array — token, user info and role permissions"
//	@Failure      400   {object} ErrorResponse  "Bad request — missing or malformed JSON body"
//	@Failure      401   {object} ErrorResponse  "Unauthorized — wrong login_id/username or password"
//	@Router       /api/auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
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


package handlers

import (
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	service UserService
}

func NewUserHandler(service UserService) *UserHandler {
	return &UserHandler{service: service}
}

// UserResponse is a single user record returned by GET /api/users.
type UserResponse struct {
	UserID   string `json:"user_id"      `
	LoginID  string `json:"login_id"     `
	UserName string `json:"user_name"    `
	// Account group name e.g. Doctor, Admin, Technician
	AccountGroup string `json:"account_group"`
	// Active | Inactive
	Status    string     `json:"status"       `
	CreatedAt time.Time  `json:"created_at"`
	LastLogin *time.Time `json:"last_login"`
}

// GetAll returns all registered system users.
//
//	@Summary      List all users
//	@Description  Returns every user account. Requires SUPER_ADMIN or ADMIN role.
//	@Tags         Users
//	@Produce      json
//	@Security     BearerAuth
//	@Success      200  {array}   UserResponse   "All user accounts"
//	@Failure      401  {object}  ErrorResponse  "Missing or invalid JWT"
//	@Failure      403  {object}  ErrorResponse  "Insufficient role — need SUPER_ADMIN or ADMIN"
//	@Failure      500  {object}  ErrorResponse  "Unexpected server error"
//	@Router       /api/users [get]
func (h *UserHandler) GetAll(c echo.Context) error {
	users, err := h.service.GetAll(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	resp := make([]UserResponse, 0, len(users))
	for _, user := range users {
		status := "Active"
		if strings.EqualFold(user.Status, "inactive") {
			status = "Inactive"
		}

		resp = append(resp, UserResponse{
			UserID:       toUserID(user.ID),
			LoginID:      user.LoginID,
			UserName:     user.DisplayName,
			AccountGroup: user.GroupName,
			Status:       status,
			CreatedAt:    user.CreatedAt,
			LastLogin:    user.LastLogin,
		})
	}

	return c.JSON(http.StatusOK, resp)
}

// UpdateUserRequest is the payload for PATCH /api/users/:userId.
// All fields are optional — include only the ones you want to change.
type UpdateUserRequest struct {
	// Active | Inactive
	Status *string `json:"status"          `
	// Account group ID (from account_groups table)
	AccountGroupID *int64  `json:"account_group_id"`
	UpdatedBy      *string `json:"updated_by"      `
	// Plain-text password — will be bcrypt-hashed server-side
	Password *string `json:"password"     `
	// Pre-hashed password — use only if you manage hashing yourself
	PasswordHash *string `json:"password_hash"`
}

// Update partially updates a user account.
//
//	@Summary      Update user
//	@Description  Partially updates a user account. Requires SUPER_ADMIN or ADMIN role.
//	@Description  Send only the fields you want to change. All fields are optional.
//	@Description  If both password and password_hash are supplied, password takes precedence (server will hash it).
//	@Tags         Users
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        userId  path      string             true  "User ID (raw DB value, not the USR- prefixed one)"
//	@Param        body    body      UpdateUserRequest  true  "Fields to update"
//	@Success      200     {array}   map[string]string  "Single-element array with update confirmation"
//	@Failure      400     {object}  ErrorResponse      "invalid id, invalid payload, or invalid status value"
//	@Failure      401     {object}  ErrorResponse      "Missing or invalid JWT"
//	@Failure      403     {object}  ErrorResponse      "Insufficient role"
//	@Failure      500     {object}  ErrorResponse      "Unexpected server error"
//	@Router       /api/users/{userId} [patch]
func (h *UserHandler) Update(c echo.Context) error {
	id := c.Param("userId")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}

	passwordHash := req.PasswordHash
	if req.Password != nil {
		hashed, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to hash password")
		}
		hash := string(hashed)
		passwordHash = &hash
	}

	if req.Status != nil {
		normalized := strings.ToUpper(strings.TrimSpace(*req.Status))
		if normalized != "ACTIVE" && normalized != "INACTIVE" {
			return echo.NewHTTPError(http.StatusBadRequest, "status must be Active or Inactive")
		}
		value := "Active"
		if normalized == "INACTIVE" {
			value = "Inactive"
		}
		req.Status = &value
	}

	if err := h.service.Update(c.Request().Context(), id, req.AccountGroupID, req.Status, passwordHash, req.UpdatedBy); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	newStatus := "Active"
	if req.Status != nil {
		newStatus = *req.Status
	}

	return c.JSON(http.StatusOK, []map[string]any{
		{
			"status":     "success",
			"message":    "User account updated",
			"user_id":    toUserID(id),
			"new_status": newStatus,
		},
	})
}


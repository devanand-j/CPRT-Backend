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

func (h *UserHandler) GetAll(c echo.Context) error {
	users, err := h.service.GetAll(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	type userResponse struct {
		UserID       string     `json:"user_id"`
		LoginID      string     `json:"login_id"`
		UserName     string     `json:"user_name"`
		AccountGroup string     `json:"account_group"`
		Status       string     `json:"status"`
		CreatedAt    time.Time  `json:"created_at"`
		LastLogin    *time.Time `json:"last_login"`
	}

	resp := make([]userResponse, 0, len(users))
	for _, user := range users {
		status := "Active"
		if strings.EqualFold(user.Status, "inactive") {
			status = "Inactive"
		}

		resp = append(resp, userResponse{
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

type updateUserRequest struct {
	Status         *string `json:"status"`
	AccountGroupID *int64  `json:"account_group_id"`
	UpdatedBy      *string `json:"updated_by"`
	Password       *string `json:"password"`
	PasswordHash   *string `json:"password_hash"`
}

func (h *UserHandler) Update(c echo.Context) error {
	id := c.Param("userId")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	var req updateUserRequest
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

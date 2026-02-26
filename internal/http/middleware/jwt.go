package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func JWTAuth(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if auth == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing token")
			}
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			token, err := jwt.Parse(parts[1], func(t *jwt.Token) (any, error) {
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				c.Set("user_uuid", claims["sub"])
				c.Set("user_role", claims["role"])
				c.Set("user_group_id", claims["group_id"])
				c.Set("user", map[string]any{
					"id":       claims["sub"],
					"role":     claims["role"],
					"group_id": claims["group_id"],
				})
			}

			return next(c)
		}
	}
}

func RolePolicyGuard() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := strings.ToLower(c.Path())
			method := strings.ToUpper(c.Request().Method)

			roleVal := c.Get("user_role")
			role, ok := roleVal.(string)
			if !ok || role == "" {
				return echo.NewHTTPError(http.StatusForbidden, "access denied")
			}

			if isUsersMaintenancePath(path) {
				if !hasRole(role, "SUPER_ADMIN", "ADMIN") {
					return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
				}
				if method == http.MethodPatch && c.Param("userId") == "" {
					return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
				}
			}

			if strings.Contains(path, "/certify") {
				if !hasRole(role, "DOCTOR") {
					return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
				}
			}

			return next(c)
		}
	}
}

func hasRole(role string, allowed ...string) bool {
	for _, a := range allowed {
		if strings.EqualFold(role, a) {
			return true
		}
	}
	return false
}

func isUsersMaintenancePath(path string) bool {
	return strings.HasSuffix(path, "/users") || strings.Contains(path, "/users/")
}

func RequireRole(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userRole := c.Get("user_role")
			if userRole == nil {
				return echo.NewHTTPError(http.StatusForbidden, "access denied")
			}

			roleStr, ok := userRole.(string)
			if !ok {
				return echo.NewHTTPError(http.StatusForbidden, "access denied")
			}

			for _, r := range roles {
				if strings.EqualFold(roleStr, r) {
					return next(c)
				}
			}

			return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
		}
	}
}

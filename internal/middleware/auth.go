package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/leconcepteur/cloud-cicd-poc/internal/auth"
)

const (
	SessionCookieName = "session_id"
	UserIDKey         = "user_id"
	UsernameKey       = "username"
)

func AuthMiddleware(authService *auth.Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var sessionID string

			// Try cookie first
			cookie, err := c.Cookie(SessionCookieName)
			if err == nil {
				sessionID = cookie.Value
			}

			// Try Authorization header
			if sessionID == "" {
				authHeader := c.Request().Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					sessionID = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if sessionID == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "unauthorized",
				})
			}

			session, err := authService.GetSession(c.Request().Context(), sessionID)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "invalid session",
				})
			}

			c.Set(UserIDKey, session.UserID)
			c.Set(UsernameKey, session.Username)

			return next(c)
		}
	}
}

func GetUserID(c echo.Context) string {
	if userID, ok := c.Get(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

func GetUsername(c echo.Context) string {
	if username, ok := c.Get(UsernameKey).(string); ok {
		return username
	}
	return ""
}

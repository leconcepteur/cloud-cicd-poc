package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/leconcepteur/cloud-cicd-poc/internal/auth"
	"github.com/leconcepteur/cloud-cicd-poc/internal/middleware"
	"github.com/leconcepteur/cloud-cicd-poc/internal/models"
)

type AuthHandler struct {
	authService   *auth.Service
	sessionMaxAge int
}

func NewAuthHandler(authService *auth.Service, sessionMaxAge int) *AuthHandler {
	return &AuthHandler{
		authService:   authService,
		sessionMaxAge: sessionMaxAge,
	}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req models.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	user, err := h.authService.Register(c.Request().Context(), &req)
	if err != nil {
		if errors.Is(err, auth.ErrUsernameExists) {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "username already exists",
			})
		}
		if errors.Is(err, auth.ErrInvalidUsername) || errors.Is(err, auth.ErrInvalidPassword) {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to register user",
		})
	}

	return c.JSON(http.StatusCreated, models.UserResponse{
		ID:       user.ID,
		Username: user.Username,
	})
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req models.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	user, sessionID, err := h.authService.Login(c.Request().Context(), &req)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "invalid credentials",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to login",
		})
	}

	cookie := &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   h.sessionMaxAge,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user": models.UserResponse{
			ID:       user.ID,
			Username: user.Username,
		},
		"session_id": sessionID,
	})
}

func (h *AuthHandler) Logout(c echo.Context) error {
	cookie, err := c.Cookie(middleware.SessionCookieName)
	if err == nil {
		_ = h.authService.Logout(c.Request().Context(), cookie.Value)
	}

	// Clear the cookie
	c.SetCookie(&http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})

	return c.JSON(http.StatusOK, map[string]string{
		"message": "logged out",
	})
}

func (h *AuthHandler) Me(c echo.Context) error {
	userID := middleware.GetUserID(c)
	username := middleware.GetUsername(c)

	return c.JSON(http.StatusOK, models.UserResponse{
		ID:       userID,
		Username: username,
	})
}

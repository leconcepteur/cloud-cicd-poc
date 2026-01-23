package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/leconcepteur/cloud-cicd-poc/internal/auth"
	"github.com/leconcepteur/cloud-cicd-poc/internal/middleware"
)

type StatsHandler struct {
	authService *auth.Service
}

func NewStatsHandler(authService *auth.Service) *StatsHandler {
	return &StatsHandler{
		authService: authService,
	}
}

func (h *StatsHandler) GetStats(c echo.Context) error {
	userID := middleware.GetUserID(c)

	stats, err := h.authService.GetUserStats(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to get stats",
		})
	}

	return c.JSON(http.StatusOK, stats)
}

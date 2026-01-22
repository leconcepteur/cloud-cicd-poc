package api

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/leconcepteur/cloud-cicd-poc/internal/matchmaking"
	"github.com/leconcepteur/cloud-cicd-poc/internal/middleware"
)

type MatchmakingHandler struct {
	matchService *matchmaking.Service
}

func NewMatchmakingHandler(matchService *matchmaking.Service) *MatchmakingHandler {
	return &MatchmakingHandler{
		matchService: matchService,
	}
}

func (h *MatchmakingHandler) Join(c echo.Context) error {
	userID := middleware.GetUserID(c)
	username := middleware.GetUsername(c)

	err := h.matchService.JoinQueue(c.Request().Context(), userID, username)
	if err != nil {
		if errors.Is(err, matchmaking.ErrAlreadyInQueue) {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "already in queue",
			})
		}
		if errors.Is(err, matchmaking.ErrAlreadyInGame) {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "already in an active game",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to join queue",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "joined queue",
	})
}

func (h *MatchmakingHandler) Leave(c echo.Context) error {
	userID := middleware.GetUserID(c)

	err := h.matchService.LeaveQueue(c.Request().Context(), userID)
	if err != nil {
		if errors.Is(err, matchmaking.ErrNotInQueue) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "not in queue",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to leave queue",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "left queue",
	})
}

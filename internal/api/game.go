package api

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/leconcepteur/cloud-cicd-poc/internal/game"
	"github.com/leconcepteur/cloud-cicd-poc/internal/matchmaking"
	"github.com/leconcepteur/cloud-cicd-poc/internal/middleware"
	"github.com/leconcepteur/cloud-cicd-poc/internal/models"
)

type GameHandler struct {
	gameService *game.Service
	eventHub    *matchmaking.EventHub
}

func NewGameHandler(gameService *game.Service, eventHub *matchmaking.EventHub) *GameHandler {
	return &GameHandler{
		gameService: gameService,
		eventHub:    eventHub,
	}
}

func (h *GameHandler) Ready(c echo.Context) error {
	userID := middleware.GetUserID(c)

	// Get the user's active game
	activeGame, err := h.gameService.GetActiveGameForUser(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to get active game",
		})
	}

	if activeGame == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "no active game found",
		})
	}

	updatedGame, err := h.gameService.SetReady(c.Request().Context(), activeGame.ID, userID)
	if err != nil {
		if errors.Is(err, game.ErrGameNotStarted) {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "game already started or finished",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to set ready",
		})
	}

	// Notify opponent
	opponentID := updatedGame.PlayerX
	if opponentID == userID {
		opponentID = updatedGame.PlayerO
	}

	h.eventHub.SendToUser(opponentID, models.SSEEvent{
		Type: models.EventTypeOpponentReady,
		Data: map[string]bool{"opponent_ready": true},
	})

	// If game started, notify both players
	if updatedGame.Status == models.GameStatusInProgress {
		h.notifyGameUpdate(updatedGame)
	}

	return c.JSON(http.StatusOK, h.buildGameStateResponse(updatedGame, userID))
}

func (h *GameHandler) Move(c echo.Context) error {
	userID := middleware.GetUserID(c)

	var req models.MoveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	// Get the user's active game
	activeGame, err := h.gameService.GetActiveGameForUser(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to get active game",
		})
	}

	if activeGame == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "no active game found",
		})
	}

	updatedGame, err := h.gameService.MakeMove(c.Request().Context(), activeGame.ID, userID, req.Position)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, game.ErrNotYourTurn) || errors.Is(err, game.ErrInvalidMove) ||
			errors.Is(err, game.ErrGameNotStarted) || errors.Is(err, game.ErrGameFinished) {
			statusCode = http.StatusBadRequest
		}
		return c.JSON(statusCode, map[string]string{
			"error": err.Error(),
		})
	}

	// Notify both players about the update
	if updatedGame.Status == models.GameStatusFinished {
		h.notifyGameEnd(updatedGame)
	} else {
		h.notifyGameUpdate(updatedGame)
	}

	return c.JSON(http.StatusOK, h.buildGameStateResponse(updatedGame, userID))
}

func (h *GameHandler) Forfeit(c echo.Context) error {
	userID := middleware.GetUserID(c)

	activeGame, err := h.gameService.GetActiveGameForUser(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to get active game",
		})
	}

	if activeGame == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "no active game found",
		})
	}

	updatedGame, err := h.gameService.Forfeit(c.Request().Context(), activeGame.ID, userID)
	if err != nil {
		if errors.Is(err, game.ErrGameFinished) {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "game already finished",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to forfeit",
		})
	}

	h.notifyGameEnd(updatedGame)

	return c.JSON(http.StatusOK, h.buildGameStateResponse(updatedGame, userID))
}

func (h *GameHandler) GetState(c echo.Context) error {
	userID := middleware.GetUserID(c)

	activeGame, err := h.gameService.GetActiveGameForUser(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to get active game",
		})
	}

	if activeGame == nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"game": nil,
		})
	}

	return c.JSON(http.StatusOK, h.buildGameStateResponse(activeGame, userID))
}

func (h *GameHandler) buildGameStateResponse(g *models.Game, userID string) *models.GameStateResponse {
	symbol := game.GetPlayerSymbol(g, userID)
	return &models.GameStateResponse{
		Game:       g,
		YourSymbol: symbol,
		IsYourTurn: g.CurrentTurn == userID && g.Status == models.GameStatusInProgress,
	}
}

func (h *GameHandler) notifyGameUpdate(g *models.Game) {
	// Notify player X
	h.eventHub.SendToUser(g.PlayerX, models.SSEEvent{
		Type: models.EventTypeGameUpdate,
		Data: models.GameUpdateData{
			Game:       g,
			YourSymbol: "X",
			IsYourTurn: g.CurrentTurn == g.PlayerX,
		},
	})

	// Notify player O
	h.eventHub.SendToUser(g.PlayerO, models.SSEEvent{
		Type: models.EventTypeGameUpdate,
		Data: models.GameUpdateData{
			Game:       g,
			YourSymbol: "O",
			IsYourTurn: g.CurrentTurn == g.PlayerO,
		},
	})
}

func (h *GameHandler) notifyGameEnd(g *models.Game) {
	// Notify player X
	h.eventHub.SendToUser(g.PlayerX, models.SSEEvent{
		Type: models.EventTypeGameEnd,
		Data: models.GameEndData{
			Game:       g,
			YourSymbol: "X",
			YouWon:     g.Winner == g.PlayerX,
			IsDraw:     g.Result == models.GameResultDraw,
		},
	})

	// Notify player O
	h.eventHub.SendToUser(g.PlayerO, models.SSEEvent{
		Type: models.EventTypeGameEnd,
		Data: models.GameEndData{
			Game:       g,
			YourSymbol: "O",
			YouWon:     g.Winner == g.PlayerO,
			IsDraw:     g.Result == models.GameResultDraw,
		},
	})
}

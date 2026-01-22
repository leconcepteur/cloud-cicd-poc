package matchmaking

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/leconcepteur/cloud-cicd-poc/internal/game"
	"github.com/leconcepteur/cloud-cicd-poc/internal/models"
)

var (
	ErrAlreadyInQueue = errors.New("already in queue")
	ErrNotInQueue     = errors.New("not in queue")
	ErrAlreadyInGame  = errors.New("already in an active game")
)

const (
	readyTimeout    = 60 * time.Second
	matchCheckDelay = 500 * time.Millisecond
)

type Service struct {
	queue       *Queue
	gameService *game.Service
	eventHub    *EventHub
	mu          sync.RWMutex
}

func NewService(queue *Queue, gameService *game.Service, eventHub *EventHub) *Service {
	return &Service{
		queue:       queue,
		gameService: gameService,
		eventHub:    eventHub,
	}
}

func (s *Service) JoinQueue(ctx context.Context, userID, username string) error {
	// Check if already in an active game
	activeGame, err := s.gameService.GetActiveGameForUser(ctx, userID)
	if err != nil {
		return err
	}
	if activeGame != nil {
		return ErrAlreadyInGame
	}

	// Check if already in queue
	inQueue, err := s.queue.IsInQueue(ctx, userID)
	if err != nil {
		return err
	}
	if inQueue {
		return ErrAlreadyInQueue
	}

	// Join the queue
	if err := s.queue.Join(ctx, userID, username); err != nil {
		return err
	}

	// Try to find a match
	go s.tryMatch(context.Background())

	return nil
}

func (s *Service) LeaveQueue(ctx context.Context, userID string) error {
	inQueue, err := s.queue.IsInQueue(ctx, userID)
	if err != nil {
		return err
	}
	if !inQueue {
		return ErrNotInQueue
	}

	return s.queue.Leave(ctx, userID)
}

func (s *Service) tryMatch(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	player1, player2, err := s.queue.FindMatch(ctx)
	if err != nil || player1 == nil || player2 == nil {
		return
	}

	// Create a new game
	newGame, err := s.gameService.CreateGame(ctx, player1.UserID, player2.UserID)
	if err != nil {
		// Put players back in queue
		s.queue.Join(ctx, player1.UserID, player1.Username)
		s.queue.Join(ctx, player2.UserID, player2.Username)
		return
	}

	// Notify both players
	s.eventHub.SendToUser(player1.UserID, models.SSEEvent{
		Type: models.EventTypeMatchFound,
		Data: models.MatchFoundEvent{
			GameID:   newGame.ID,
			Opponent: player2.Username,
		},
	})

	s.eventHub.SendToUser(player2.UserID, models.SSEEvent{
		Type: models.EventTypeMatchFound,
		Data: models.MatchFoundEvent{
			GameID:   newGame.ID,
			Opponent: player1.Username,
		},
	})

	// Start ready timeout
	go s.handleReadyTimeout(context.Background(), newGame.ID, player1.UserID, player2.UserID)
}

func (s *Service) handleReadyTimeout(ctx context.Context, gameID, player1ID, player2ID string) {
	time.Sleep(readyTimeout)

	gameState, err := s.gameService.GetGame(ctx, gameID)
	if err != nil {
		return
	}

	// If game is still waiting, timeout occurred
	if gameState.Status == models.GameStatusWaiting {
		// Notify both players about timeout
		s.eventHub.SendToUser(player1ID, models.SSEEvent{
			Type: models.EventTypeReadyTimeout,
			Data: map[string]string{"message": "Ready timeout - returning to queue"},
		})
		s.eventHub.SendToUser(player2ID, models.SSEEvent{
			Type: models.EventTypeReadyTimeout,
			Data: map[string]string{"message": "Ready timeout - returning to queue"},
		})

		// Could re-queue players here if desired
	}
}

func (s *Service) StartMatchmaking(ctx context.Context) {
	ticker := time.NewTicker(matchCheckDelay)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.tryMatch(ctx)
		}
	}
}

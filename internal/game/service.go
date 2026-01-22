package game

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/leconcepteur/cloud-cicd-poc/internal/database"
	"github.com/leconcepteur/cloud-cicd-poc/internal/models"
)

var (
	ErrGameNotFound   = errors.New("game not found")
	ErrNotYourTurn    = errors.New("not your turn")
	ErrInvalidMove    = errors.New("invalid move")
	ErrGameNotStarted = errors.New("game not started")
	ErrGameFinished   = errors.New("game already finished")
	ErrNotInGame      = errors.New("you are not in this game")
)

type Service struct {
	db *database.PostgresDB
}

func NewService(db *database.PostgresDB) *Service {
	return &Service{db: db}
}

func (s *Service) CreateGame(ctx context.Context, playerX, playerO string) (*models.Game, error) {
	game := &models.Game{
		ID:          uuid.New().String(),
		PlayerX:     playerX,
		PlayerO:     playerO,
		Board:       [9]string{},
		CurrentTurn: playerX, // X goes first
		Status:      models.GameStatusWaiting,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	boardStr := boardToString(game.Board)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO games (id, player_x, player_o, board, current_turn, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		game.ID, game.PlayerX, game.PlayerO, boardStr, game.CurrentTurn, game.Status, game.CreatedAt, game.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	return game, nil
}

func (s *Service) GetGame(ctx context.Context, gameID string) (*models.Game, error) {
	var game models.Game
	var boardStr string
	var winner sql.NullString

	err := s.db.QueryRowContext(ctx,
		`SELECT id, player_x, player_o, board, current_turn, status, result, winner,
		        player_x_ready, player_o_ready, created_at, updated_at
		 FROM games WHERE id = $1`,
		gameID,
	).Scan(
		&game.ID, &game.PlayerX, &game.PlayerO, &boardStr, &game.CurrentTurn,
		&game.Status, &game.Result, &winner, &game.PlayerXReady, &game.PlayerOReady,
		&game.CreatedAt, &game.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGameNotFound
		}
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	game.Board = stringToBoard(boardStr)
	if winner.Valid {
		game.Winner = winner.String
	}

	return &game, nil
}

func (s *Service) GetActiveGameForUser(ctx context.Context, userID string) (*models.Game, error) {
	var game models.Game
	var boardStr string
	var winner sql.NullString

	err := s.db.QueryRowContext(ctx,
		`SELECT id, player_x, player_o, board, current_turn, status, result, winner,
		        player_x_ready, player_o_ready, created_at, updated_at
		 FROM games
		 WHERE (player_x = $1 OR player_o = $1) AND status != 'finished'
		 ORDER BY created_at DESC LIMIT 1`,
		userID,
	).Scan(
		&game.ID, &game.PlayerX, &game.PlayerO, &boardStr, &game.CurrentTurn,
		&game.Status, &game.Result, &winner, &game.PlayerXReady, &game.PlayerOReady,
		&game.CreatedAt, &game.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get active game: %w", err)
	}

	game.Board = stringToBoard(boardStr)
	if winner.Valid {
		game.Winner = winner.String
	}

	return &game, nil
}

func (s *Service) SetReady(ctx context.Context, gameID, userID string) (*models.Game, error) {
	game, err := s.GetGame(ctx, gameID)
	if err != nil {
		return nil, err
	}

	if game.PlayerX != userID && game.PlayerO != userID {
		return nil, ErrNotInGame
	}

	if game.Status != models.GameStatusWaiting {
		return nil, ErrGameNotStarted
	}

	var readyField string
	if game.PlayerX == userID {
		readyField = "player_x_ready"
		game.PlayerXReady = true
	} else {
		readyField = "player_o_ready"
		game.PlayerOReady = true
	}

	_, err = s.db.ExecContext(ctx,
		fmt.Sprintf(`UPDATE games SET %s = TRUE, updated_at = $1 WHERE id = $2`, readyField),
		time.Now(), gameID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set ready: %w", err)
	}

	// Check if both players are ready
	if game.PlayerXReady && game.PlayerOReady {
		_, err = s.db.ExecContext(ctx,
			`UPDATE games SET status = $1, updated_at = $2 WHERE id = $3`,
			models.GameStatusInProgress, time.Now(), gameID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to start game: %w", err)
		}
		game.Status = models.GameStatusInProgress
	}

	return game, nil
}

func (s *Service) MakeMove(ctx context.Context, gameID, userID string, position int) (*models.Game, error) {
	game, err := s.GetGame(ctx, gameID)
	if err != nil {
		return nil, err
	}

	if game.PlayerX != userID && game.PlayerO != userID {
		return nil, ErrNotInGame
	}

	if game.Status != models.GameStatusInProgress {
		if game.Status == models.GameStatusFinished {
			return nil, ErrGameFinished
		}
		return nil, ErrGameNotStarted
	}

	if !IsPlayerTurn(game, userID) {
		return nil, ErrNotYourTurn
	}

	if !IsValidMove(game.Board, position) {
		return nil, ErrInvalidMove
	}

	// Make the move
	symbol := GetPlayerSymbol(game, userID)
	game.Board[position] = symbol

	// Check for winner
	result, _ := CheckWinner(game.Board)
	if result != models.GameResultNone {
		game.Status = models.GameStatusFinished
		game.Result = result
		if result == models.GameResultWinX {
			game.Winner = game.PlayerX
		} else if result == models.GameResultWinO {
			game.Winner = game.PlayerO
		}
	} else {
		// Switch turns
		if game.CurrentTurn == game.PlayerX {
			game.CurrentTurn = game.PlayerO
		} else {
			game.CurrentTurn = game.PlayerX
		}
	}

	game.UpdatedAt = time.Now()

	boardStr := boardToString(game.Board)
	_, err = s.db.ExecContext(ctx,
		`UPDATE games SET board = $1, current_turn = $2, status = $3, result = $4, winner = $5, updated_at = $6 WHERE id = $7`,
		boardStr, game.CurrentTurn, game.Status, game.Result, sql.NullString{String: game.Winner, Valid: game.Winner != ""}, game.UpdatedAt, gameID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	// Update stats if game finished
	if game.Status == models.GameStatusFinished {
		if err := s.updateStats(ctx, game); err != nil {
			return nil, err
		}
	}

	return game, nil
}

func (s *Service) Forfeit(ctx context.Context, gameID, userID string) (*models.Game, error) {
	game, err := s.GetGame(ctx, gameID)
	if err != nil {
		return nil, err
	}

	if game.PlayerX != userID && game.PlayerO != userID {
		return nil, ErrNotInGame
	}

	if game.Status == models.GameStatusFinished {
		return nil, ErrGameFinished
	}

	// The player who forfeits loses
	game.Status = models.GameStatusFinished
	if game.PlayerX == userID {
		game.Result = models.GameResultWinO
		game.Winner = game.PlayerO
	} else {
		game.Result = models.GameResultWinX
		game.Winner = game.PlayerX
	}
	game.UpdatedAt = time.Now()

	_, err = s.db.ExecContext(ctx,
		`UPDATE games SET status = $1, result = $2, winner = $3, updated_at = $4 WHERE id = $5`,
		game.Status, game.Result, game.Winner, game.UpdatedAt, gameID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to forfeit game: %w", err)
	}

	if err := s.updateStats(ctx, game); err != nil {
		return nil, err
	}

	return game, nil
}

func (s *Service) updateStats(ctx context.Context, game *models.Game) error {
	if game.Result == models.GameResultDraw {
		_, err := s.db.ExecContext(ctx,
			`UPDATE user_stats SET ties = ties + 1 WHERE user_id IN ($1, $2)`,
			game.PlayerX, game.PlayerO,
		)
		return err
	}

	var winner, loser string
	if game.Result == models.GameResultWinX {
		winner = game.PlayerX
		loser = game.PlayerO
	} else {
		winner = game.PlayerO
		loser = game.PlayerX
	}

	_, err := s.db.ExecContext(ctx, `UPDATE user_stats SET wins = wins + 1 WHERE user_id = $1`, winner)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `UPDATE user_stats SET losses = losses + 1 WHERE user_id = $1`, loser)
	return err
}

func boardToString(board [9]string) string {
	var sb strings.Builder
	for _, cell := range board {
		if cell == "" {
			sb.WriteRune(' ')
		} else {
			sb.WriteString(cell)
		}
	}
	return sb.String()
}

func stringToBoard(s string) [9]string {
	var board [9]string
	for i, c := range s {
		if i >= 9 {
			break
		}
		if c != ' ' {
			board[i] = string(c)
		}
	}
	return board
}

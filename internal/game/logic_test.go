package game

import (
	"testing"

	"github.com/leconcepteur/cloud-cicd-poc/internal/models"
)

func TestCheckWinner_XWinsRow(t *testing.T) {
	tests := []struct {
		name   string
		board  [9]string
		result models.GameResult
		winner string
	}{
		{
			name:   "X wins top row",
			board:  [9]string{"X", "X", "X", "", "", "", "", "", ""},
			result: models.GameResultWinX,
			winner: "X",
		},
		{
			name:   "X wins middle row",
			board:  [9]string{"", "", "", "X", "X", "X", "", "", ""},
			result: models.GameResultWinX,
			winner: "X",
		},
		{
			name:   "X wins bottom row",
			board:  [9]string{"", "", "", "", "", "", "X", "X", "X"},
			result: models.GameResultWinX,
			winner: "X",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, winner := CheckWinner(tt.board)
			if result != tt.result {
				t.Errorf("expected result %v, got %v", tt.result, result)
			}
			if winner != tt.winner {
				t.Errorf("expected winner %v, got %v", tt.winner, winner)
			}
		})
	}
}

func TestCheckWinner_OWinsColumn(t *testing.T) {
	tests := []struct {
		name   string
		board  [9]string
		result models.GameResult
		winner string
	}{
		{
			name:   "O wins left column",
			board:  [9]string{"O", "", "", "O", "", "", "O", "", ""},
			result: models.GameResultWinO,
			winner: "O",
		},
		{
			name:   "O wins middle column",
			board:  [9]string{"", "O", "", "", "O", "", "", "O", ""},
			result: models.GameResultWinO,
			winner: "O",
		},
		{
			name:   "O wins right column",
			board:  [9]string{"", "", "O", "", "", "O", "", "", "O"},
			result: models.GameResultWinO,
			winner: "O",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, winner := CheckWinner(tt.board)
			if result != tt.result {
				t.Errorf("expected result %v, got %v", tt.result, result)
			}
			if winner != tt.winner {
				t.Errorf("expected winner %v, got %v", tt.winner, winner)
			}
		})
	}
}

func TestCheckWinner_Diagonal(t *testing.T) {
	tests := []struct {
		name   string
		board  [9]string
		result models.GameResult
		winner string
	}{
		{
			name:   "X wins diagonal",
			board:  [9]string{"X", "", "", "", "X", "", "", "", "X"},
			result: models.GameResultWinX,
			winner: "X",
		},
		{
			name:   "O wins anti-diagonal",
			board:  [9]string{"", "", "O", "", "O", "", "O", "", ""},
			result: models.GameResultWinO,
			winner: "O",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, winner := CheckWinner(tt.board)
			if result != tt.result {
				t.Errorf("expected result %v, got %v", tt.result, result)
			}
			if winner != tt.winner {
				t.Errorf("expected winner %v, got %v", tt.winner, winner)
			}
		})
	}
}

func TestCheckWinner_Draw(t *testing.T) {
	// Full board with no winner
	board := [9]string{"X", "O", "X", "X", "O", "O", "O", "X", "X"}
	result, winner := CheckWinner(board)

	if result != models.GameResultDraw {
		t.Errorf("expected draw, got %v", result)
	}
	if winner != "" {
		t.Errorf("expected no winner, got %v", winner)
	}
}

func TestCheckWinner_InProgress(t *testing.T) {
	// Game still in progress
	board := [9]string{"X", "O", "", "", "X", "", "", "", ""}
	result, winner := CheckWinner(board)

	if result != models.GameResultNone {
		t.Errorf("expected none, got %v", result)
	}
	if winner != "" {
		t.Errorf("expected no winner, got %v", winner)
	}
}

func TestCheckWinner_EmptyBoard(t *testing.T) {
	board := [9]string{}
	result, winner := CheckWinner(board)

	if result != models.GameResultNone {
		t.Errorf("expected none, got %v", result)
	}
	if winner != "" {
		t.Errorf("expected no winner, got %v", winner)
	}
}

func TestIsValidMove(t *testing.T) {
	tests := []struct {
		name     string
		board    [9]string
		position int
		valid    bool
	}{
		{
			name:     "valid move on empty cell",
			board:    [9]string{"X", "", "", "", "", "", "", "", ""},
			position: 1,
			valid:    true,
		},
		{
			name:     "invalid move on occupied cell",
			board:    [9]string{"X", "", "", "", "", "", "", "", ""},
			position: 0,
			valid:    false,
		},
		{
			name:     "invalid position negative",
			board:    [9]string{},
			position: -1,
			valid:    false,
		},
		{
			name:     "invalid position too high",
			board:    [9]string{},
			position: 9,
			valid:    false,
		},
		{
			name:     "valid move on position 0",
			board:    [9]string{},
			position: 0,
			valid:    true,
		},
		{
			name:     "valid move on position 8",
			board:    [9]string{},
			position: 8,
			valid:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidMove(tt.board, tt.position)
			if result != tt.valid {
				t.Errorf("expected %v, got %v", tt.valid, result)
			}
		})
	}
}

func TestGetPlayerSymbol(t *testing.T) {
	game := &models.Game{
		PlayerX: "user1",
		PlayerO: "user2",
	}

	tests := []struct {
		name     string
		userID   string
		expected string
	}{
		{"player X", "user1", "X"},
		{"player O", "user2", "O"},
		{"non-player", "user3", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPlayerSymbol(game, tt.userID)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsPlayerTurn(t *testing.T) {
	game := &models.Game{
		PlayerX:     "user1",
		PlayerO:     "user2",
		CurrentTurn: "user1",
	}

	tests := []struct {
		name     string
		userID   string
		expected bool
	}{
		{"is player turn", "user1", true},
		{"not player turn", "user2", false},
		{"non-player", "user3", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPlayerTurn(game, tt.userID)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

package game

import "github.com/leconcepteur/cloud-cicd-poc/internal/models"

var winPatterns = [][]int{
	{0, 1, 2}, // top row
	{3, 4, 5}, // middle row
	{6, 7, 8}, // bottom row
	{0, 3, 6}, // left column
	{1, 4, 7}, // middle column
	{2, 5, 8}, // right column
	{0, 4, 8}, // diagonal
	{2, 4, 6}, // anti-diagonal
}

func CheckWinner(board [9]string) (models.GameResult, string) {
	for _, pattern := range winPatterns {
		a, b, c := board[pattern[0]], board[pattern[1]], board[pattern[2]]
		if a != "" && a == b && b == c {
			if a == "X" {
				return models.GameResultWinX, a
			}
			return models.GameResultWinO, a
		}
	}

	// Check for draw
	isFull := true
	for _, cell := range board {
		if cell == "" {
			isFull = false
			break
		}
	}

	if isFull {
		return models.GameResultDraw, ""
	}

	return models.GameResultNone, ""
}

func IsValidMove(board [9]string, position int) bool {
	if position < 0 || position > 8 {
		return false
	}
	return board[position] == ""
}

func GetPlayerSymbol(game *models.Game, userID string) string {
	if game.PlayerX == userID {
		return "X"
	}
	if game.PlayerO == userID {
		return "O"
	}
	return ""
}

func IsPlayerTurn(game *models.Game, userID string) bool {
	return game.CurrentTurn == userID
}

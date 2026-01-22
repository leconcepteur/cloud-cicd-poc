package models

import "time"

type GameStatus string

const (
	GameStatusWaiting    GameStatus = "waiting"
	GameStatusInProgress GameStatus = "in_progress"
	GameStatusFinished   GameStatus = "finished"
)

type GameResult string

const (
	GameResultNone GameResult = ""
	GameResultWinX GameResult = "win_x"
	GameResultWinO GameResult = "win_o"
	GameResultDraw GameResult = "draw"
)

type Game struct {
	ID           string     `json:"id"`
	PlayerX      string     `json:"player_x"`
	PlayerO      string     `json:"player_o"`
	Board        [9]string  `json:"board"`
	CurrentTurn  string     `json:"current_turn"`
	Status       GameStatus `json:"status"`
	Result       GameResult `json:"result"`
	Winner       string     `json:"winner,omitempty"`
	PlayerXReady bool       `json:"player_x_ready"`
	PlayerOReady bool       `json:"player_o_ready"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type MoveRequest struct {
	Position int `json:"position" validate:"min=0,max=8"`
}

type GameStateResponse struct {
	Game       *Game  `json:"game"`
	YourSymbol string `json:"your_symbol"`
	IsYourTurn bool   `json:"is_your_turn"`
}

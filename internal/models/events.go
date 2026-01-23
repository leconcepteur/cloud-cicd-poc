package models

type EventType string

const (
	EventTypeMatchFound      EventType = "match_found"
	EventTypeGameStart       EventType = "game_start"
	EventTypeGameUpdate      EventType = "game_update"
	EventTypeGameEnd         EventType = "game_end"
	EventTypeOpponentReady   EventType = "opponent_ready"
	EventTypeOpponentLeft    EventType = "opponent_left"
	EventTypeQueueUpdate     EventType = "queue_update"
	EventTypeError           EventType = "error"
	EventTypeReadyTimeout    EventType = "ready_timeout"
	EventTypeDisconnectGrace EventType = "disconnect_grace"
)

type SSEEvent struct {
	Type EventType   `json:"type"`
	Data interface{} `json:"data"`
}

type GameUpdateData struct {
	Game       *Game  `json:"game"`
	YourSymbol string `json:"your_symbol"`
	IsYourTurn bool   `json:"is_your_turn"`
}

type GameEndData struct {
	Game       *Game  `json:"game"`
	YourSymbol string `json:"your_symbol"`
	YouWon     bool   `json:"you_won"`
	IsDraw     bool   `json:"is_draw"`
}

type ErrorData struct {
	Message string `json:"message"`
}

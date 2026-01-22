package models

import "time"

type QueueEntry struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	JoinedAt  time.Time `json:"joined_at"`
}

type MatchFoundEvent struct {
	GameID   string `json:"game_id"`
	Opponent string `json:"opponent"`
}

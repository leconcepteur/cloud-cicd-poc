package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/leconcepteur/cloud-cicd-poc/internal/database"
	"github.com/leconcepteur/cloud-cicd-poc/internal/models"
)

const sessionPrefix = "session:"

type SessionManager struct {
	redis  *database.RedisDB
	maxAge time.Duration
}

type SessionData struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

func NewSessionManager(redis *database.RedisDB, maxAgeSeconds int) *SessionManager {
	return &SessionManager{
		redis:  redis,
		maxAge: time.Duration(maxAgeSeconds) * time.Second,
	}
}

func (sm *SessionManager) Create(ctx context.Context, user *models.User) (string, error) {
	sessionID := uuid.New().String()

	data := SessionData{
		UserID:   user.ID,
		Username: user.Username,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal session data: %w", err)
	}

	key := sessionPrefix + sessionID
	if err := sm.redis.Set(ctx, key, jsonData, sm.maxAge).Err(); err != nil {
		return "", fmt.Errorf("failed to store session: %w", err)
	}

	return sessionID, nil
}

func (sm *SessionManager) Get(ctx context.Context, sessionID string) (*SessionData, error) {
	key := sessionPrefix + sessionID

	jsonData, err := sm.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	var data SessionData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return &data, nil
}

func (sm *SessionManager) Delete(ctx context.Context, sessionID string) error {
	key := sessionPrefix + sessionID
	return sm.redis.Del(ctx, key).Err()
}

func (sm *SessionManager) Refresh(ctx context.Context, sessionID string) error {
	key := sessionPrefix + sessionID
	return sm.redis.Expire(ctx, key, sm.maxAge).Err()
}

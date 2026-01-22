package matchmaking

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/leconcepteur/cloud-cicd-poc/internal/database"
	"github.com/leconcepteur/cloud-cicd-poc/internal/models"
)

const (
	queueKey       = "matchmaking:queue"
	queueEntryTTL  = 5 * time.Minute
)

type Queue struct {
	redis *database.RedisDB
}

func NewQueue(redis *database.RedisDB) *Queue {
	return &Queue{redis: redis}
}

func (q *Queue) Join(ctx context.Context, userID, username string) error {
	entry := models.QueueEntry{
		UserID:   userID,
		Username: username,
		JoinedAt: time.Now(),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal queue entry: %w", err)
	}

	// Use sorted set with timestamp as score
	score := float64(entry.JoinedAt.UnixNano())
	if err := q.redis.ZAdd(ctx, queueKey, float64(score), string(data)).Err(); err != nil {
		return fmt.Errorf("failed to add to queue: %w", err)
	}

	return nil
}

func (q *Queue) Leave(ctx context.Context, userID string) error {
	// Get all entries and find the one with matching userID
	entries, err := q.redis.ZRange(ctx, queueKey, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("failed to get queue entries: %w", err)
	}

	for _, entryStr := range entries {
		var entry models.QueueEntry
		if err := json.Unmarshal([]byte(entryStr), &entry); err != nil {
			continue
		}
		if entry.UserID == userID {
			q.redis.ZRem(ctx, queueKey, entryStr)
			break
		}
	}

	return nil
}

func (q *Queue) IsInQueue(ctx context.Context, userID string) (bool, error) {
	entries, err := q.redis.ZRange(ctx, queueKey, 0, -1).Result()
	if err != nil {
		return false, fmt.Errorf("failed to get queue entries: %w", err)
	}

	for _, entryStr := range entries {
		var entry models.QueueEntry
		if err := json.Unmarshal([]byte(entryStr), &entry); err != nil {
			continue
		}
		if entry.UserID == userID {
			return true, nil
		}
	}

	return false, nil
}

func (q *Queue) FindMatch(ctx context.Context) (*models.QueueEntry, *models.QueueEntry, error) {
	// Get first two entries from the queue
	entries, err := q.redis.ZRange(ctx, queueKey, 0, 1).Result()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get queue entries: %w", err)
	}

	if len(entries) < 2 {
		return nil, nil, nil // Not enough players
	}

	var player1, player2 models.QueueEntry
	if err := json.Unmarshal([]byte(entries[0]), &player1); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal player1: %w", err)
	}
	if err := json.Unmarshal([]byte(entries[1]), &player2); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal player2: %w", err)
	}

	// Remove both from queue
	q.redis.ZRem(ctx, queueKey, entries[0], entries[1])

	return &player1, &player2, nil
}

func (q *Queue) GetQueueLength(ctx context.Context) (int64, error) {
	return q.redis.ZCard(ctx, queueKey).Result()
}

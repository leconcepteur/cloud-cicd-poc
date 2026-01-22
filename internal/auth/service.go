package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/leconcepteur/cloud-cicd-poc/internal/database"
	"github.com/leconcepteur/cloud-cicd-poc/internal/models"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUsernameExists     = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidUsername    = errors.New("username must be 3-20 characters, alphanumeric and underscores only")
	ErrInvalidPassword    = errors.New("password must be at least 8 characters")
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)

type Service struct {
	db      *database.PostgresDB
	session *SessionManager
}

func NewService(db *database.PostgresDB, session *SessionManager) *Service {
	return &Service{
		db:      db,
		session: session,
	}
}

func (s *Service) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	if !usernameRegex.MatchString(req.Username) {
		return nil, ErrInvalidUsername
	}

	if len(req.Password) < 8 {
		return nil, ErrInvalidPassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		ID:        uuid.New().String(),
		Username:  req.Username,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO users (id, username, password, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`,
		user.ID, user.Username, user.Password, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrUsernameExists
		}
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO user_stats (user_id, wins, losses, ties) VALUES ($1, 0, 0, 0)`,
		user.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert user stats: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return user, nil
}

func (s *Service) Login(ctx context.Context, req *models.LoginRequest) (*models.User, string, error) {
	var user models.User
	err := s.db.QueryRowContext(ctx,
		`SELECT id, username, password, created_at, updated_at FROM users WHERE username = $1`,
		req.Username,
	).Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", fmt.Errorf("failed to query user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	sessionID, err := s.session.Create(ctx, &user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create session: %w", err)
	}

	return &user, sessionID, nil
}

func (s *Service) Logout(ctx context.Context, sessionID string) error {
	return s.session.Delete(ctx, sessionID)
}

func (s *Service) GetSession(ctx context.Context, sessionID string) (*SessionData, error) {
	return s.session.Get(ctx, sessionID)
}

func (s *Service) GetUser(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	err := s.db.QueryRowContext(ctx,
		`SELECT id, username, password, created_at, updated_at FROM users WHERE id = $1`,
		userID,
	).Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return &user, nil
}

func (s *Service) GetUserStats(ctx context.Context, userID string) (*models.UserStats, error) {
	var stats models.UserStats
	err := s.db.QueryRowContext(ctx,
		`SELECT user_id, wins, losses, ties FROM user_stats WHERE user_id = $1`,
		userID,
	).Scan(&stats.UserID, &stats.Wins, &stats.Losses, &stats.Ties)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query user stats: %w", err)
	}

	return &stats, nil
}

func isUniqueViolation(err error) bool {
	return err != nil && (contains(err.Error(), "unique") || contains(err.Error(), "duplicate"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

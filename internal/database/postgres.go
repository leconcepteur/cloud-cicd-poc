package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type PostgresDB struct {
	*sql.DB
}

func NewPostgres(databaseURL string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresDB{db}, nil
}

func (db *PostgresDB) Migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(36) PRIMARY KEY,
			username VARCHAR(20) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS user_stats (
			user_id VARCHAR(36) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			wins INTEGER DEFAULT 0,
			losses INTEGER DEFAULT 0,
			ties INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS games (
			id VARCHAR(36) PRIMARY KEY,
			player_x VARCHAR(36) REFERENCES users(id),
			player_o VARCHAR(36) REFERENCES users(id),
			board VARCHAR(9) DEFAULT '         ',
			current_turn VARCHAR(36),
			status VARCHAR(20) DEFAULT 'waiting',
			result VARCHAR(20) DEFAULT '',
			winner VARCHAR(36),
			player_x_ready BOOLEAN DEFAULT FALSE,
			player_o_ready BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_games_status ON games(status)`,
		`CREATE INDEX IF NOT EXISTS idx_games_players ON games(player_x, player_o)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

func (db *PostgresDB) Close() error {
	return db.DB.Close()
}

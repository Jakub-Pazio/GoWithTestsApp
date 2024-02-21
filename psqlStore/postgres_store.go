package psqlstore

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/jackc/pgx/v5"
)

// TODO: Have connection injected into Store stuct, so we can test against test DB, or mock it
type PostgreSQLStore struct {
	conn pgx.Conn

	mu sync.Mutex
}

func New() (*PostgreSQLStore, error) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	store := &PostgreSQLStore{
		conn: *conn,
	}

	return store, nil
}

func (p *PostgreSQLStore) GetPlayerScore(player string) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var score int

	err := p.conn.QueryRow(context.Background(), "select score from players where name=$1", player).Scan(&score)
	if err != nil {
		return 0, err
	}

	return score, nil
}

func (p *PostgreSQLStore) RecordWin(player string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if the player exists in the database.
	var count int
	err := p.conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM players WHERE name = $1", player).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		// If the player doesn't exist, insert a new record with a score of 1.
		_, err := p.conn.Exec(context.Background(), "INSERT INTO players (name, score) VALUES ($1, $2)", player, 1)
		if err != nil {
			return err
		}
	} else {
		// If the player exists, update the score by incrementing it by 1.
		_, err := p.conn.Exec(context.Background(), "UPDATE players SET score = score + 1 WHERE name = $1", player)
		if err != nil {
			return err
		}
	}

	return nil
}

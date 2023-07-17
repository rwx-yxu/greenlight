package brokers

import (
	"context"
	"database/sql"
	"time"

	"github.com/rwx-yxu/greenlight/internal/models"
)

type token struct {
	db *sql.DB
}

type TokenWriter interface {
	Insert(token *models.Token) error
}

type TokenDeleter interface {
	DeleteAllForUser(scope string, userID int64) error
}

type TokenWriteDeleter interface {
	TokenWriter
	TokenDeleter
}

func NewToken(db *sql.DB) TokenWriteDeleter {
	return &token{db: db}
}

func (t token) Insert(token *models.Token) error {
	query := `
        INSERT INTO tokens (hash, user_id, expiry, scope)
        VALUES ($1, $2, $3, $4)`

	args := []any{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := t.db.ExecContext(ctx, query, args...)
	return err
}

func (t token) DeleteAllForUser(scope string, userID int64) error {
	query := `
        DELETE FROM tokens
        WHERE scope = $1 AND user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := t.db.ExecContext(ctx, query, scope, userID)
	return err
}

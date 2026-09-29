package bling

import (
	"context"
	"errors"
	"fmt"

	"github.com/bracomil/bracomil-internal-api-back/cmd/internal/db"
	"github.com/bracomil/bracomil-internal-api-back/cmd/tokenmanager"
	"github.com/jackc/pgx/v5"
)

var _ (tokenmanager.TokenStorage) = (*TokenRepository)(nil)

var (
	ErrNotFound = errors.New("tokenss not found")
)

type TokenRepository struct {
	db *db.DB
}

// Da pra melhorar
func NewTokenRepository(db *db.DB) *TokenRepository {
	return &TokenRepository{
		db: db,
	}
}

// Save faz upsert da linha única.
func (r *TokenRepository) Save(ctx context.Context, t tokenmanager.StorageTokens) error {
	const query = `
        INSERT INTO bling_tokens (id, access_token, refresh_token, scope, expires_in, expires_at, updated_at)
        VALUES (1, $1, $2, $3, $4, $5, now())
        ON CONFLICT (id) DO UPDATE SET
            access_token  = EXCLUDED.access_token,
            refresh_token = EXCLUDED.refresh_token,
            scope         = EXCLUDED.scope,
            expires_in    = EXCLUDED.expires_in,
            expires_at    = EXCLUDED.expires_at,
            updated_at    = now()
    `
	_, err := r.db.Exec(ctx, query,
		t.AccessToken, t.RefreshToken, t.Scope, t.ExpiresIn, t.ExpiresAt,
	)
	return err
}

// Load carrega a linha única.
func (r *TokenRepository) Load(ctx context.Context) (*tokenmanager.StorageTokens, error) {
	const query = `
        SELECT access_token, refresh_token, scope, expires_in, expires_at, created_at, updated_at
        FROM bling_tokens
        WHERE id = 1
    `
	var t tokenmanager.StorageTokens
	err := r.db.QueryRow(ctx, query).
		Scan(&t.AccessToken, &t.RefreshToken, &t.Scope, &t.ExpiresIn, &t.ExpiresAt, &t.CreatedAt, &t.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan tokens: %w", err)
	}

	return &t, nil
}

package tokenmanager

import (
	"context"
	"errors"
	"time"
)

var ErrNoStoredTokens = errors.New("no stored tokens found")

// StoredTokens é o formato do arquivo tokens.json
type StorageTokens struct {
	TokensInfoAPI
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type TokenStorage interface {
	Save(ctx context.Context, token StorageTokens) error
	Load(ctx context.Context) (*StorageTokens, error)
}

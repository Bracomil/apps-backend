package tokenmanager

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bracomil/bracomil-internal-api-back/cmd/pkg/logger"
	"golang.org/x/time/rate"
)

var (
	ErrEmptyAuthorizationToken = errors.New("could not find access tokens. Human intervention necessary")
	ErrExpiredRefreshToken     = errors.New("currently registered refresh token is expired. Human intervantion necessary")
	ErrMarshallingData         = errors.New("error while marshalling data")
	ErrUnmarshallingData       = errors.New("error while unmarshalling data")
	ErrReadingBody             = errors.New("error while reading body")
	ErrInvalidJWTParams        = errors.New("invalid jwt params provided")
	ErrEmptyTokensProvided     = errors.New("empty tokens provided by bling")
)

// type NewOAuth2Params struct {
// 	AccessToken     string
// 	RefreshToken    string
// 	ExpiresIn       int32
// 	CreatedAt       time.Time
// 	RefreshTokenTTL time.Duration
// }

// func (p NewOAuth2Params) Validate() error {
// 	if p.AccessToken == "" || p.RefreshToken == "" {
// 		return ErrInvalidJWTParams
// 	}
// 	if p.ExpiresIn <= 0 {
// 		return ErrInvalidJWTParams
// 	}
// 	if p.CreatedAt.IsZero() {
// 		return ErrInvalidJWTParams
// 	}
// 	if p.CreatedAt.Add(p.RefreshTokenTTL).Before(time.Now()) {
// 		return ErrExpiredRefreshToken
// 	}
// 	return nil
// }

// func NewOAuth2Manager(log *logger.Logger, client *http.Client, storage *TokenStorage, opts ...Option) (*OAuth2Manager, error) {
// 	stored, err := storage.Load()
// 	if err != nil && !errors.Is(err, ErrNoStoredTokens) {
// 		return nil, err
// 	}
// 	params := NewOauth2Params{
// 		AccessToken:     stored.AccessToken,
// 		RefreshToken:    stored.RefreshToken,
// 		ExpiresIn:       stored.ExpiresIn,
// 		CreatedAt:       stored.CreatedAt,
// 		RefreshTokenTTL: stored.RefreshTokenTTL,
// 	}
// 	return newJWT(log, client, storage, params, opts...)
// }

type OAuth2ManagerParams struct {
	ProviderURL     string
	RefreshTokenTTL time.Duration
}

func NewOAuth2Manager(
	log *logger.Logger,
	client *http.Client,
	storage TokenStorage,
	params OAuth2ManagerParams,
	opts ...Option,
) (*OAuth2Manager, error) {
	if client == nil {
		client = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	manager := &OAuth2Manager{
		log:        log,
		mu:         sync.Mutex{},
		httpClient: client,
		storage:    storage,

		providerUrl:     params.ProviderURL,
		refreshTokenTTL: params.RefreshTokenTTL,
	}

	for _, opt := range opts {
		opt(manager)
	}

	manager.loadTokens()

	return manager, nil
}

func (m *OAuth2Manager) loadTokens() {
	tokens, _ := m.storage.Load(context.Background())

	m.tokens.AccessToken = Token{
		Value:    tokens.AccessToken,
		ExpireAt: tokens.CreatedAt.Add(time.Duration(tokens.ExpiresIn)),
	}
	m.tokens.AccessToken = Token{
		Value:    tokens.RefreshToken,
		ExpireAt: tokens.ExpiresAt,
	}

}

func WithLimiter(l *rate.Limiter) Option {
	return func(m *OAuth2Manager) {
		m.limiter = l
	}
}

// func newJWT(
// 	log *logger.Logger,
// 	client *http.Client,
// 	params NewOauth2Params,
// 	storage *TokenStorage,
// 	opts ...Option,
// ) (*OAuth2Manager, error) {
// 	if client == nil {
// 		client = &http.Client{
// 			Timeout: 10 * time.Second,
// 		}
// 	}
// 	m := &OAuth2Manager{
// 		log:        log,
// 		mu:         sync.Mutex{},
// 		httpClient: client,
// 		storage:    storage,
// 		tokens: Tokens{
// 			AccessToken: Token{
// 				Value:    params.AccessToken,
// 				ExpireAt: params.CreatedAt.Add(time.Duration(params.ExpiresIn) * time.Second),
// 			},
// 			RefreshToken: Token{
// 				Value:    params.RefreshToken,
// 				ExpireAt: params.CreatedAt.Add(params.RefreshTokenTTL),
// 			},
// 		},
// 	}
// 	for _, opt := range opts {
// 		opt(m)
// 	}
// 	// Return new JWT
// 	return m, nil
// }

// Getter
func (m *OAuth2Manager) GetAccessToken(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.tokens.AccessToken.Value == "" {
		return "", ErrEmptyAuthorizationToken
	}
	if m.tokens.AccessToken.ExpireAt.After(time.Now()) {
		return m.tokens.AccessToken.Value, nil
	}

	m.log.Debug("Access token is expired. Getting a new one")
	return m.refresh(ctx)
}

// Refresh Tokens - assumes m.mu is locked
func (m *OAuth2Manager) refresh(ctx context.Context) (string, error) {
	// Check if refresh token is expired
	if m.tokens.RefreshToken.ExpireAt.Before(time.Now()) {
		return "", ErrExpiredRefreshToken
	}

	tokens, err := m.RequestTokensWithRefreshToken(ctx, m.tokens.RefreshToken.Value)
	if err != nil {
		return "", err
	}

	// Update values of tokens
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		return "", ErrEmptyTokensProvided
	}
	// Verify scope

	// Update values
	now := time.Now()
	m.tokens.AccessToken = Token{
		Value:    tokens.AccessToken,
		ExpireAt: now.Add(time.Duration(tokens.ExpiresIn) * time.Second),
	}
	m.tokens.RefreshToken = Token{
		Value:    tokens.RefreshToken,
		ExpireAt: now.Add(m.refreshTokenTTL),
	}

	if err := m.persist(ctx, now, tokens.ExpiresIn); err != nil {
		// Loga mas NÃO falha — a request já foi feita e o token está em memória.
		m.log.Error("failed to persist tokens:", err)
	}

	return tokens.AccessToken, nil
}

func (m *OAuth2Manager) UpdateTokens(ctx context.Context, accessToken string, refreshToken string, expiresIn int32, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// if err := params.Validate(); err != nil {
	// 	return err
	// }

	m.tokens = Tokens{
		AccessToken: Token{
			Value:    accessToken,
			ExpireAt: at.Add(time.Duration(expiresIn)),
		},
		RefreshToken: Token{
			Value:    refreshToken,
			ExpireAt: at.Add(m.refreshTokenTTL),
		},
	}

	if err := m.persist(ctx, at, expiresIn); err != nil {
		// Loga mas NÃO falha — a request já foi feita e o token está em memória.
		m.log.Error("failed to persist tokens:", err)
	}

	return nil
}

func (m *OAuth2Manager) persist(ctx context.Context, createdAt time.Time, expiresIn int32) error {
	if m.storage == nil {
		return nil
	}
	m.log.Debug("trying to update persist")
	return m.storage.Save(ctx, StorageTokens{
		TokensInfoAPI: TokensInfoAPI{
			AccessToken:  m.tokens.AccessToken.Value,
			RefreshToken: m.tokens.RefreshToken.Value,
			ExpiresIn:    expiresIn,
		},
		CreatedAt: createdAt,
		ExpiresAt: createdAt.Add(m.refreshTokenTTL),
	})
}

func (m *OAuth2Manager) ValidateState(state string) error {
	return nil
}

func (m *OAuth2Manager) RequestTokensWithCode(ctx context.Context, code string) (GetTokensResponse, error) {
	reqBody := RequestTokenReq{
		GrantType: "authorization_code",
		Code:      code,
	}
	return m.requestTokens(ctx, reqBody)
}

func (m *OAuth2Manager) RequestTokensWithRefreshToken(ctx context.Context, refreshToken string) (GetTokensResponse, error) {
	reqBody := RequestTokenReq{
		GrantType:    "refresh_token",
		RefreshToken: refreshToken,
	}
	return m.requestTokens(ctx, reqBody)
}

func (m *OAuth2Manager) requestTokens(ctx context.Context, reqBody RequestTokenReq) (GetTokensResponse, error) {
	form := url.Values{}
	form.Set("grant_type", reqBody.GrantType)
	switch reqBody.GrantType {
	case "refresh_token":
		form.Set("refresh_token", reqBody.RefreshToken)
	case "authorization_code":
		form.Set("code", reqBody.Code)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		m.providerUrl+"/oauth/token",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return GetTokensResponse{}, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "1.0")
	req.Header.Set("enable-jwt", "1")

	// base64 encoding of client_id:client_secret
	token := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", os.Getenv("BLING_CLIENT_ID"), os.Getenv("BLING_CLIENT_SECRET"))))
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", token))

	if m.limiter != nil {
		if err := m.limiter.Wait(ctx); err != nil {
			return GetTokensResponse{}, fmt.Errorf("rate limiter wait: %w", err)
		}
	}

	res, err := m.httpClient.Do(req)
	if err != nil {
		return GetTokensResponse{}, fmt.Errorf("%w, %v", err)
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return GetTokensResponse{}, fmt.Errorf("%w, %v", ErrReadingBody)
	}

	if res.StatusCode >= 400 {
		return GetTokensResponse{}, fmt.Errorf("refresh failed (status %d): %s", res.StatusCode, string(resBody))
	}

	var tokens GetTokensResponse
	if err := json.Unmarshal(resBody, &tokens); err != nil {
		return tokens, ErrUnmarshallingData
	}

	return tokens, nil
}

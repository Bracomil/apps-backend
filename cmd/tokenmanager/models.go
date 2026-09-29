package tokenmanager

import (
	"net/http"
	"sync"
	"time"

	"github.com/bracomil/bracomil-internal-api-back/cmd/pkg/logger"
	"golang.org/x/time/rate"
)

type OAuth2Manager struct {
	log        *logger.Logger
	mu         sync.Mutex
	httpClient *http.Client
	storage    TokenStorage
	limiter    *rate.Limiter

	providerUrl     string
	refreshTokenTTL time.Duration

	clientId string
	tokens   Tokens
}

func (j *OAuth2Manager) Limiter() *rate.Limiter {
	return j.limiter
}

type Tokens struct {
	AccessToken  Token
	RefreshToken Token
}

type Token struct {
	Value    string
	ExpireAt time.Time
}

type RequestTokenReq struct {
	GrantType    string `json:"grant_type"`
	Code         string `json:"code"`
	RefreshToken string `json:"refresh_token"`
}

type TokensInfoAPI struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int32  `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
}

type GetTokensResponse TokensInfoAPI

type Option func(*OAuth2Manager)

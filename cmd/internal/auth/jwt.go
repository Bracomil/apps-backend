package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/bracomil/bracomil-internal-api-back/cmd/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)

type Claims struct {
	Id                   int64    `json:"id"`
	Email                string   `json:"email"`
	Name                 string   `json:"name"`
	Permissions          []string `json:"permissions"`
	jwt.RegisteredClaims          // Embute os claims padrão (exp, iat, iss, etc.)
}

type JWTIssuer struct {
	secretName string
	issuer     string
	expiration time.Duration
}

func NewJWTIssuer(secretName string, issuer string, expiration time.Duration) *JWTIssuer {
	return &JWTIssuer{
		secretName: secretName,
		issuer:     issuer,
		expiration: expiration,
	}
}

func (j *JWTIssuer) Issue(id int64, email string, name string, permissions []string) (string, error) {
	now := time.Now()
	claims := Claims{
		Id:          id,
		Email:       email,
		Name:        name,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			ExpiresAt: jwt.NewNumericDate(now.Add(j.expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	authKey := utils.GetEnv(j.secretName, "")
	if authKey == "" {
		return "", fmt.Errorf("could not find any secret to sign")
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(authKey))
}

func (j *JWTIssuer) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method %v", ErrInvalidToken, t.Header["alg"])
		}
		return []byte(os.Getenv("AUTH_SECRET")), nil
	},
		jwt.WithIssuer(j.issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	if !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

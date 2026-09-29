package auth

import (
	"context"
	"errors"
	"fmt"

	permission "github.com/bracomil/bracomil-internal-api-back/cmd/internal/permissions"
	"github.com/bracomil/bracomil-internal-api-back/cmd/internal/user"
)

type Service struct {
	userService *user.Service
	permRepo    permission.Repository
	google      *GoogleVerifier
	jwt         *JWTIssuer
}

func NewService(ur *user.Service, pr permission.Repository, google *GoogleVerifier, jwt *JWTIssuer) *Service {
	return &Service{
		userService: ur,
		permRepo:    pr,
		google:      google,
		jwt:         jwt,
	}
}

func (s *Service) LoginWithGoogle(ctx context.Context, idToken string) (*LoginResult, error) {
	// 1. Valida o id_token do Google
	claims, err := s.google.Verify(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("verify google token: %w", err)
	}

	// 2. Busca usuário por subject
	u, err := s.userService.GetBySubject(ctx, claims.Subject)
	if err != nil {
		if !errors.Is(err, user.ErrNotFound) {
			return nil, fmt.Errorf("get user by subject: %w", err)
		}
		if u, err = s.createUserFromGoogle(ctx, claims); err != nil {
			return nil, fmt.Errorf("create user from google: %w", err)
		}
	}

	// 3. Verifica se está ativo
	if !u.Active {
		return nil, errors.New("user is inactive")
	}

	// 4. Busca permissões
	permissions, err := s.permRepo.ListByUserID(ctx, u.ID)
	if err != nil {
		return nil, fmt.Errorf("list permissions: %w", err)
	}

	// 5. Gera JWT
	token, err := s.jwt.Issue(u.ID, u.Email, u.Name, permissions)
	if err != nil {
		return nil, fmt.Errorf("issue jwt: %w", err)
	}

	return &LoginResult{
		AccessToken: token,
		Permission:  permissions,
		ExpiresIn:   int(s.jwt.expiration.Seconds()),
		User:        *u,
	}, nil
}

func (s *Service) createUserFromGoogle(ctx context.Context, claims *GoogleClaims) (*user.User, error) {
	//
	in := user.CreateInput{
		Subject: claims.Subject,
		Email:   claims.Email,
		Name:    claims.Name,
		Picture: claims.Picture,
	}
	return s.userService.Create(ctx, in)
}

// -------------------------------------- MODELOS -------------------------------------- //
type LoginResult struct {
	AccessToken string    `json:"access_token"`
	Permission  []string  `json:"permissions"`
	ExpiresIn   int       `json:"expires_in"`
	User        user.User `json:"user"`
}

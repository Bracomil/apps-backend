package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// CreateInput é o que o service recebe para criar um usuário.
type CreateInput struct {
	Subject string
	Email   string
	Name    string
	Picture string
}

// Create cria um novo usuário com validações.
func (s *Service) Create(ctx context.Context, in CreateInput) (*User, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}

	u := &User{
		Subject: strings.TrimSpace(in.Subject),
		Email:   strings.ToLower(strings.TrimSpace(in.Email)),
		Name:    strings.TrimSpace(in.Name),
		Picture: strings.TrimSpace(in.Picture),
		Active:  true,
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return u, nil
}

// GetByID retorna um usuário pelo ID.
func (s *Service) GetByID(ctx context.Context, id int64) (*User, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByEmail retorna um usuário pelo email.
func (s *Service) GetByEmail(ctx context.Context, email string) (*User, error) {
	return s.repo.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
}

// GetByEmail retorna um usuário pelo email.
func (s *Service) GetBySubject(ctx context.Context, subject string) (*User, error) {
	return s.repo.GetBySubject(ctx, strings.ToLower(strings.TrimSpace(subject)))
}

// Validate valida os campos obrigatórios.
func (in CreateInput) Validate() error {
	if strings.TrimSpace(in.Subject) == "" {
		return errors.New("subject is required")
	}
	if strings.TrimSpace(in.Email) == "" {
		return errors.New("email is required")
	}
	if !strings.Contains(in.Email, "@") {
		return errors.New("invalid email")
	}
	if strings.TrimSpace(in.Name) == "" {
		return errors.New("name is required")
	}
	return nil
}

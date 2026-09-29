package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/bracomil/bracomil-internal-api-back/cmd/internal/db"
	"github.com/jackc/pgx/v5"
)

// Erros de domínio do repositório
var (
	ErrNotFound = errors.New("user not found")
)

// Repository é a interface que o service usa.
// O service NÃO conhece pgx, só conhece estes métodos.
type Repository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetBySubject(ctx context.Context, subject string) (*User, error)
	List(ctx context.Context, limit, offset int) ([]User, error)
	Update(ctx context.Context, u *User) error
	Delete(ctx context.Context, id int64) error
}

// pgRepository é a implementação concreta usando Postgres.
type pgRepository struct {
	db *db.DB
}

// NewRepository cria um repositório de usuários.
func NewRepository(database *db.DB) Repository {
	return &pgRepository{db: database}
}

// ============================================================
// Implementação
// ============================================================

func (r *pgRepository) Create(ctx context.Context, u *User) error {
	const query = `
		INSERT INTO users (subject, email, name, active, picture)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, email, name, picture
	`

	err := r.db.QueryRow(ctx, query, u.Subject, u.Email, u.Name, u.Active, u.Picture).
		Scan(&u.ID, &u.Email, &u.Name, &u.Picture)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (r *pgRepository) GetByID(ctx context.Context, id int64) (*User, error) {
	const query = `
		SELECT id, subject, email, name, active, picture, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	return r.scanOne(ctx, query, id)
}

func (r *pgRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	const query = `
		SELECT id, subject, email, name, active, picture, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	return r.scanOne(ctx, query, email)
}

func (r *pgRepository) GetBySubject(ctx context.Context, subject string) (*User, error) {
	const query = `
		SELECT id, subject, email, name, active, picture, created_at, updated_at
		FROM users
		WHERE subject = $1
	`
	return r.scanOne(ctx, query, subject)
}

func (r *pgRepository) List(ctx context.Context, limit, offset int) ([]User, error) {
	const query = `
		SELECT id, subject, email, name, active, picture, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Subject, &u.Email, &u.Name, &u.Active, &u.Picture, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return users, nil
}

func (r *pgRepository) Update(ctx context.Context, u *User) error {
	const query = `
		UPDATE users
		SET subject = $1, email = $2, name = $3, active = $4, picture = $5, updated_at = now()
		WHERE id = $1
	`

	tag, err := r.db.Exec(ctx, query, u.Subject, u.Email, u.Name, u.Active, u.Picture, u.ID)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *pgRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM users WHERE id = $1`

	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// ============================================================
// Helpers
// ============================================================

// scanOne executa uma query que retorna uma linha e converte pra *User.
func (r *pgRepository) scanOne(ctx context.Context, query string, args ...any) (*User, error) {
	var u User
	err := r.db.QueryRow(ctx, query, args...).
		Scan(&u.ID, &u.Subject, &u.Email, &u.Name, &u.Active, &u.Picture, &u.CreatedAt, &u.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}

	return &u, nil
}

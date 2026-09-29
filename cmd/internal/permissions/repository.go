package permissions

import (
	"context"
	"errors"
	"fmt"

	"github.com/bracomil/bracomil-internal-api-back/cmd/internal/db"
	"github.com/jackc/pgx/v5"
)

type Repository interface {
	// ListByUserID retorna os nomes das permissões do usuário.
	ListByUserID(ctx context.Context, userID int64) ([]string, error)
	// Grant concede uma permissão ao usuário (idempotente).
	Grant(ctx context.Context, userID, permissionID int64) error
	// Revoke remove uma permissão do usuário.
	Revoke(ctx context.Context, userID, permissionID int64) error
	// ListAll lista todas as permissões cadastradas.
	ListAll(ctx context.Context) ([]Permission, error)
	// GetByName busca uma permissão pelo nome.
	GetByName(ctx context.Context, name string) (*Permission, error)
}

var ErrNotFound = errors.New("permission not found")

type pgRepository struct {
	db *db.DB
}

func NewRepository(database *db.DB) Repository {
	return &pgRepository{db: database}
}

func (r *pgRepository) ListByUserID(ctx context.Context, userID int64) ([]string, error) {
	const query = `
		SELECT p.name
		FROM permissions p
		INNER JOIN user_permissions up ON up.permission_id = p.id
		WHERE up.user_id = $1
		ORDER BY p.name
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list permissions by user: %w", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan permission: %w", err)
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return names, nil
}

func (r *pgRepository) Grant(ctx context.Context, userID, permissionID int64) error {
	const query = `
		INSERT INTO user_permissions (user_id, permission_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, permission_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, userID, permissionID)
	if err != nil {
		return fmt.Errorf("grant permission: %w", err)
	}
	return nil
}

func (r *pgRepository) Revoke(ctx context.Context, userID, permissionID int64) error {
	const query = `DELETE FROM user_permissions WHERE user_id = $1 AND permission_id = $2`
	_, err := r.db.Exec(ctx, query, userID, permissionID)
	if err != nil {
		return fmt.Errorf("revoke permission: %w", err)
	}
	return nil
}

func (r *pgRepository) ListAll(ctx context.Context) ([]Permission, error) {
	const query = `SELECT id, name, description FROM permissions ORDER BY name`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list permissions: %w", err)
	}
	defer rows.Close()

	var perms []Permission
	for rows.Next() {
		var p Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description); err != nil {
			return nil, fmt.Errorf("scan permission: %w", err)
		}
		perms = append(perms, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return perms, nil
}

func (r *pgRepository) GetByName(ctx context.Context, name string) (*Permission, error) {
	const query = `SELECT id, name, description FROM permissions WHERE name = $1`

	var p Permission
	err := r.db.QueryRow(ctx, query, name).Scan(&p.ID, &p.Name, &p.Description)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get permission by name: %w", err)
	}
	return &p, nil
}

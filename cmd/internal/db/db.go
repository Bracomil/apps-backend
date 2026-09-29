package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Config agrupa os parâmetros de conexão.
type Config struct {
	URL             string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
	ConnectTimeout  time.Duration
}

// DefaultConfig retorna uma configuração sensata.
func DefaultConfig(url string) Config {
	return Config{
		URL:             url,
		MaxConns:        10,
		MinConns:        2,
		MaxConnLifetime: time.Hour,
		MaxConnIdleTime: 30 * time.Minute,
		ConnectTimeout:  5 * time.Second,
	}
}

// DB é o ponto único de acesso ao banco.
// O pool é não-exportado de propósito: ninguém fora deste pacote
// consegue fazer queries diretamente sem passar pelos métodos abaixo.
type DB struct {
	pool *pgxpool.Pool
}

// New cria o pool, testa a conexão e retorna a DB.
func New(ctx context.Context, cfg Config) (*DB, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolCfg.ConnConfig.ConnectTimeout = cfg.ConnectTimeout

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &DB{pool: pool}, nil
}

// Close fecha o pool.
func (d *DB) Close() {
	if d.pool != nil {
		d.pool.Close()
	}
}

// ============================================================
// Métodos de infraestrutura (usados pelos repositórios)
// ============================================================

// Exec executa um comando que não retorna linhas (INSERT, UPDATE, DELETE).
func (d *DB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return d.pool.Exec(ctx, sql, args...)
}

// Query executa uma consulta que retorna múltiplas linhas.
func (d *DB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return d.pool.Query(ctx, sql, args...)
}

// QueryRow executa uma consulta que retorna uma única linha.
func (d *DB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return d.pool.QueryRow(ctx, sql, args...)
}

// Ping verifica a saúde da conexão (útil em health check).
func (d *DB) Ping(ctx context.Context) error {
	return d.pool.Ping(ctx)
}

// PoolStats retorna estatísticas do pool (útil para métricas).
func (d *DB) PoolStats() *pgxpool.Stat {
	return d.pool.Stat()
}

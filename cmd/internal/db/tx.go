package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// TxFn é a assinatura da função que roda dentro da transação.
// Se retornar erro, a transação sofre rollback.
type TxFn func(tx pgx.Tx) error

// WithTx executa fn dentro de uma transação.
// Commit automático se fn retornar nil; rollback caso contrário.
func (d *DB) WithTx(ctx context.Context, fn TxFn) error {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	// Rollback em caso de panic ou erro
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

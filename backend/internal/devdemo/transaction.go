package devdemo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	entsql "entgo.io/ent/dialect/sql"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/dialect"
)

type transactionRunner struct {
	client                *ent.Client
	rawDB                 *sql.DB
	postgresAdvisoryLocks bool
	mu                    sync.Mutex
}

func newTransactionRunner(client *ent.Client, rawDB *sql.DB, postgresAdvisoryLocks bool) (*transactionRunner, error) {
	if client == nil {
		return nil, errors.New("devdemo transaction runner: ent client is required")
	}
	if postgresAdvisoryLocks && rawDB == nil {
		return nil, errors.New("devdemo transaction runner: raw database is required for postgres advisory locks")
	}
	return &transactionRunner{
		client:                client,
		rawDB:                 rawDB,
		postgresAdvisoryLocks: postgresAdvisoryLocks,
	}, nil
}

func (runner *transactionRunner) run(ctx context.Context, advisoryKey int64, fn func(*ent.Client) error) error {
	if runner == nil || runner.client == nil {
		return errors.New("devdemo transaction runner: runner is required")
	}
	if ctx == nil {
		return errors.New("devdemo transaction runner: context is required")
	}
	if advisoryKey == 0 {
		return errors.New("devdemo transaction runner: advisory key is required")
	}
	if fn == nil {
		return errors.New("devdemo transaction runner: callback is required")
	}
	if runner.postgresAdvisoryLocks {
		return runner.runPostgres(ctx, advisoryKey, fn)
	}
	return runner.runFallback(ctx, fn)
}

func (runner *transactionRunner) runPostgres(ctx context.Context, advisoryKey int64, fn func(*ent.Client) error) error {
	sqlTx, err := runner.rawDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin devdemo sql transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = sqlTx.Rollback()
		}
	}()

	if _, err := sqlTx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, advisoryKey); err != nil {
		return fmt.Errorf("acquire devdemo advisory lock: %w", err)
	}
	driver := entsql.NewDriver(dialect.Postgres, entsql.Conn{ExecQuerier: sqlTx})
	txClient := ent.NewClient(ent.Driver(driver))
	if err := fn(txClient); err != nil {
		return fmt.Errorf("run devdemo transaction callback: %w", err)
	}
	if err := sqlTx.Commit(); err != nil {
		return fmt.Errorf("commit devdemo sql transaction: %w", err)
	}
	committed = true
	return nil
}

func (runner *transactionRunner) runFallback(ctx context.Context, fn func(*ent.Client) error) error {
	runner.mu.Lock()
	defer runner.mu.Unlock()

	tx, err := runner.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin devdemo ent transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if err := fn(tx.Client()); err != nil {
		return fmt.Errorf("run devdemo transaction callback: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit devdemo ent transaction: %w", err)
	}
	committed = true
	return nil
}

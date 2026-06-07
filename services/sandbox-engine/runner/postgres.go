package runner

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStatusUpdater struct {
	pool *pgxpool.Pool
}

func NewPostgresStatusUpdater(dsn string) *PostgresStatusUpdater {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		panic(fmt.Sprintf("sandbox-engine: postgres connect failed: %v", err))
	}
	return &PostgresStatusUpdater{pool: pool}
}

func (p *PostgresStatusUpdater) UpdateStatus(ctx context.Context, submissionID, status, errMsg string) error {
	_, err := p.pool.Exec(ctx,
		`UPDATE submissions SET status = $1, error_message = $2 WHERE id = $3`,
		status, errMsg, submissionID,
	)
	if err != nil {
		return fmt.Errorf("postgres: update status: %w", err)
	}
	return nil
}


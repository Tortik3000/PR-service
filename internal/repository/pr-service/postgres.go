package pr_service

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const uniqueKeyViolationCode = "23505"

type postgresRepo struct {
	db           *pgxpool.Pool
	logger       *zap.Logger
	queryBuilder sq.StatementBuilderType
}

func NewPostgresRepo(
	logger *zap.Logger,
	db *pgxpool.Pool,
) *postgresRepo {
	return &postgresRepo{
		db:           db,
		logger:       logger,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (p *postgresRepo) beginTx(
	ctx context.Context,
) (pgx.Tx, func(txErr error), error) {
	rollbackFunc := func(error) {}

	tx, err := extractTx(ctx)
	if err != nil {
		tx, err = p.db.Begin(ctx)
		if err != nil {
			return nil, nil, err
		}

		rollbackFunc = func(txErr error) {
			if txErr != nil {
				err := tx.Rollback(ctx)
				if err != nil {
					p.logger.Debug("failed to rollback transaction", zap.Error(err))
				}
				return
			}
			err := tx.Commit(ctx)
			if err != nil {
				p.logger.Debug("failed to commit transaction", zap.Error(err))
			}
		}
	}

	return tx, rollbackFunc, nil
}

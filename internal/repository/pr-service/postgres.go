package pr_service

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const uniqueKeyViolationCode = "23505"

type postgresRepo struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewPostgresRepo(
	logger *zap.Logger,
	db *pgxpool.Pool,
) *postgresRepo {
	return &postgresRepo{
		db:     db,
		logger: logger,
	}
}

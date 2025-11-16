package pr_service

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"go.uber.org/zap"

	"github.com/Tortik3000/PR-service/internal/models"
	modelsErr "github.com/Tortik3000/PR-service/internal/models/errors"
)

func (p *postgresRepo) GetReview(
	ctx context.Context,
	userID string,
) ([]models.PRShort, error) {
	logger := p.logger.With(zap.String("user_id", userID))

	getPRs := p.queryBuilder.Select(
		"pr.id",
		"pr.name",
		"pr.author_id",
		"pr.status",
	).
		From("assigned_reviewer ar").
		Join("pull_request pr ON ar.pr_id = pr.id").
		Where(sq.Eq{"ar.user_id": userID})

	getPRsStr, args, err := getPRs.ToSql()
	if err != nil {
		logger.Error("build SQL (GetReview)", zap.Error(err))
		return nil, err
	}

	logger.Debug("Executing GetReview SQL",
		zap.String("query", getPRsStr),
		zap.Any("args", args),
	)

	rows, err := p.db.Query(ctx, getPRsStr, args...)
	if err != nil {
		logger.Error("GetReview query", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var prs []models.PRShort
	for rows.Next() {
		var dbPR models.PRShort
		if err = rows.Scan(
			&dbPR.ID,
			&dbPR.Name,
			&dbPR.AuthorID,
			&dbPR.Status,
		); err != nil {
			logger.Error("scan pr row", zap.Error(err))
			return nil, err
		}
		prs = append(prs, dbPR)
	}

	return prs, nil
}

func (p *postgresRepo) SetIsActive(
	ctx context.Context,
	userID string,
	isActive bool,
) (*models.User, error) {
	logger := p.logger.With(
		zap.String("user_id", userID),
		zap.Bool("is_active", isActive),
	)

	setIsActive := p.queryBuilder.Update("users").
		Set("is_active", isActive).
		Where(sq.Eq{"id": userID}).
		Suffix("RETURNING name, team_id")

	setIsActiveStr, args, err := setIsActive.ToSql()
	if err != nil {
		logger.Error("build SQL (SetIsActive)", zap.Error(err))
		return nil, err
	}

	logger.Debug("Executing SetIsActive SQL",
		zap.String("query", setIsActiveStr),
		zap.Any("args", args),
	)

	var user models.User
	user.ID = userID
	user.IsActive = isActive

	err = p.db.QueryRow(ctx, setIsActiveStr, args...).Scan(
		&user.Name,
		&user.TeamName,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Error("SetIsActive query", zap.Error(err), zap.Error(modelsErr.ErrUserNotFound))
			return nil, modelsErr.ErrUserNotFound
		}
		logger.Error("SetIsActive query", zap.Error(err))
		return nil, err
	}

	return &user, nil
}

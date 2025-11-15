package pr_service

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/Tortik3000/PR-service/internal/models"
	modelsErr "github.com/Tortik3000/PR-service/internal/models/errors"
)

func (p *postgresRepo) GetReview(
	ctx context.Context,
	userID string,
) ([]models.PRShort, error) {
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
		return nil, err
	}

	rows, err := p.db.Query(ctx, getPRsStr, args...)
	if err != nil {
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
			return nil, err
		}
		prs = append(prs, dbPR)
	}

	return prs, nil
}

func (p *postgresRepo) SetIsActive(
	ctx context.Context,
	userId string,
	isActive bool,
) (*models.User, error) {
	setIsActive := p.queryBuilder.Update("users").
		Set("is_active", isActive).
		Where(sq.Eq{"id": userId}).
		Suffix("RETURNING name, team_id")

	setIsActiveStr, args, err := setIsActive.ToSql()
	if err != nil {
		return nil, err
	}

	var user models.User
	user.ID = userId
	user.IsActive = isActive
	err = p.db.QueryRow(ctx, setIsActiveStr, args...).Scan(
		&user.Name,
		&user.TeamName,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, modelsErr.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

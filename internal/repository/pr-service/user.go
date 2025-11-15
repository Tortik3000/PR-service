package pr_service

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	modelsErr "github.com/Tortik3000/PR-service/internal/models/errors"
	pr "github.com/Tortik3000/PR-service/internal/models/pull_request"
	"github.com/Tortik3000/PR-service/internal/models/user"
)

func (p *postgresRepo) GetReview(
	ctx context.Context,
	userID string,
) ([]pr.DBPullRequestShort, error) {
	getPRs := sq.Select("pr.id", "pr.name", "pr.author_id", "pr.status").
		From("assigned_reviewer ar").
		Join("pull_request pr ON ar.pr_id = pr.id").
		Where(sq.Eq{"ar.user_id": userID}).
		PlaceholderFormat(sq.Dollar)

	getPRsStr, args, err := getPRs.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := p.db.Query(ctx, getPRsStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []pr.DBPullRequestShort
	for rows.Next() {
		var dbPR pr.DBPullRequestShort
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
) (*user.DBUser, error) {
	setIsActive := sq.Update("users").
		Set("is_active", isActive).
		Where(sq.Eq{"id": userId}).
		Suffix("RETURNING name, team_id").
		PlaceholderFormat(sq.Dollar)

	setIsActiveStr, args, err := setIsActive.ToSql()
	if err != nil {
		return nil, err
	}

	var dbUser user.DBUser
	dbUser.ID = userId
	dbUser.IsActive = isActive
	err = p.db.QueryRow(ctx, setIsActiveStr, args...).Scan(
		&dbUser.Name,
		&dbUser.TeamName,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, modelsErr.ErrUserNotFound
		}
		return nil, err
	}

	return &dbUser, nil
}

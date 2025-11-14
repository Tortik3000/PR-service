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
	getPRs := sq.Select("pr_id").
		From("assigned_reviewer").
		Where(sq.Eq{"user_id": userID}).PlaceholderFormat(sq.Dollar)

	getPRsStr, args, err := getPRs.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := p.db.Query(ctx, getPRsStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var prID string
		err = rows.Scan(&prID)
		if err != nil {
			return nil, err
		}
		ids = append(ids, prID)
	}

	getPRs = sq.Select("id", "name", "author_id", "status").
		From("pull_request").
		Where(sq.Eq{"id": ids}).PlaceholderFormat(sq.Dollar)

	getPRsStr, args, err = getPRs.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err = p.db.Query(ctx, getPRsStr, args...)
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
	updateActive := sq.Update("users").
		Set("is_active", isActive).
		Where(sq.Eq{"id": userId}).
		Suffix("RETURNING name, team_id")

	updateActiveStr, args, err := updateActive.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, modelsErr.ErrUserNotFound
		}
		return nil, err
	}

	var dbUser user.DBUser
	dbUser.ID = userId
	dbUser.IsActive = isActive
	err = p.db.QueryRow(ctx, updateActiveStr, args...).Scan(
		&dbUser.Name,
		&dbUser.TeamName,
	)

	if err != nil {
		return nil, err
	}

	return &dbUser, nil
}

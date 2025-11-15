package pr_service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	modelsErr "github.com/Tortik3000/PR-service/internal/models/errors"
	pr "github.com/Tortik3000/PR-service/internal/models/pull_request"
	"github.com/jackc/pgx/v5/pgconn"
)

func (p *postgresRepo) PullRequestCreate(
	ctx context.Context,
	authorID, prID, prName string,
	reviewers []string,
) (dbPR *pr.DBPullRequest, txErr error) {
	tx, rollback, err := p.beginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(txErr)

	createPR := sq.Insert("pull_request").
		Columns("id", "name", "author_id").
		Values(prID, prName, authorID).
		Suffix("RETURNING created_at").PlaceholderFormat(sq.Dollar)

	createPRStr, args, err := createPR.ToSql()
	if err != nil {
		return nil, err
	}

	var createdAt *time.Time
	err = tx.QueryRow(ctx, createPRStr, args...).Scan(&createdAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueKeyViolationCode {
			return nil, modelsErr.ErrPullRequestExist
		}
		return nil, err
	}

	updateAssignedReviewers := sq.Insert("assigned_reviewer").
		Columns("user_id", "pr_id").
		PlaceholderFormat(sq.Dollar)

	for _, reviewerID := range reviewers {
		updateAssignedReviewers = updateAssignedReviewers.
			Values(reviewerID, prID)
	}

	updateAssignedReviewersStr, args, err := updateAssignedReviewers.ToSql()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, updateAssignedReviewersStr, args...)
	if err != nil {
		return nil, err
	}

	dbPR = &pr.DBPullRequest{
		ID:                prID,
		Name:              prName,
		AuthorID:          authorID,
		AssignedReviewers: reviewers,
		CreatedAt:         createdAt,
		Status:            pr.OPEN,
	}

	return dbPR, nil
}

func (p *postgresRepo) PullRequestMerge(
	ctx context.Context,
	prID string,
) (*pr.DBPullRequest, error) {
	updateStatus := sq.Update("pull_request").
		Set("status", 1).
		SetMap(map[string]interface{}{
			"merged_at": sq.Expr("COALESCE(merged_at, ?)", time.Now())}).
		Where(sq.Eq{"id": prID}).
		Suffix("RETURNING id, name, author_id, created_at, merged_at, status").
		PlaceholderFormat(sq.Dollar)

	updateStatusStr, args, err := updateStatus.ToSql()
	if err != nil {
		return nil, err
	}

	var dbPr pr.DBPullRequest
	err = p.db.QueryRow(ctx, updateStatusStr, args...).Scan(
		&dbPr.ID,
		&dbPr.Name,
		&dbPr.AuthorID,
		&dbPr.CreatedAt,
		&dbPr.MergedAt,
		&dbPr.Status,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, modelsErr.ErrPRNotFound
		}
		return nil, err
	}

	return &dbPr, nil
}

func (p *postgresRepo) GetPullRequest(
	ctx context.Context,
	prID string,
) (dbPR *pr.DBPullRequest, txErr error) {
	tx, rollback, err := p.beginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(txErr)
	getPR := sq.Select("id", "name", "author_id", "created_at", "merged_at", "status").
		From("pull_request").
		Where(sq.Eq{"id": prID}).PlaceholderFormat(sq.Dollar)

	getPRSql, args, err := getPR.ToSql()
	if err != nil {
		return nil, err
	}

	dbPR = &pr.DBPullRequest{}
	err = tx.QueryRow(ctx, getPRSql, args...).Scan(
		&dbPR.ID,
		&dbPR.Name,
		&dbPR.AuthorID,
		&dbPR.CreatedAt,
		&dbPR.MergedAt,
		&dbPR.Status,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, modelsErr.ErrPRNotFound
		}
		return nil, err
	}

	if dbPR.Status == pr.MERGED {
		return nil, modelsErr.ErrPRMerged
	}

	getReviewers := sq.Select("user_id").
		From("assigned_reviewer").
		Where(sq.Eq{"pr_id": prID}).
		PlaceholderFormat(sq.Dollar)

	getReviewersStr, args, err := getReviewers.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, getReviewersStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var reviewer string
		err = rows.Scan(&reviewer)
		if err != nil {
			return nil, err
		}
		dbPR.AssignedReviewers = append(dbPR.AssignedReviewers, reviewer)
	}

	return dbPR, nil
}
func (p *postgresRepo) PullRequestReassign(
	ctx context.Context,
	prID, oldReviewerID, newReviewerID string,
) (txErr error) {
	tx, rollback, err := p.beginTx(ctx)
	if err != nil {
		return err
	}
	defer rollback(txErr)

	updateReviewers := sq.Update("assigned_reviewer").
		Set("user_id", newReviewerID).
		Where(sq.And{
			sq.Eq{"user_id": oldReviewerID},
			sq.Eq{"pr_id": prID},
		}).
		PlaceholderFormat(sq.Dollar)

	updateReviewersStr, args, err := updateReviewers.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, updateReviewersStr, args...)
	if err != nil {
		return err
	}

	return nil
}

func (p *postgresRepo) GetTeamIDByUserID(
	ctx context.Context,
	userID string,
) (teamID string, err error) {
	getTeamID := sq.Select("team_id").
		From("users").
		Where(sq.Eq{"id": userID}).PlaceholderFormat(sq.Dollar)

	getTeamIDStr, args, err := getTeamID.ToSql()
	if err != nil {
		return "", err
	}

	err = p.db.QueryRow(ctx, getTeamIDStr, args...).Scan(&teamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", modelsErr.ErrUserNotFound
		}
		return "", err
	}
	if teamID == "" {
		return "", modelsErr.ErrTeamNotFound
	}

	return teamID, nil
}

func (p *postgresRepo) GetActiveTeammates(
	ctx context.Context,
	teamID string,
	excludedUsers []string,
	count uint64,
) (teammateIDs []string, txErr error) {
	tx, rollback, err := p.beginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(txErr)

	getTeammate := sq.Select("id").
		From("users").
		Where(
			sq.And{
				sq.Eq{"team_id": teamID},
				sq.Eq{"is_active": true},
				sq.NotEq{"id": excludedUsers},
			},
		).
		Limit(count).
		Suffix("FOR UPDATE").PlaceholderFormat(sq.Dollar)
	
	getTeammateStr, args, err := getTeammate.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, getTeammateStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		if err = rows.Scan(&userID); err != nil {
			return nil, err
		}
		teammateIDs = append(teammateIDs, userID)
	}

	return teammateIDs, nil
}

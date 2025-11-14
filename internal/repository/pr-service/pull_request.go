package pr_service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	modelsErr "github.com/Tortik3000/PR-service/internal/models/errors"
	pr "github.com/Tortik3000/PR-service/internal/models/pull_request"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (p *postgresRepo) PullRequestCreate(
	ctx context.Context,
	authorID, prID, prName string,
) (*pr.DBPullRequest, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	createPR := sq.Insert("pull_request").
		Columns("id", "name", "author_id").
		Values(prID, prName, authorID).
		Suffix("RETURNING created_at").PlaceholderFormat(sq.Dollar)

	createPRStr, args, err := createPR.ToSql()
	if err != nil {
		return nil, err
	}

	var createdAt *time.Time
	err = tx.QueryRow(ctx, createPRStr, args...).Scan(
		&createdAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueKeyViolationCode {
			return nil, modelsErr.ErrPullRequestExist
		}
		return nil, err
	}

	teamId, err := getTeamIDByUserID(ctx, tx, authorID)
	if err != nil {
		return nil, err
	}

	reviewersIDs, err := getTeammates(ctx, tx, teamId, authorID, 2)
	if err != nil {
		return nil, err
	}

	assignedReviewerInsert := sq.Insert("assigned_reviewer").
		Columns("user_id", "pr_id")

	for _, reviewerID := range reviewersIDs {
		assignedReviewerInsert = assignedReviewerInsert.
			Values(reviewerID, prID)
	}
	assignedReviewerInsert.PlaceholderFormat(sq.Dollar)

	assignedReviewerInsertStr, args, err := assignedReviewerInsert.ToSql()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, assignedReviewerInsertStr, args...)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	dbPR := &pr.DBPullRequest{
		ID:                prID,
		Name:              prName,
		AuthorID:          authorID,
		AssignedReviewers: reviewersIDs,
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
		Suffix("RETURNING name, author_id, created_at, merged_at").
		PlaceholderFormat(sq.Dollar)

	updateStatusStr, args, err := updateStatus.ToSql()
	if err != nil {
		return nil, err
	}

	var dbPr pr.DBPullRequest
	dbPr.ID = prID
	dbPr.Status = pr.MERGED
	err = p.db.QueryRow(ctx, updateStatusStr, args...).Scan(
		&dbPr.Name,
		&dbPr.AuthorID,
		&dbPr.CreatedAt,
		&dbPr.MergedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, modelsErr.ErrPRNotFound
		}
		return nil, err
	}

	return &dbPr, nil
}

func (p *postgresRepo) PullRequestReassign(
	ctx context.Context,
	prID, oldUserID string,
) (*pr.DBPullRequest, string, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return nil, "", err
	}
	defer tx.Rollback(ctx)

	teamId, err := getTeamIDByUserID(ctx, tx, oldUserID)
	if err != nil {
		return nil, "", err
	}

	reviewersIDs, err := getTeammates(ctx, tx, teamId, oldUserID, 1)
	if err != nil {
		return nil, "", err
	}
	if len(reviewersIDs) == 0 {
		return nil, "", modelsErr.ErrNotCandidate
	}
	newUserID := reviewersIDs[0]

	update := sq.Update("assigned_reviewer").
		Set("user_id", newUserID).
		Where(sq.Eq{"pr_id": prID}, sq.Eq{"user_id": oldUserID}).PlaceholderFormat(sq.Dollar)

	updateStr, args, err := update.ToSql()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", modelsErr.ErrNotAssigned
		}
		return nil, "", err
	}

	_, err = tx.Exec(ctx, updateStr, args...)
	if err != nil {
		return nil, "", err
	}

	var dbPR pr.DBPullRequest
	getPR := sq.Select("id", "name", "author_id", "created_at", "merged_at", "status").
		From("pull_request").
		Where(sq.Eq{"id": prID}).PlaceholderFormat(sq.Dollar)

	getPRSql, args, err := getPR.ToSql()
	if err != nil {
		return nil, "", err
	}
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
			return nil, "", modelsErr.ErrPRNotFound
		}
		return nil, "", err
	}
	if dbPR.Status == pr.MERGED {
		return nil, "", modelsErr.ErrPRMerged
	}

	getReview := sq.Select("user_id").
		From("assigned_reviewer").
		Where(sq.Eq{"pr_id": prID}).PlaceholderFormat(sq.Dollar)

	getReviewStr, args, err := getReview.ToSql()
	if err != nil {
		return nil, "", err
	}

	rows, err := tx.Query(ctx, getReviewStr, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		if err = rows.Scan(&userID); err != nil {
			return nil, "", err
		}
		dbPR.AssignedReviewers = append(dbPR.AssignedReviewers, userID)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, "", err
	}

	return &dbPR, newUserID, nil
}

func getTeamIDByUserID(
	ctx context.Context,
	tx pgx.Tx,
	userID string,
) (teamID string, err error) {
	getTeamID := sq.Select("team_id").
		From("users").
		Where(sq.Eq{"id": userID}).PlaceholderFormat(sq.Dollar)

	getTeamIDStr, args, err := getTeamID.ToSql()
	if err != nil {
		return "", err
	}

	err = tx.QueryRow(ctx, getTeamIDStr, args...).Scan(&teamID)
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

func getTeammates(
	ctx context.Context,
	tx pgx.Tx,
	teamID, authorID string,
	count uint64,
) (teammateIDs []string, err error) {
	getTeammate := sq.Select("id").
		From("users").
		Where(sq.Eq{"team_id": teamID, "is_active": true}, sq.NotEq{"id": authorID}).
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

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
	"go.uber.org/zap"
)

func (p *postgresRepo) PullRequestCreate(
	ctx context.Context,
	authorID, prID, prName string,
) (*pr.DBPullRequest, error) {
	logger := p.logger.With(
		zap.String("method", "PullRequestCreate"),
		zap.String("pr_id", prID),
		zap.String("author_id", authorID),
	)

	logger.Info("creating pull request")

	tx, err := p.db.Begin(ctx)
	if err != nil {
		logger.Error("failed to begin transaction", zap.Error(err))
		return nil, err
	}
	defer tx.Rollback(ctx)

	teamID, err := getTeamIDByUserID(ctx, tx, authorID, logger)
	if err != nil {
		return nil, err
	}

	createPR := sq.Insert("pull_request").
		Columns("id", "name", "author_id").
		Values(prID, prName, authorID).
		Suffix("RETURNING created_at").PlaceholderFormat(sq.Dollar)

	createPRStr, args, err := createPR.ToSql()
	if err != nil {
		logger.Error("failed to build SQL query", zap.Error(err))
		return nil, err
	}

	var createdAt *time.Time
	err = tx.QueryRow(ctx, createPRStr, args...).Scan(&createdAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueKeyViolationCode {
			logger.Warn("pull request already exists", zap.String("pr_id", prID))
			return nil, modelsErr.ErrPullRequestExist
		}
		logger.Error("failed to create pull request", zap.Error(err))
		return nil, err
	}

	authorIDs := []string{authorID}
	reviewersIDs, err := getTeammates(ctx, tx, teamID, authorIDs, 2, logger)
	if err != nil {
		return nil, err
	}

	logger.Info("assigning reviewers", zap.Strings("reviewer_ids", reviewersIDs))

	assignedReviewerInsert := sq.Insert("assigned_reviewer").
		Columns("user_id", "pr_id").
		PlaceholderFormat(sq.Dollar)

	for _, reviewerID := range reviewersIDs {
		assignedReviewerInsert = assignedReviewerInsert.
			Values(reviewerID, prID)
	}

	assignedReviewerInsertStr, args, err := assignedReviewerInsert.ToSql()
	if err != nil {
		logger.Error("failed to build assigned reviewer SQL", zap.Error(err))
		return nil, err
	}

	_, err = tx.Exec(ctx, assignedReviewerInsertStr, args...)
	if err != nil {
		logger.Error("failed to assign reviewers", zap.Error(err))
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		logger.Error("failed to commit transaction", zap.Error(err))
		return nil, err
	}

	logger.Info("pull request created successfully")

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
	logger := p.logger.With(
		zap.String("method", "PullRequestMerge"),
		zap.String("pr_id", prID),
	)

	logger.Info("merging pull request")

	updateStatus := sq.Update("pull_request").
		Set("status", 1).
		SetMap(map[string]interface{}{
			"merged_at": sq.Expr("COALESCE(merged_at, ?)", time.Now())}).
		Where(sq.Eq{"id": prID}).
		Suffix("RETURNING name, author_id, created_at, merged_at").
		PlaceholderFormat(sq.Dollar)

	updateStatusStr, args, err := updateStatus.ToSql()
	if err != nil {
		logger.Error("failed to build merge SQL", zap.Error(err))
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
			logger.Warn("pull request not found for merge")
			return nil, modelsErr.ErrPRNotFound
		}
		logger.Error("failed to merge pull request", zap.Error(err))
		return nil, err
	}

	logger.Info("pull request merged successfully")
	return &dbPr, nil
}

func (p *postgresRepo) PullRequestReassign(
	ctx context.Context,
	prID, oldUserID string,
) (*pr.DBPullRequest, string, error) {
	logger := p.logger.With(
		zap.String("method", "PullRequestReassign"),
		zap.String("pr_id", prID),
		zap.String("old_user_id", oldUserID),
	)

	logger.Info("reassigning pull request reviewer")

	tx, err := p.db.Begin(ctx)
	if err != nil {
		logger.Error("failed to begin transaction", zap.Error(err))
		return nil, "", err
	}
	defer tx.Rollback(ctx)

	var dbPR pr.DBPullRequest
	getPR := sq.Select("id", "name", "author_id", "created_at", "merged_at", "status").
		From("pull_request").
		Where(sq.Eq{"id": prID}).PlaceholderFormat(sq.Dollar)

	getPRSql, args, err := getPR.ToSql()
	if err != nil {
		logger.Error("failed to build get PR SQL", zap.Error(err))
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
			logger.Warn("pull request not found")
			return nil, "", modelsErr.ErrPRNotFound
		}
		logger.Error("failed to get pull request", zap.Error(err))
		return nil, "", err
	}

	if dbPR.Status == pr.MERGED {
		logger.Warn("attempt to reassign merged PR")
		return nil, "", modelsErr.ErrPRMerged
	}

	getReviewers := sq.Select("user_id").
		From("assigned_reviewer").
		Where(sq.Eq{"pr_id": prID}).
		PlaceholderFormat(sq.Dollar)

	getReviewersStr, args, err := getReviewers.ToSql()
	if err != nil {
		logger.Error("failed to build get PR SQL", zap.Error(err))
		return nil, "", err
	}

	exceptReviewers := make([]string, 0, 2)
	rows, err := tx.Query(ctx, getReviewersStr, args...)
	if err != nil {
		logger.Error("failed to get old reviewer", zap.Error(err))
		return nil, "", err
	}
	defer rows.Close()
	for rows.Next() {
		var reviewer string
		err = rows.Scan(&reviewer)
		if err != nil {
			return nil, "", err
		}
		exceptReviewers = append(exceptReviewers, reviewer)
	}

	isWasReviewer := false
	for _, reviewer := range exceptReviewers {
		if oldUserID == reviewer {
			isWasReviewer = true
		}
	}
	if !isWasReviewer {
		logger.Warn("reviewer not assigned to this PR")
		return nil, "", modelsErr.ErrNotAssigned
	}

	teamID, err := getTeamIDByUserID(ctx, tx, oldUserID, logger)
	if err != nil {
		return nil, "", err
	}

	exceptReviewers = append(exceptReviewers, dbPR.AuthorID)
	reviewersIDs, err := getTeammates(ctx, tx, teamID, exceptReviewers, 1, logger)
	if err != nil {
		return nil, "", err
	}
	if len(reviewersIDs) == 0 {
		logger.Warn("no candidate found for reassignment")
		return nil, "", modelsErr.ErrNotCandidate
	}
	newUserID := reviewersIDs[0]

	logger.Info("found replacement reviewer",
		zap.String("new_user_id", newUserID))

	update := sq.Update("assigned_reviewer").
		Set("user_id", newUserID).
		Where(sq.Eq{"pr_id": prID}, sq.Eq{"user_id": oldUserID}).
		PlaceholderFormat(sq.Dollar)

	updateStr, args, err := update.ToSql()
	if err != nil {
		logger.Error("failed to build reassign SQL", zap.Error(err))
		return nil, "", err
	}

	result, err := tx.Exec(ctx, updateStr, args...)
	if err != nil {
		logger.Error("failed to execute reassign", zap.Error(err))
		return nil, "", err
	}

	if result.RowsAffected() == 0 {
		logger.Warn("reviewer not assigned to this PR")
		return nil, "", modelsErr.ErrNotAssigned
	}

	getReview := sq.Select("user_id").
		From("assigned_reviewer").
		Where(sq.Eq{"pr_id": prID}).PlaceholderFormat(sq.Dollar)

	getReviewStr, args, err := getReview.ToSql()
	if err != nil {
		logger.Error("failed to build get reviewers SQL", zap.Error(err))
		return nil, "", err
	}

	rows, err = tx.Query(ctx, getReviewStr, args...)
	if err != nil {
		logger.Error("failed to query reviewers", zap.Error(err))
		return nil, "", err
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		if err = rows.Scan(&userID); err != nil {
			logger.Error("failed to scan reviewer", zap.Error(err))
			return nil, "", err
		}
		dbPR.AssignedReviewers = append(dbPR.AssignedReviewers, userID)
	}

	if err = tx.Commit(ctx); err != nil {
		logger.Error("failed to commit transaction", zap.Error(err))
		return nil, "", err
	}

	logger.Info("reviewer reassigned successfully")
	return &dbPR, newUserID, nil
}

func getTeamIDByUserID(
	ctx context.Context,
	tx pgx.Tx,
	userID string,
	logger *zap.Logger,
) (teamID string, err error) {
	getTeamID := sq.Select("team_id").
		From("users").
		Where(sq.Eq{"id": userID}).PlaceholderFormat(sq.Dollar)

	getTeamIDStr, args, err := getTeamID.ToSql()
	if err != nil {
		logger.Error("failed to build get team ID SQL", zap.Error(err))
		return "", err
	}

	err = tx.QueryRow(ctx, getTeamIDStr, args...).Scan(&teamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("user not found", zap.String("user_id", userID))
			return "", modelsErr.ErrUserNotFound
		}
		logger.Error("failed to get team ID", zap.Error(err))
		return "", err
	}
	if teamID == "" {
		logger.Warn("team not found for user", zap.String("user_id", userID))
		return "", modelsErr.ErrTeamNotFound
	}

	return teamID, nil
}

func getTeammates(
	ctx context.Context,
	tx pgx.Tx,
	teamID string,
	exceptUsers []string,
	count uint64,
	logger *zap.Logger,
) (teammateIDs []string, err error) {
	getTeammate := sq.Select("id").
		From("users").
		Where(
			sq.And{
				sq.Eq{"team_id": teamID, "is_active": true},
				sq.NotEq{"id": exceptUsers},
			},
		).
		Limit(count).
		Suffix("FOR UPDATE").PlaceholderFormat(sq.Dollar)

	getTeammateStr, args, err := getTeammate.ToSql()
	if err != nil {
		logger.Error("failed to build get teammates SQL", zap.Error(err))
		return nil, err
	}

	rows, err := tx.Query(ctx, getTeammateStr, args...)
	if err != nil {
		logger.Error("failed to query teammates", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		if err = rows.Scan(&userID); err != nil {
			logger.Error("failed to scan teammate", zap.Error(err))
			return nil, err
		}
		teammateIDs = append(teammateIDs, userID)
	}

	logger.Debug("found teammates",
		zap.String("team_id", teamID),
		zap.Int("count", len(teammateIDs)),
		zap.Strings("teammate_ids", teammateIDs))

	return teammateIDs, nil
}

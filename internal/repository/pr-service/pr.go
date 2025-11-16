package pr_service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"

	"github.com/Tortik3000/PR-service/internal/models"
	modelsErr "github.com/Tortik3000/PR-service/internal/models/errors"
)

func (p *postgresRepo) PullRequestCreate(
	ctx context.Context,
	authorID, prID, prName string,
	reviewers []string,
) (pr *models.PR, txErr error) {
	logger := p.logger.With(
		zap.String("author_id", authorID),
		zap.String("pr_id", prID),
		zap.String("pr_name", prName),
		zap.Any("reviewers", reviewers),
	)

	tx, rollback, err := p.beginTx(ctx)
	if err != nil {
		logger.Error("beginTx", zap.Error(err))
		return nil, err
	}
	defer rollback(txErr)

	createPR := p.queryBuilder.Insert("pull_request").
		Columns("id", "name", "author_id").
		Values(prID, prName, authorID).
		Suffix("RETURNING created_at")

	createPRStr, args, err := createPR.ToSql()
	if err != nil {
		logger.Error("build SQL (create PR)", zap.Error(err))
		return nil, err
	}

	logger.Debug("Executing create PR SQL",
		zap.String("query", createPRStr),
		zap.Any("args", args),
	)

	var createdAt *time.Time
	err = tx.QueryRow(ctx, createPRStr, args...).Scan(&createdAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueKeyViolationCode {
			logger.Error("create PR query", zap.Error(modelsErr.ErrPullRequestExist))
			return nil, modelsErr.ErrPullRequestExist
		}
		logger.Error("create PR query", zap.Error(err))
		return nil, err
	}
	if len(reviewers) > 0 {
		updateAssignedReviewers := p.queryBuilder.Insert("assigned_reviewer").
			Columns("user_id", "pr_id")

		for _, reviewerID := range reviewers {
			updateAssignedReviewers = updateAssignedReviewers.
				Values(reviewerID, prID)
		}

		updateAssignedReviewersStr, args, err := updateAssignedReviewers.ToSql()
		if err != nil {
			logger.Error("build SQL (insert reviewers)", zap.Error(err))
			return nil, err
		}

		logger.Debug("Executing insert reviewers SQL",
			zap.String("query", updateAssignedReviewersStr),
			zap.Any("args", args),
		)

		_, err = tx.Exec(ctx, updateAssignedReviewersStr, args...)
		if err != nil {
			logger.Error("insert reviewers", zap.Error(err))
			return nil, err
		}
	}

	pr = &models.PR{
		ID:                prID,
		Name:              prName,
		AuthorID:          authorID,
		AssignedReviewers: reviewers,
		CreatedAt:         createdAt,
		Status:            models.PRStatusOPEN,
	}

	return pr, nil
}

func (p *postgresRepo) PullRequestMerge(
	ctx context.Context,
	prID string,
) (*models.PR, error) {
	logger := p.logger.With(zap.String("pr_id", prID))

	updateStatus := p.queryBuilder.Update("pull_request").
		Set("status", models.PRStatusMERGED).
		SetMap(map[string]interface{}{
			"merged_at": sq.Expr("COALESCE(merged_at, ?)", time.Now())}).
		Where(sq.Eq{"id": prID}).
		Suffix("RETURNING id, name, author_id, created_at, merged_at, status")

	updateStatusStr, args, err := updateStatus.ToSql()
	if err != nil {
		logger.Error("build SQL (merge)", zap.Error(err))
		return nil, err
	}

	logger.Debug("Executing merge SQL",
		zap.String("query", updateStatusStr),
		zap.Any("args", args),
	)

	var dbPr models.PR
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
			logger.Error("merge query", zap.Error(modelsErr.ErrPRNotFound))
			return nil, modelsErr.ErrPRNotFound
		}
		logger.Error("merge query", zap.Error(err))
		return nil, err
	}

	return &dbPr, nil
}

func (p *postgresRepo) GetPullRequest(
	ctx context.Context,
	prID string,
) (pr *models.PR, txErr error) {
	logger := p.logger.With(zap.String("pr_id", prID))

	tx, rollback, err := p.beginTx(ctx)
	if err != nil {
		logger.Error("beginTx", zap.Error(err))
		return nil, err
	}
	defer rollback(txErr)

	getPR := p.queryBuilder.Select(
		"id",
		"name",
		"author_id",
		"created_at",
		"merged_at",
		"status",
	).
		From("pull_request").
		Where(sq.Eq{"id": prID}).
		Suffix("FOR UPDATE")

	getPRSql, args, err := getPR.ToSql()
	if err != nil {
		logger.Error("build SQL (get PR)", zap.Error(err))
		return nil, err
	}

	logger.Debug("Executing get PR SQL",
		zap.String("query", getPRSql),
		zap.Any("args", args),
	)

	pr = &models.PR{}
	err = tx.QueryRow(ctx, getPRSql, args...).Scan(
		&pr.ID,
		&pr.Name,
		&pr.AuthorID,
		&pr.CreatedAt,
		&pr.MergedAt,
		&pr.Status,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Error("get PR query", zap.Error(modelsErr.ErrPRNotFound))
			return nil, modelsErr.ErrPRNotFound
		}
		logger.Error("get PR query", zap.Error(err))
		return nil, err
	}

	if pr.Status == models.PRStatusMERGED {
		logger.Error("PR is already merged", zap.Error(modelsErr.ErrPRMerged))
		return nil, modelsErr.ErrPRMerged
	}

	getReviewers := p.queryBuilder.Select("user_id").
		From("assigned_reviewer").
		Where(sq.Eq{"pr_id": prID}).
		Suffix("FOR UPDATE")

	getReviewersStr, args, err := getReviewers.ToSql()
	if err != nil {
		logger.Error("build SQL (get reviewers)", zap.Error(err))
		return nil, err
	}

	p.logger.Debug("Executing get reviewers SQL",
		zap.String("query", getReviewersStr),
		zap.Any("args", args),
	)

	rows, err := tx.Query(ctx, getReviewersStr, args...)
	if err != nil {
		logger.Error("get reviewers", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var reviewer string
		err = rows.Scan(&reviewer)
		if err != nil {
			logger.Error("scan reviewer", zap.Error(err))
			return nil, err
		}
		pr.AssignedReviewers = append(pr.AssignedReviewers, reviewer)
	}

	return pr, nil
}

func (p *postgresRepo) PullRequestReassign(
	ctx context.Context,
	prID, oldReviewerID, newReviewerID string,
) (txErr error) {
	logger := p.logger.With(
		zap.String("pr_id", prID),
		zap.String("old_reviewer_id", oldReviewerID),
		zap.String("new_reviewer_id", newReviewerID),
	)

	tx, rollback, err := p.beginTx(ctx)
	if err != nil {
		logger.Error("beginTx", zap.Error(err))
		return err
	}
	defer rollback(txErr)

	updateReviewers := p.queryBuilder.Update("assigned_reviewer").
		Set("user_id", newReviewerID).
		Where(sq.And{
			sq.Eq{"user_id": oldReviewerID},
			sq.Eq{"pr_id": prID},
		})

	updateReviewersStr, args, err := updateReviewers.ToSql()
	if err != nil {
		logger.Error("build SQL (reassign reviewer)", zap.Error(err))
		return err
	}

	logger.Debug("Executing reassign SQL",
		zap.String("query", updateReviewersStr),
		zap.Any("args", args),
	)

	_, err = tx.Exec(ctx, updateReviewersStr, args...)
	if err != nil {
		p.logger.Error("reassign reviewer", zap.Error(err))
		return err
	}

	return nil
}

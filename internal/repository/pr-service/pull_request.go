package pr_service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/Tortik3000/PR-service/internal/models"
	modelsErr "github.com/Tortik3000/PR-service/internal/models/errors"
	"github.com/jackc/pgx/v5/pgconn"
)

func (p *postgresRepo) PullRequestCreate(
	ctx context.Context,
	authorID, prID, prName string,
	reviewers []string,
) (pr *models.PR, txErr error) {
	tx, rollback, err := p.beginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(txErr)

	createPR := p.queryBuilder.Insert("pull_request").
		Columns("id", "name", "author_id").
		Values(prID, prName, authorID).
		Suffix("RETURNING created_at")

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

	updateAssignedReviewers := p.queryBuilder.Insert("assigned_reviewer").
		Columns("user_id", "pr_id")

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
	updateStatus := p.queryBuilder.Update("pull_request").
		Set("status", models.PRStatusMERGED).
		SetMap(map[string]interface{}{
			"merged_at": sq.Expr("COALESCE(merged_at, ?)", time.Now())}).
		Where(sq.Eq{"id": prID}).
		Suffix("RETURNING id, name, author_id, created_at, merged_at, status")

	updateStatusStr, args, err := updateStatus.ToSql()
	if err != nil {
		return nil, err
	}

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
			return nil, modelsErr.ErrPRNotFound
		}
		return nil, err
	}

	return &dbPr, nil
}

func (p *postgresRepo) GetPullRequest(
	ctx context.Context,
	prID string,
) (pr *models.PR, txErr error) {
	tx, rollback, err := p.beginTx(ctx)
	if err != nil {
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
		Where(sq.Eq{"id": prID})

	getPRSql, args, err := getPR.ToSql()
	if err != nil {
		return nil, err
	}

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
			return nil, modelsErr.ErrPRNotFound
		}
		return nil, err
	}

	if pr.Status == models.PRStatusMERGED {
		return nil, modelsErr.ErrPRMerged
	}

	getReviewers := p.queryBuilder.Select("user_id").
		From("assigned_reviewer").
		Where(sq.Eq{"pr_id": prID})

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
		pr.AssignedReviewers = append(pr.AssignedReviewers, reviewer)
	}

	return pr, nil
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

	updateReviewers := p.queryBuilder.Update("assigned_reviewer").
		Set("user_id", newReviewerID).
		Where(sq.And{
			sq.Eq{"user_id": oldReviewerID},
			sq.Eq{"pr_id": prID},
		})

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

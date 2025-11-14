package pr_service

import (
	"context"

	models "github.com/Tortik3000/PR-service/internal/models/pull_request"
)

func (u *useCase) PullRequestCreate(
	ctx context.Context,
	authorID, prID, prName string,
) (*models.PR, error) {
	dbPR, err := u.pullRequestsRepository.PullRequestCreate(ctx, authorID, prID, prName)
	if err != nil {
		return nil, err
	}

	return models.FromDB(dbPR), nil
}

func (u *useCase) PullRequestMerge(
	ctx context.Context,
	prID string,
) (*models.PR, error) {
	dbPR, err := u.pullRequestsRepository.PullRequestMerge(ctx, prID)
	if err != nil {
		return nil, err
	}

	return models.FromDB(dbPR), nil
}

func (u *useCase) PullRequestReassign(
	ctx context.Context,
	prID, oldUserID string,
) (*models.PR, string, error) {
	dbPR, replacedBy, err := u.pullRequestsRepository.PullRequestReassign(ctx, prID, oldUserID)
	if err != nil {
		return nil, "", err
	}

	return models.FromDB(dbPR), replacedBy, nil
}

package pr_service

import (
	"context"

	modelsErr "github.com/Tortik3000/PR-service/internal/models/errors"
	models "github.com/Tortik3000/PR-service/internal/models/pull_request"
)

func (u *useCase) PullRequestCreate(
	ctx context.Context,
	authorID, prID, prName string,
) (*models.PR, error) {
	var dbPR *models.DBPullRequest
	err := u.transactor.WithTx(ctx, func(ctx context.Context) error {
		teamID, err := u.pullRequestsRepository.GetTeamIDByUserID(ctx, authorID)
		if err != nil {
			return err
		}

		excludedUsers := []string{authorID}
		teammates, err := u.pullRequestsRepository.GetActiveTeammates(ctx, teamID, excludedUsers, 2)
		if err != nil {
			return err
		}

		dbPR, err = u.pullRequestsRepository.PullRequestCreate(ctx, authorID, prID, prName, teammates)
		if err != nil {
			return err
		}

		return nil
	})

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
	prID, oldReviewerID string,
) (*models.PR, string, error) {
	var dbPR *models.DBPullRequest
	var newReviewerID string

	err := u.transactor.WithTx(ctx, func(ctx context.Context) error {
		teamID, err := u.pullRequestsRepository.GetTeamIDByUserID(ctx, oldReviewerID)
		if err != nil {
			return err
		}

		dbPR, err = u.pullRequestsRepository.GetPullRequest(ctx, prID)
		if err != nil {
			return err
		}
		wasReviewer := false
		for _, reviewerID := range dbPR.AssignedReviewers {
			if reviewerID == oldReviewerID {
				wasReviewer = true
			}
		}
		if !wasReviewer {
			return modelsErr.ErrNotAssigned
		}
		excludedUsers := []string{dbPR.AuthorID}
		excludedUsers = append(excludedUsers, dbPR.AssignedReviewers...)

		teammates, err := u.pullRequestsRepository.GetActiveTeammates(ctx, teamID, excludedUsers, 1)
		if err != nil {
			return err
		}

		if len(teammates) == 0 {
			return modelsErr.ErrNotActiveCandidate
		}

		newReviewerID = teammates[0]
		err = u.pullRequestsRepository.PullRequestReassign(ctx, prID, oldReviewerID, newReviewerID)
		if err != nil {
			return err
		}

		newReviewers := []string{newReviewerID}
		for _, reviewer := range dbPR.AssignedReviewers {
			if reviewer != oldReviewerID {
				newReviewers = append(newReviewers, reviewer)
			}
		}

		dbPR.AssignedReviewers = newReviewers

		return nil
	})

	if err != nil {
		return nil, "", err
	}

	return models.FromDB(dbPR), newReviewerID, nil
}

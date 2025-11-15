package pr_service

import (
	"context"

	"github.com/Tortik3000/PR-service/internal/models"
	modelsErr "github.com/Tortik3000/PR-service/internal/models/errors"
)

func (u *useCase) PullRequestCreate(
	ctx context.Context,
	authorID, prID, prName string,
) (*models.PR, error) {
	var pr *models.PR
	err := u.transactor.WithTx(ctx, func(ctx context.Context) error {
		teamID, err := u.teamRepository.GetTeamIDByUserID(ctx, authorID)
		if err != nil {
			return err
		}

		excludedUsers := []string{authorID}
		teammates, err := u.teamRepository.GetActiveTeammates(ctx, teamID, excludedUsers, 2)
		if err != nil {
			return err
		}

		pr, err = u.pullRequestsRepository.PullRequestCreate(ctx, authorID, prID, prName, teammates)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return pr, nil
}

func (u *useCase) PullRequestMerge(
	ctx context.Context,
	prID string,
) (*models.PR, error) {
	pr, err := u.pullRequestsRepository.PullRequestMerge(ctx, prID)
	if err != nil {
		return nil, err
	}

	return pr, nil
}

func (u *useCase) PullRequestReassign(
	ctx context.Context,
	prID, oldReviewerID string,
) (*models.PR, string, error) {
	var pr *models.PR
	var newReviewerID string

	err := u.transactor.WithTx(ctx, func(ctx context.Context) error {
		teamID, err := u.teamRepository.GetTeamIDByUserID(ctx, oldReviewerID)
		if err != nil {
			return err
		}

		pr, err = u.pullRequestsRepository.GetPullRequest(ctx, prID)
		if err != nil {
			return err
		}
		wasReviewer := false
		for _, reviewerID := range pr.AssignedReviewers {
			if reviewerID == oldReviewerID {
				wasReviewer = true
			}
		}
		if !wasReviewer {
			return modelsErr.ErrNotAssigned
		}
		excludedUsers := []string{pr.AuthorID}
		excludedUsers = append(excludedUsers, pr.AssignedReviewers...)

		teammates, err := u.teamRepository.GetActiveTeammates(ctx, teamID, excludedUsers, 1)
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
		for _, reviewer := range pr.AssignedReviewers {
			if reviewer != oldReviewerID {
				newReviewers = append(newReviewers, reviewer)
			}
		}

		pr.AssignedReviewers = newReviewers

		return nil
	})

	if err != nil {
		return nil, "", err
	}

	return pr, newReviewerID, nil
}

package repo_middleware

import (
	"context"

	"github.com/Tortik3000/PR-service/internal/models"
)

type (
	metricsRepo interface {
		GetReview(ctx context.Context, userID string) ([]models.PRShort, error)
		SetIsActive(ctx context.Context, userID string, isActive bool) (*models.User, error)
		TeamAdd(ctx context.Context, team models.Team) error
		TeamGet(ctx context.Context, teamName string) (*models.Team, error)
		PullRequestCreate(ctx context.Context, authorID, prID, prName string, teammates []string) (*models.PR, error)
		PullRequestMerge(ctx context.Context, prID string) (*models.PR, error)
		PullRequestReassign(ctx context.Context, prID, oldReviewerID, newReviewerID string) error
		GetActiveTeammates(ctx context.Context, teamID string, excludedUsers []string, count uint64) ([]string, error)
		GetTeamIDByUserID(ctx context.Context, userID string) (teamID string, err error)
		GetPullRequest(ctx context.Context, prID string) (*models.PR, error)
	}
)

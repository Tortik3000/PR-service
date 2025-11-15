package pr_service

import (
	"context"

	"github.com/Tortik3000/PR-service/internal/models"

	"go.uber.org/zap"
)

type (
	userRepository interface {
		GetReview(ctx context.Context, userID string) ([]models.PRShort, error)
		SetIsActive(ctx context.Context, userID string, isActive bool) (*models.User, error)
	}

	teamRepository interface {
		TeamAdd(ctx context.Context, team models.Team) error
		TeamGet(ctx context.Context, teamName string) (*models.Team, error)
		GetActiveTeammates(ctx context.Context, teamID string, excludedUsers []string, count uint64) ([]string, error)
		GetTeamIDByUserID(ctx context.Context, userID string) (teamID string, err error)
	}

	pullRequestsRepository interface {
		PullRequestCreate(ctx context.Context, authorID, prID, prName string, teammates []string) (*models.PR, error)
		PullRequestMerge(ctx context.Context, prID string) (*models.PR, error)
		PullRequestReassign(ctx context.Context, prID, oldReviewerID, newReviewerID string) error
		GetPullRequest(ctx context.Context, prID string) (*models.PR, error)
	}

	transactor interface {
		WithTx(ctx context.Context, function func(ctx context.Context) error) error
	}
)

type useCase struct {
	logger                 *zap.Logger
	pullRequestsRepository pullRequestsRepository
	teamRepository         teamRepository
	userRepository         userRepository
	transactor             transactor
}

func NewUseCase(
	logger *zap.Logger,
	pullRequestsRepository pullRequestsRepository,
	teamRepository teamRepository,
	userRepository userRepository,
	transactor transactor,
) *useCase {
	return &useCase{
		logger:                 logger,
		pullRequestsRepository: pullRequestsRepository,
		teamRepository:         teamRepository,
		userRepository:         userRepository,
		transactor:             transactor,
	}
}

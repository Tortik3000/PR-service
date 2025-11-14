package pr_service

import (
	"context"

	pr "github.com/Tortik3000/PR-service/internal/models/pull_request"
	"github.com/Tortik3000/PR-service/internal/models/team"
	"github.com/Tortik3000/PR-service/internal/models/user"
	"go.uber.org/zap"
)

// todo transactor + rollback

type (
	userRepository interface {
		GetReview(ctx context.Context, userID string) ([]pr.DBPullRequestShort, error)
		SetIsActive(ctx context.Context, userID string, isActive bool) (*user.DBUser, error)
	}

	teamRepository interface {
		TeamAdd(ctx context.Context, team *team.DBTeam) error
		TeamGet(ctx context.Context, teamName string) (*team.DBTeam, error)
	}

	pullRequestsRepository interface {
		PullRequestCreate(ctx context.Context, authorID, prID, prName string) (*pr.DBPullRequest, error)
		PullRequestMerge(ctx context.Context, prID string) (*pr.DBPullRequest, error)
		PullRequestReassign(ctx context.Context, prID, oldUserID string) (*pr.DBPullRequest, string, error)
	}
)

type useCase struct {
	logger                 *zap.Logger
	pullRequestsRepository pullRequestsRepository
	teamRepository         teamRepository
	userRepository         userRepository
}

func NewUseCase(
	logger *zap.Logger,
	pullRequestsRepository pullRequestsRepository,
	teamRepository teamRepository,
	userRepository userRepository,
) *useCase {
	return &useCase{
		logger:                 logger,
		pullRequestsRepository: pullRequestsRepository,
		teamRepository:         teamRepository,
		userRepository:         userRepository,
	}
}

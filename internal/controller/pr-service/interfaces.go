package pr_service

import (
	"context"

	pr "github.com/Tortik3000/PR-service/internal/models/pull_request"
	"github.com/Tortik3000/PR-service/internal/models/team"
	"github.com/Tortik3000/PR-service/internal/models/user"
)

type (
	userUseCase interface {
		GetReview(ctx context.Context, userID string) ([]pr.PRShort, error)
		SetIsActive(ctx context.Context, userID string, isActive bool) (*user.User, error)
	}

	teamUseCase interface {
		TeamAdd(ctx context.Context, team *team.Team) error
		TeamGet(ctx context.Context, teamName string) (*team.Team, error)
	}

	pullRequestUseCase interface {
		PullRequestCreate(ctx context.Context, authorID, prID, prName string) (*pr.PR, error)
		PullRequestMerge(ctx context.Context, prID string) (*pr.PR, error)
		PullRequestReassign(ctx context.Context, prID, oldUserID string) (*pr.PR, string, error)
	}
)

package middleware

import (
	"context"

	pr "github.com/Tortik3000/PR-service/internal/models/pull_request"
	"github.com/Tortik3000/PR-service/internal/models/team"
	"github.com/Tortik3000/PR-service/internal/models/user"
)

type (
	metricsRepo interface {
		GetReview(ctx context.Context, userID string) ([]pr.DBPullRequestShort, error)
		SetIsActive(ctx context.Context, userID string, isActive bool) (*user.DBUser, error)
		TeamAdd(ctx context.Context, team *team.DBTeam) error
		TeamGet(ctx context.Context, teamName string) (*team.DBTeam, error)
		PullRequestCreate(ctx context.Context, authorID, prID, prName string, teammates []string) (*pr.DBPullRequest, error)
		PullRequestMerge(ctx context.Context, prID string) (*pr.DBPullRequest, error)
		PullRequestReassign(ctx context.Context, prID, oldReviewerID, newReviewerID string) error
		GetActiveTeammates(ctx context.Context, teamID string, excludedUsers []string, count uint64) ([]string, error)
		GetTeamIDByUserID(ctx context.Context, userID string) (teamID string, err error)
		GetPullRequest(ctx context.Context, prID string) (*pr.DBPullRequest, error)
	}
)

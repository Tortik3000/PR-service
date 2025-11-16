package pr_service

import (
	"context"

	"github.com/Tortik3000/PR-service/internal/models"
)

//go:generate mockgen_uber -source=interfaces.go -destination=mocks/use_case_mock.go -package=mocks

type (
	userUseCase interface {
		GetReview(ctx context.Context, userID string) ([]models.PRShort, error)
		SetIsActive(ctx context.Context, userID string, isActive bool) (*models.User, error)
	}

	teamUseCase interface {
		TeamAdd(ctx context.Context, team models.Team) error
		TeamGet(ctx context.Context, teamName string) (*models.Team, error)
	}

	pullRequestUseCase interface {
		PullRequestCreate(ctx context.Context, authorID, prID, prName string) (*models.PR, error)
		PullRequestMerge(ctx context.Context, prID string) (*models.PR, error)
		PullRequestReassign(ctx context.Context, prID, oldUserID string) (*models.PR, string, error)
	}
)

package pr_service

import (
	"context"

	pr "github.com/Tortik3000/PR-service/internal/models/pull_request"
	"github.com/Tortik3000/PR-service/internal/models/user"
)

func (u *useCase) GetReview(
	ctx context.Context,
	userID string,
) ([]pr.PRShort, error) {
	dbPRs, err := u.userRepository.GetReview(ctx, userID)
	if err != nil {
		return nil, err
	}

	return pr.FromDBShortSlice(dbPRs), nil
}

func (u *useCase) SetIsActive(
	ctx context.Context,
	userID string,
	isActive bool,
) (*user.User, error) {
	dbUser, err := u.userRepository.SetIsActive(ctx, userID, isActive)
	if err != nil {
		return nil, err
	}

	return user.FromDB(dbUser), nil
}

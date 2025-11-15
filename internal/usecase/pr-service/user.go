package pr_service

import (
	"context"

	"github.com/Tortik3000/PR-service/internal/models"
)

func (u *useCase) GetReview(
	ctx context.Context,
	userID string,
) ([]models.PRShort, error) {
	prs, err := u.userRepository.GetReview(ctx, userID)
	if err != nil {
		return nil, err
	}

	return prs, nil
}

func (u *useCase) SetIsActive(
	ctx context.Context,
	userID string,
	isActive bool,
) (*models.User, error) {
	user, err := u.userRepository.SetIsActive(ctx, userID, isActive)
	if err != nil {
		return nil, err
	}

	return user, nil
}

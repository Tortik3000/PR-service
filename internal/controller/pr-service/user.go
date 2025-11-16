package pr_service

import (
	"context"
	"errors"

	"go.uber.org/zap"

	api "github.com/Tortik3000/PR-service/generated/api/pr-service"
	"github.com/Tortik3000/PR-service/internal/controller/pr-service/dto"
	modelsErr "github.com/Tortik3000/PR-service/internal/models/errors"
)

func (p *prService) GetUsersGetReview(
	ctx context.Context,
	request api.GetUsersGetReviewRequestObject,
) (api.GetUsersGetReviewResponseObject, error) {
	p.logger.Info("GetUsersGetReview called",
		zap.String("user_id", request.Params.UserId),
	)

	prs, err := p.userUseCase.GetReview(ctx, request.Params.UserId)
	if err != nil {
		return nil, modelsErr.ErrInternal
	}

	p.logger.Info("GetUsersGetReview success",
		zap.Any("prs", prs),
	)

	return api.GetUsersGetReview200JSONResponse{
		UserId:       request.Params.UserId,
		PullRequests: dto.ToAPIShortSlice(prs),
	}, nil
}

func (p *prService) PostUsersSetIsActive(
	ctx context.Context,
	request api.PostUsersSetIsActiveRequestObject,
) (api.PostUsersSetIsActiveResponseObject, error) {
	body := request.Body
	p.logger.Info("PostUsersSetIsActive called",
		zap.String("user_id", body.UserId),
		zap.Any("is_active", body.IsActive),
	)

	user, err := p.userUseCase.SetIsActive(ctx, body.UserId, body.IsActive)
	if err != nil {
		switch {
		case errors.Is(err, modelsErr.ErrUserNotFound):
			return api.PostUsersSetIsActive404JSONResponse{
				Error: newErrorResponse(api.NOTFOUND, err.Error()).Error,
			}, nil

		default:
			return nil, modelsErr.ErrInternal
		}
	}

	p.logger.Info("PostUsersSetIsActive success",
		zap.String("user_id", user.ID),
		zap.Any("is_active", user.IsActive),
	)

	return api.PostUsersSetIsActive200JSONResponse{
		User: dto.ToAPIUser(user),
	}, nil
}

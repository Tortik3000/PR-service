package pr_service

import (
	"context"
	"errors"

	generated "github.com/Tortik3000/PR-service/generated/api/pr-service"
	"github.com/Tortik3000/PR-service/internal/controller/pr-service/dto"
	modelsErr "github.com/Tortik3000/PR-service/internal/models/errors"
)

func (p *prService) GetUsersGetReview(
	ctx context.Context,
	request generated.GetUsersGetReviewRequestObject,
) (generated.GetUsersGetReviewResponseObject, error) {
	prs, err := p.userUseCase.GetReview(ctx, request.Params.UserId)
	if err != nil {
		return nil, err
	}

	return generated.GetUsersGetReview200JSONResponse{
		UserId:       request.Params.UserId,
		PullRequests: dto.ToAPIShortSlice(prs),
	}, nil
}

func (p *prService) PostUsersSetIsActive(
	ctx context.Context,
	request generated.PostUsersSetIsActiveRequestObject,
) (generated.PostUsersSetIsActiveResponseObject, error) {
	body := request.Body

	user, err := p.userUseCase.SetIsActive(ctx, body.UserId, body.IsActive)
	if err != nil {
		switch {
		case errors.Is(err, modelsErr.ErrUserNotFound):
			return generated.PostUsersSetIsActive404JSONResponse{
				Error: newErrorResponse(generated.NOTFOUND, err.Error()).Error,
			}, nil

		default:
			return nil, err
		}

	}

	return generated.PostUsersSetIsActive200JSONResponse{
		User: dto.ToAPIUser(user),
	}, nil
}

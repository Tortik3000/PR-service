package pr_service

import (
	"context"
	"errors"

	generated "github.com/Tortik3000/PR-service/generated/api/pr-service"
	modelsErr "github.com/Tortik3000/PR-service/internal/models/errors"
	models "github.com/Tortik3000/PR-service/internal/models/pull_request"
)

func (p *prService) PostPullRequestCreate(
	ctx context.Context,
	request generated.PostPullRequestCreateRequestObject,
) (generated.PostPullRequestCreateResponseObject, error) {
	body := request.Body

	pr, err := p.pullRequestUseCase.PullRequestCreate(
		ctx,
		body.AuthorId,
		body.PullRequestId,
		body.PullRequestName,
	)

	if err != nil {
		switch {
		case errors.Is(err, modelsErr.ErrTeamNotFound) || errors.Is(err, modelsErr.ErrUserNotFound):
			return generated.PostPullRequestCreate404JSONResponse{
				Error: newErrorResponse(generated.NOTFOUND, err.Error()).Error,
			}, nil

		case errors.Is(err, modelsErr.ErrPullRequestExist):
			return generated.PostPullRequestCreate409JSONResponse{
				Error: newErrorResponse(generated.PREXISTS, err.Error()).Error,
			}, nil
		default:
			return nil, err
		}
	}

	apiPR := models.ToAPIPullRequest(pr)
	return generated.PostPullRequestCreate201JSONResponse{
		Pr: apiPR,
	}, nil
}

func (p *prService) PostPullRequestMerge(
	ctx context.Context,
	request generated.PostPullRequestMergeRequestObject,
) (generated.PostPullRequestMergeResponseObject, error) {
	body := request.Body

	pr, err := p.pullRequestUseCase.PullRequestMerge(ctx, body.PullRequestId)
	if err != nil {
		switch {
		case errors.Is(err, modelsErr.ErrPRNotFound):
			return generated.PostPullRequestMerge404JSONResponse{
				Error: newErrorResponse(generated.NOTFOUND, "pull request not found").Error,
			}, nil

		default:
			return nil, err
		}
	}

	apiPR := models.ToAPIPullRequest(pr)
	return generated.PostPullRequestMerge200JSONResponse{
		Pr: apiPR,
	}, nil
}

func (p *prService) PostPullRequestReassign(
	ctx context.Context,
	request generated.PostPullRequestReassignRequestObject,
) (generated.PostPullRequestReassignResponseObject, error) {
	body := request.Body

	pr, replacedBy, err := p.pullRequestUseCase.PullRequestReassign(
		ctx,
		body.PullRequestId,
		body.OldUserId,
	)
	if err != nil {
		switch {
		case errors.Is(err, modelsErr.ErrPRNotFound) || errors.Is(err, modelsErr.ErrUserNotFound):
			return generated.PostPullRequestReassign404JSONResponse{
				Error: newErrorResponse(generated.NOTFOUND, err.Error()).Error,
			}, nil
		case errors.Is(err, modelsErr.ErrPRMerged):
			return generated.PostPullRequestReassign409JSONResponse{
				Error: newErrorResponse(generated.PRMERGED, err.Error()).Error,
			}, nil
		case errors.Is(err, modelsErr.ErrNotAssigned):
			return generated.PostPullRequestReassign409JSONResponse{
				Error: newErrorResponse(generated.NOTASSIGNED, err.Error()).Error,
			}, nil

		case errors.Is(err, modelsErr.ErrNotActiveCandidate):
			return generated.PostPullRequestReassign409JSONResponse{
				Error: newErrorResponse(generated.NOCANDIDATE, err.Error()).Error,
			}, nil

		default:
			return nil, err
		}
	}

	apiPR := models.ToAPIPullRequest(pr)
	return generated.PostPullRequestReassign200JSONResponse{
		Pr:         *apiPR,
		ReplacedBy: replacedBy,
	}, nil
}

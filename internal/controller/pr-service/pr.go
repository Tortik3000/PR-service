package pr_service

import (
	"context"
	"errors"

	"go.uber.org/zap"

	api "github.com/Tortik3000/PR-service/generated/api/pr-service"
	"github.com/Tortik3000/PR-service/internal/controller/pr-service/dto"
	modelsErr "github.com/Tortik3000/PR-service/internal/models/errors"
)

func (p *prService) PostPullRequestCreate(
	ctx context.Context,
	request api.PostPullRequestCreateRequestObject,
) (api.PostPullRequestCreateResponseObject, error) {
	body := request.Body
	p.logger.Info("PostPullRequestCreate called",
		zap.String("author_id", body.AuthorId),
		zap.String("pr_id", body.PullRequestId),
		zap.String("pr_name", body.PullRequestName),
	)

	pr, err := p.pullRequestUseCase.PullRequestCreate(
		ctx,
		body.AuthorId,
		body.PullRequestId,
		body.PullRequestName,
	)

	if err != nil {
		switch {
		case errors.Is(err, modelsErr.ErrTeamNotFound) || errors.Is(err, modelsErr.ErrUserNotFound):
			return api.PostPullRequestCreate404JSONResponse{
				Error: newErrorResponse(api.NOTFOUND, err.Error()).Error,
			}, nil

		case errors.Is(err, modelsErr.ErrPullRequestExist):
			return api.PostPullRequestCreate409JSONResponse{
				Error: newErrorResponse(api.PREXISTS, err.Error()).Error,
			}, nil
		default:
			return nil, modelsErr.ErrInternal
		}
	}

	apiPR := dto.ToAPIPullRequest(pr)
	p.logger.Info("PostPullRequestCreate success",
		zap.String("author_id", apiPR.AuthorId),
		zap.String("pr_id", apiPR.PullRequestId),
		zap.String("pr_name", apiPR.PullRequestName),
		zap.Strings("pr_assigned_reviewers", apiPR.AssignedReviewers),
	)
	return api.PostPullRequestCreate201JSONResponse{
		Pr: apiPR,
	}, nil
}

func (p *prService) PostPullRequestMerge(
	ctx context.Context,
	request api.PostPullRequestMergeRequestObject,
) (api.PostPullRequestMergeResponseObject, error) {
	body := request.Body
	p.logger.Info("PostPullRequestMerge called",
		zap.String("pr_id", body.PullRequestId),
	)

	pr, err := p.pullRequestUseCase.PullRequestMerge(ctx, body.PullRequestId)
	if err != nil {
		switch {
		case errors.Is(err, modelsErr.ErrPRNotFound):
			return api.PostPullRequestMerge404JSONResponse{
				Error: newErrorResponse(api.NOTFOUND, "pull request not found").Error,
			}, nil

		default:
			return nil, modelsErr.ErrInternal
		}
	}

	apiPR := dto.ToAPIPullRequest(pr)
	p.logger.Info("PostPullRequestCreate success",
		zap.String("pr_id", apiPR.PullRequestId),
		zap.String("pr_status", string(apiPR.Status)),
	)
	return api.PostPullRequestMerge200JSONResponse{
		Pr: apiPR,
	}, nil
}

func (p *prService) PostPullRequestReassign(
	ctx context.Context,
	request api.PostPullRequestReassignRequestObject,
) (api.PostPullRequestReassignResponseObject, error) {
	body := request.Body
	p.logger.Info("PostPullRequestReassign called",
		zap.String("pr_id", body.PullRequestId),
		zap.String("old_user_id", body.OldUserId),
	)

	pr, replacedBy, err := p.pullRequestUseCase.PullRequestReassign(
		ctx,
		body.PullRequestId,
		body.OldUserId,
	)
	if err != nil {
		switch {
		case errors.Is(err, modelsErr.ErrPRNotFound) || errors.Is(err, modelsErr.ErrUserNotFound):
			return api.PostPullRequestReassign404JSONResponse{
				Error: newErrorResponse(api.NOTFOUND, err.Error()).Error,
			}, nil
		case errors.Is(err, modelsErr.ErrPRMerged):
			return api.PostPullRequestReassign409JSONResponse{
				Error: newErrorResponse(api.PRMERGED, err.Error()).Error,
			}, nil
		case errors.Is(err, modelsErr.ErrNotAssigned):
			return api.PostPullRequestReassign409JSONResponse{
				Error: newErrorResponse(api.NOTASSIGNED, err.Error()).Error,
			}, nil

		case errors.Is(err, modelsErr.ErrNotActiveCandidate):
			return api.PostPullRequestReassign409JSONResponse{
				Error: newErrorResponse(api.NOCANDIDATE, err.Error()).Error,
			}, nil

		default:
			return nil, modelsErr.ErrInternal
		}
	}

	apiPR := dto.ToAPIPullRequest(pr)
	p.logger.Info("PostPullRequestReassign success",
		zap.String("pr_id", apiPR.PullRequestId),
		zap.Strings("pr_assigned_reviewers", apiPR.AssignedReviewers),
		zap.String("new_user_id", replacedBy),
	)
	return api.PostPullRequestReassign200JSONResponse{
		Pr:         *apiPR,
		ReplacedBy: replacedBy,
	}, nil
}

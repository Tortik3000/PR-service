package controller

import (
	"context"

	"go.uber.org/zap"

	generated "github.com/Tortik3000/PR-service/generated/api/pr-service"
)

type prService struct {
	logger *zap.Logger
}

func (p prService) PostPullRequestCreate(ctx context.Context, request generated.PostPullRequestCreateRequestObject) (generated.PostPullRequestCreateResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (p prService) PostPullRequestMerge(ctx context.Context, request generated.PostPullRequestMergeRequestObject) (generated.PostPullRequestMergeResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (p prService) PostPullRequestReassign(ctx context.Context, request generated.PostPullRequestReassignRequestObject) (generated.PostPullRequestReassignResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (p prService) PostTeamAdd(ctx context.Context, request generated.PostTeamAddRequestObject) (generated.PostTeamAddResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (p prService) GetTeamGet(ctx context.Context, request generated.GetTeamGetRequestObject) (generated.GetTeamGetResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (p prService) GetUsersGetReview(ctx context.Context, request generated.GetUsersGetReviewRequestObject) (generated.GetUsersGetReviewResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (p prService) PostUsersSetIsActive(ctx context.Context, request generated.PostUsersSetIsActiveRequestObject) (generated.PostUsersSetIsActiveResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

var _ generated.StrictServerInterface = (*prService)(nil)

func NewPRService(logger *zap.Logger) *prService {
	return &prService{
		logger: logger,
	}
}

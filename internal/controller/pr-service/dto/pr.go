package dto

import (
	api "github.com/Tortik3000/PR-service/generated/api/pr-service"
	"github.com/Tortik3000/PR-service/internal/models"
)

func ToAPIShortSlice(prs []models.PRShort) (ret []api.PullRequestShort) {
	ret = make([]api.PullRequestShort, len(prs))
	for i, dbPR := range prs {
		ret[i] = *ToAPIPullRequestShort(&dbPR)
	}

	return
}

func ToAPIPullRequest(pr *models.PR) *api.PullRequest {
	if pr == nil {
		return nil
	}
	return &api.PullRequest{
		PullRequestId:     pr.ID,
		PullRequestName:   pr.Name,
		AuthorId:          pr.AuthorID,
		Status:            statusToApiStatus(pr.Status),
		AssignedReviewers: pr.AssignedReviewers,
		CreatedAt:         pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}

func ToAPIPullRequestShort(pr *models.PRShort) *api.PullRequestShort {
	if pr == nil {
		return nil
	}
	return &api.PullRequestShort{
		PullRequestId:   pr.ID,
		PullRequestName: pr.Name,
		AuthorId:        pr.AuthorID,
		Status:          api.PullRequestShortStatus(statusToApiStatus(pr.Status)),
	}
}

func statusToApiStatus(status models.PRStatus) api.PullRequestStatus {
	switch status {
	case models.PRStatusMERGED:
		return api.PullRequestStatusMERGED
	case models.PRStatusOPEN:
		return api.PullRequestStatusOPEN
	default:
		return ""
	}
}

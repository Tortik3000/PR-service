package pull_request

import (
	api "github.com/Tortik3000/PR-service/generated/api/pr-service"
)

func FromDB(dbPR *DBPullRequest) *PR {
	return &PR{
		Id:                dbPR.ID,
		Name:              dbPR.Name,
		AuthorID:          dbPR.AuthorID,
		Status:            dbPR.Status,
		AssignedReviewers: dbPR.AssignedReviewers,
		CreatedAt:         dbPR.CreatedAt,
		MergedAt:          dbPR.MergedAt,
	}
}

func ToDB(pr *PR) *DBPullRequest {
	return &DBPullRequest{
		ID:                pr.Id,
		Name:              pr.Name,
		AuthorID:          pr.AuthorID,
		Status:            pr.Status,
		AssignedReviewers: pr.AssignedReviewers,
		CreatedAt:         pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}

func FromDBShort(dbPR *DBPullRequestShort) *PRShort {
	return &PRShort{
		Id:       dbPR.ID,
		Name:     dbPR.Name,
		AuthorID: dbPR.AuthorID,
		Status:   dbPR.Status,
	}
}

func FromDBShortSlice(dbPRs []DBPullRequestShort) (ret []PRShort) {
	ret = make([]PRShort, len(dbPRs))
	for i, dbPR := range dbPRs {
		ret[i] = *FromDBShort(&dbPR)
	}

	return
}

func ToAPIShortSlice(prs []PRShort) (ret []api.PullRequestShort) {
	ret = make([]api.PullRequestShort, len(prs))
	for i, dbPR := range prs {
		ret[i] = *ToAPIPullRequestShort(&dbPR)
	}

	return
}
func ToDBShort(pr *PRShort) *DBPullRequestShort {
	return &DBPullRequestShort{
		ID:       pr.Id,
		Name:     pr.Name,
		AuthorID: pr.AuthorID,
		Status:   pr.Status,
	}
}

func ToAPIPullRequest(pr *PR) *api.PullRequest {
	return &api.PullRequest{
		PullRequestId:     pr.Id,
		PullRequestName:   pr.Name,
		AuthorId:          pr.AuthorID,
		Status:            statusToApiStatus(pr.Status),
		AssignedReviewers: pr.AssignedReviewers,
		CreatedAt:         pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}

func ToAPIPullRequestShort(pr *PRShort) *api.PullRequestShort {
	return &api.PullRequestShort{
		PullRequestId:   pr.Id,
		PullRequestName: pr.Name,
		AuthorId:        pr.AuthorID,
		Status:          api.PullRequestShortStatus(statusToApiStatus(pr.Status)),
	}
}

func statusToApiStatus(status PRStatus) api.PullRequestStatus {
	switch status {
	case MERGED:
		return api.PullRequestStatusMERGED
	case OPEN:
		return api.PullRequestStatusOPEN
	default:
		return ""
	}
}

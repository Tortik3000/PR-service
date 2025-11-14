package pull_request

import "time"

type DBPullRequest struct {
	AssignedReviewers []string
	AuthorID          string
	CreatedAt         *time.Time
	MergedAt          *time.Time
	ID                string
	Name              string
	Status            PRStatus
}

type DBPullRequestShort struct {
	AuthorID string
	ID       string
	Name     string
	Status   PRStatus
}

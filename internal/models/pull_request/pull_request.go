package pull_request

import (
	"time"
)

type PR struct {
	AssignedReviewers []string
	AuthorID          string
	CreatedAt         *time.Time
	MergedAt          *time.Time
	Id                string
	Name              string
	Status            PRStatus
}

type PRShort struct {
	AuthorID string
	Id       string
	Name     string
	Status   PRStatus
}

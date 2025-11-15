package models

import (
	"time"
)

type PR struct {
	AssignedReviewers []string
	AuthorID          string
	CreatedAt         *time.Time
	MergedAt          *time.Time
	ID                string
	Name              string
	Status            PRStatus
}

type PRShort struct {
	AuthorID string
	ID       string
	Name     string
	Status   PRStatus
}

type PRStatus int

const (
	PRStatusOPEN PRStatus = iota
	PRStatusMERGED
)

package errors

import "errors"

var (
	ErrTeamExist        = errors.New("team with this name already exists")
	ErrPullRequestExist = errors.New("pull request with this ID already exists")

	ErrUserNotFound = errors.New("user not found")
	ErrTeamNotFound = errors.New("team not found")
	ErrPRNotFound   = errors.New("pull request not found")

	ErrPRMerged     = errors.New("pr already merged")
	ErrNotActive    = errors.New("not active candidate")
	ErrNotAssigned  = errors.New("the user was not assigned as a reviewer for this PR")
	ErrNotCandidate = errors.New("not candidate")
)

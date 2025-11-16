package errors

import "errors"

var (
	ErrTeamExist        = errors.New("team with this name already exists")
	ErrPullRequestExist = errors.New("pull request with this ID already exists")

	ErrUserNotFound = errors.New("user not found")
	ErrTeamNotFound = errors.New("team not found")
	ErrPRNotFound   = errors.New("pull request not found")

	ErrPRMerged           = errors.New("pr already merged")
	ErrNotAssigned        = errors.New("the user was not assigned as a reviewer for this PR")
	ErrNotActiveCandidate = errors.New("no active replacement candidate in team")

	ErrInternal = errors.New("internal error")
)

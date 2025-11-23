package storage

import "errors"

var (
	ErrTeamNotFound        = errors.New("team not found")
	ErrTeamAlreadyExists   = errors.New("team already exists")
	ErrPRAlreadyExists     = errors.New("PR already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrPRNotFound          = errors.New("pull request not found")
	ErrPRAlreadyMerged     = errors.New("pull request already merged")
	ErrReviewerNotAssigned = errors.New("reviewer not assigned")
	ErrNoCandidate         = errors.New("no candidate")
	ErrRowsNotClosed       = errors.New("rows not closed")
	ErrRollbackFailed      = errors.New("rollback failed")
)

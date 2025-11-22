package domain

import "time"

type User struct {
	ID       string
	Name     string
	IsActive bool
}

type Team struct {
	Name  string
	Users []User
}

type PullRequest struct {
	ID        string
	Name      string
	Author    User
	Status    string
	Reviewers []User
	MergedAt  time.Time
}

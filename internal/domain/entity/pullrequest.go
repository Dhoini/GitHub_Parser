package entity

import "time"

type PullRequest struct {
	ID           int64
	Number       int
	Title        string
	Body         string
	State        string
	AuthorLogin  string
	RepositoryID int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
	MergedAt     *time.Time
	ClosedAt     *time.Time
}

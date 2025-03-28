package entity

import "time"

type Issue struct {
	ID           int64
	Number       int
	Title        string
	Body         string
	State        string
	AuthorLogin  string
	RepositoryID int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ClosedAt     *time.Time
}

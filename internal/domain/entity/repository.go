package entity

import "time"

type Repository struct {
	ID              int64
	Name            string
	FullName        string
	Description     string
	IsPrivate       bool
	OwnerLogin      string
	Language        string
	StarsCount      int
	ForksCount      int
	OpenIssuesCount int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

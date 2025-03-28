package service

import (
	"context"
	"github.com/Dhoini/GitHub_Parser/internal/domain/entity"
)

type ParsingJobParams struct {
	OwnerName   string
	RepoName    string
	ParseIssues bool
	ParsePRs    bool
	ParseUsers  bool
}

type ParsingJobStatus struct {
	ID           string
	Status       string // "pending", "in_progress", "completed", "failed"
	Progress     int    // 0-100
	ErrorMessage string
	CreatedAt    string
	UpdatedAt    string
}

type ParserService interface {
	ParseRepository(ctx context.Context, owner, name string) (*entity.Repository, error)
	ParseIssues(ctx context.Context, owner, repo string) ([]*entity.Issue, error)
	ParsePullRequests(ctx context.Context, owner, repo string) ([]*entity.PullRequest, error)
	ParseUser(ctx context.Context, username string) (*entity.User, error)

	StartParsingJob(ctx context.Context, params ParsingJobParams) (string, error)
	GetParsingJobStatus(ctx context.Context, jobID string) (*ParsingJobStatus, error)
}

package service

import (
	"context"
	"github.com/Dhoini/GitHub_Parser/internal/domain/entity"
)

type GithubService interface {
	GetRepository(ctx context.Context, owner, name string) (*entity.Repository, error)
	GetIssues(ctx context.Context, owner, repo string, page, perPage int) ([]*entity.Issue, error)
	GetPullRequests(ctx context.Context, owner, repo string, page, perPage int) ([]*entity.PullRequest, error)
	GetUser(ctx context.Context, username string) (*entity.User, error)
}

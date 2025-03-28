package repository

import (
	"context"
	"github.com/Dhoini/GitHub_Parser/internal/domain/entity"
)

type PullRequestFilter struct {
	RepositoryID int64
	State        string // "open", "closed", "all"
	Limit        int
	Offset       int
}

type PullRequestRepository interface {
	Save(ctx context.Context, pr *entity.PullRequest) error
	FindByID(ctx context.Context, id int64) (*entity.PullRequest, error)
	FindByNumber(ctx context.Context, repoID int64, number int) (*entity.PullRequest, error)
	List(ctx context.Context, filter PullRequestFilter) ([]*entity.PullRequest, error)
}

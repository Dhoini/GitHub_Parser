package repository

import (
	"context"
	"github.com/Dhoini/GitHub_Parser/internal/domain/entity"
)

type IssueFilter struct {
	RepositoryID int64
	State        string
	Limit        int
	Offset       int
}

type IssueRepository interface {
	Save(ctx context.Context, issue *entity.Issue) error
	FindByID(ctx context.Context, id int64) (*entity.Issue, error)
	List(ctx context.Context, filter IssueFilter) ([]*entity.Issue, error)
}

package repository

import (
	"context"
	"github.com/Dhoini/GitHub_Parser/internal/domain/entity"
)

type RepositoryFilter struct {
	OwnerLogin string
	Language   string
	MinStars   int
	Limit      int
	Offset     int
}

type RepositoryRepository interface {
	Save(ctx context.Context, repo *entity.Repository) error
	FindByID(ctx context.Context, id int64) (*entity.Repository, error)
	FindByOwnerAndName(ctx context.Context, owner, name string) (*entity.Repository, error)
	List(ctx context.Context, filter RepositoryFilter) ([]*entity.Repository, error)
}

package repository

import (
	"context"
	"github.com/Dhoini/GitHub_Parser/internal/domain/entity"
)

type UserFilter struct {
	Login  string
	Limit  int
	Offset int
}

type UserRepository interface {
	Save(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id int64) (*entity.User, error)
	FindByLogin(ctx context.Context, login string) (*entity.User, error)
	List(ctx context.Context, filter UserFilter) ([]*entity.User, error)
}

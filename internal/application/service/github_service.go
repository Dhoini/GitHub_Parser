package service

import (
	"context"
	"github.com/Dhoini/GitHub_Parser/internal/domain/entity"

	"github.com/google/go-github/v39/github"
	"go.uber.org/zap"
)

type GithubServiceImpl struct {
	client *github.Client
	logger *zap.Logger
}

func NewGithubService(client *github.Client, logger *zap.Logger) *GithubServiceImpl {
	return &GithubServiceImpl{
		client: client,
		logger: logger,
	}
}

func (s *GithubServiceImpl) GetRepository(ctx context.Context, owner, name string) (*entity.Repository, error) {
	repo, _, err := s.client.Repositories.Get(ctx, owner, name)
	if err != nil {
		s.logger.Error("Error getting repository", zap.Error(err), zap.String("owner", owner), zap.String("name", name))
		return nil, err
	}

	return &entity.Repository{
		ID:              repo.GetID(),
		Name:            repo.GetName(),
		FullName:        repo.GetFullName(),
		Description:     repo.GetDescription(),
		IsPrivate:       repo.GetPrivate(),
		OwnerLogin:      repo.GetOwner().GetLogin(),
		Language:        repo.GetLanguage(),
		StarsCount:      repo.GetStargazersCount(),
		ForksCount:      repo.GetForksCount(),
		OpenIssuesCount: repo.GetOpenIssuesCount(),
		CreatedAt:       repo.GetCreatedAt().Time,
		UpdatedAt:       repo.GetUpdatedAt().Time,
	}, nil
}

// Аналогично реализуйте остальные методы интерфейса

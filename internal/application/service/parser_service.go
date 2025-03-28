package service

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/Dhoini/GitHub_Parser/internal/domain/entity"
	"github.com/Dhoini/GitHub_Parser/internal/domain/repository"
	domainService "github.com/Dhoini/GitHub_Parser/internal/domain/service"
)

type ParserServiceImpl struct {
	githubService domainService.GithubService
	repoRepo      repository.RepositoryRepository
	issueRepo     repository.IssueRepository
	prRepo        repository.PullRequestRepository
	userRepo      repository.UserRepository
	logger        *zap.Logger
}

func NewParserService(
	githubService domainService.GithubService,
	repoRepo repository.RepositoryRepository,
	issueRepo repository.IssueRepository,
	prRepo repository.PullRequestRepository,
	userRepo repository.UserRepository,
	logger *zap.Logger,
) *ParserServiceImpl {
	return &ParserServiceImpl{
		githubService: githubService,
		repoRepo:      repoRepo,
		issueRepo:     issueRepo,
		prRepo:        prRepo,
		userRepo:      userRepo,
		logger:        logger,
	}
}

func (s *ParserServiceImpl) ParseRepository(ctx context.Context, owner, name string) (*entity.Repository, error) {
	// Получаем репозиторий из GitHub API
	repo, err := s.githubService.GetRepository(ctx, owner, name)
	if err != nil {
		return nil, err
	}

	// Сохраняем в БД
	if err := s.repoRepo.Save(ctx, repo); err != nil {
		s.logger.Error("Error saving repository", zap.Error(err))
		return nil, err
	}

	return repo, nil
}

// Аналогично реализуйте остальные методы интерфейса

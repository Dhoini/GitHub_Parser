package service

import (
	"context"
	"github.com/Dhoini/GitHub_Parser/internal/domain/entity"
	"github.com/Dhoini/GitHub_Parser/pkg/utils/logger"
	"github.com/google/go-github/v39/github"
)

type GithubServiceImpl struct {
	client *github.Client
	logger *logger.Logger // Замените на ваш логгер
}

func NewGithubService(client *github.Client, logger *logger.Logger) *GithubServiceImpl {
	return &GithubServiceImpl{
		client: client,
		logger: logger, // Замените на ваш логгер
	}
}

func (s *GithubServiceImpl) GetRepository(ctx context.Context, owner, name string) (*entity.Repository, error) {
	repo, _, err := s.client.Repositories.Get(ctx, owner, name)
	if err != nil {
		s.logger.Error("Ошибка получения репозитория: %v", err) // Замените на ваш логгер
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

func (s *GithubServiceImpl) GetIssues(ctx context.Context, owner, repo string, page, perPage int) ([]*entity.Issue, error) {
	opts := &github.IssueListByRepoOptions{
		State:     "all", // Get all issues (open, closed)
		Sort:      "created",
		Direction: "desc",
		ListOptions: github.ListOptions{
			Page:    page,
			PerPage: perPage,
		},
	}

	issues, _, err := s.client.Issues.ListByRepo(ctx, owner, repo, opts)
	if err != nil {
		s.logger.Error("Error getting issues: %v", err)
		return nil, err
	}

	// Получаем ID репозитория
	repository, _, err := s.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		s.logger.Error("Error getting repository for issues: %v", err)
		return nil, err
	}
	repoID := repository.GetID()

	var result []*entity.Issue
	for _, issue := range issues {
		// Пропускаем pull requests, так как у них есть поле PullRequestLinks
		if issue.PullRequestLinks != nil {
			continue
		}

		issueEntity := &entity.Issue{
			ID:           issue.GetID(),
			Number:       issue.GetNumber(),
			Title:        issue.GetTitle(),
			Body:         issue.GetBody(),
			State:        issue.GetState(),
			AuthorLogin:  issue.GetUser().GetLogin(),
			RepositoryID: repoID,
			CreatedAt:    issue.GetCreatedAt().Time,
			UpdatedAt:    issue.GetUpdatedAt().Time,
		}

		if issue.ClosedAt != nil {
			closedAt := issue.GetClosedAt().Time
			issueEntity.ClosedAt = &closedAt
		}

		result = append(result, issueEntity)
	}

	return result, nil
}

func (s *GithubServiceImpl) GetPullRequests(ctx context.Context, owner, repo string, page, perPage int) ([]*entity.PullRequest, error) {
	opts := &github.PullRequestListOptions{
		State:     "all", // Get all PRs (open, closed, merged)
		Sort:      "created",
		Direction: "desc",
		ListOptions: github.ListOptions{
			Page:    page,
			PerPage: perPage,
		},
	}

	prs, _, err := s.client.PullRequests.List(ctx, owner, repo, opts)
	if err != nil {
		s.logger.Error("Error getting pull requests: %v", err)
		return nil, err
	}

	// Получаем ID репозитория
	repository, _, err := s.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		s.logger.Error("Error getting repository for PRs: %v", err)
		return nil, err
	}
	repoID := repository.GetID()

	var result []*entity.PullRequest
	for _, pr := range prs {
		prEntity := &entity.PullRequest{
			ID:           pr.GetID(),
			Number:       pr.GetNumber(),
			Title:        pr.GetTitle(),
			Body:         pr.GetBody(),
			State:        pr.GetState(),
			AuthorLogin:  pr.GetUser().GetLogin(),
			RepositoryID: repoID,
			CreatedAt:    pr.GetCreatedAt().Time,
			UpdatedAt:    pr.GetUpdatedAt().Time,
		}

		if pr.ClosedAt != nil {
			closedAt := pr.GetClosedAt().Time
			prEntity.ClosedAt = &closedAt
		}

		if pr.MergedAt != nil {
			mergedAt := pr.GetMergedAt().Time
			prEntity.MergedAt = &mergedAt
		}

		result = append(result, prEntity)
	}

	return result, nil
}

func (s *GithubServiceImpl) GetUser(ctx context.Context, username string) (*entity.User, error) {
	user, _, err := s.client.Users.Get(ctx, username)
	if err != nil {
		s.logger.Error("Error getting user: %v", err)
		return nil, err
	}

	userEntity := &entity.User{
		ID:        user.GetID(),
		Login:     user.GetLogin(),
		Name:      user.GetName(),
		Email:     user.GetEmail(),
		AvatarURL: user.GetAvatarURL(),
		Bio:       user.GetBio(),
		Company:   user.GetCompany(),
		Location:  user.GetLocation(),
		CreatedAt: user.GetCreatedAt().Time,
		UpdatedAt: user.GetUpdatedAt().Time,
	}

	return userEntity, nil
}

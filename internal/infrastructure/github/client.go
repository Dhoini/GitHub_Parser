package github

import (
	"context"
	"github.com/Dhoini/GitHub_Parser/pkg/utils/RateLimiter"
	"github.com/Dhoini/GitHub_Parser/pkg/utils/logger"
	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

type Client struct {
	client    *github.Client
	rateLimit *RateLimiter.RateLimit
	logger    *logger.Logger
}

func NewGithubClient(token string, logger *logger.Logger) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		client:    github.NewClient(tc),
		rateLimit: RateLimiter.NewRateLimit(logger),
		logger:    logger,
	}
}

// GetClient возвращает клиент GitHub
func (c *Client) GetClient() *github.Client {
	return c.client
}

// GetRateLimit возвращает контроллер ограничения запросов
func (c *Client) GetRateLimit() *RateLimiter.RateLimit {
	return c.rateLimit
}

// GetRepository получает репозиторий с соблюдением ограничений API
func (c *Client) GetRepository(ctx context.Context, owner, repo string) (*github.Repository, error) {
	// Ожидаем, если необходимо, чтобы соблюсти ограничения API
	if err := c.rateLimit.Wait(ctx); err != nil {
		return nil, err
	}

	repository, _, err := c.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		c.logger.Error("Failed to get repository: %v", err)
		return nil, err
	}

	return repository, nil
}

// GetIssues получает issues репозитория с соблюдением ограничений API
func (c *Client) GetIssues(ctx context.Context, owner, repo string, opts *github.IssueListByRepoOptions) ([]*github.Issue, error) {
	// Ожидаем, если необходимо, чтобы соблюсти ограничения API
	if err := c.rateLimit.Wait(ctx); err != nil {
		return nil, err
	}

	issues, _, err := c.client.Issues.ListByRepo(ctx, owner, repo, opts)
	if err != nil {
		c.logger.Error("Failed to get issues: %v", err)
		return nil, err
	}

	return issues, nil
}

// GetPullRequests получает pull requests репозитория с соблюдением ограничений API
func (c *Client) GetPullRequests(ctx context.Context, owner, repo string, opts *github.PullRequestListOptions) ([]*github.PullRequest, error) {
	// Ожидаем, если необходимо, чтобы соблюсти ограничения API
	if err := c.rateLimit.Wait(ctx); err != nil {
		return nil, err
	}

	prs, _, err := c.client.PullRequests.List(ctx, owner, repo, opts)
	if err != nil {
		c.logger.Error("Failed to get pull requests: %v", err)
		return nil, err
	}

	return prs, nil
}

// GetUser получает пользователя с соблюдением ограничений API
func (c *Client) GetUser(ctx context.Context, username string) (*github.User, error) {
	// Ожидаем, если необходимо, чтобы соблюсти ограничения API
	if err := c.rateLimit.Wait(ctx); err != nil {
		return nil, err
	}

	user, _, err := c.client.Users.Get(ctx, username)
	if err != nil {
		c.logger.Error("Failed to get user: %v", err)
		return nil, err
	}

	return user, nil
}

package github

import (
	"context"
	"time"

	"github.com/Dhoini/GitHub_Parser/internal/infrastructure/metrics"
	"github.com/Dhoini/GitHub_Parser/pkg/utils/RateLimiter"
	"github.com/Dhoini/GitHub_Parser/pkg/utils/logger"
	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

type Client struct {
	client    *github.Client
	rateLimit *RateLimiter.RateLimit
	metrics   *metrics.Metrics
	logger    *logger.Logger
}

func NewGithubClient(token string, metrics *metrics.Metrics, logger *logger.Logger) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		client:    github.NewClient(tc),
		rateLimit: RateLimiter.NewRateLimit(logger),
		metrics:   metrics,
		logger:    logger,
	}
}

// GetClient returns the GitHub client
func (c *Client) GetClient() *github.Client {
	return c.client
}

// GetRateLimit returns the rate limit controller
func (c *Client) GetRateLimit() *RateLimiter.RateLimit {
	return c.rateLimit
}

// GetRepository gets a repository with rate limiting
func (c *Client) GetRepository(ctx context.Context, owner, repo string) (*github.Repository, error) {
	// Wait if necessary to comply with API rate limits
	if err := c.rateLimit.Wait(ctx); err != nil {
		return nil, err
	}

	// Record metrics
	c.metrics.APIRequests.WithLabelValues("GetRepository").Inc()
	start := time.Now()
	defer func() {
		c.metrics.APILatency.WithLabelValues("GetRepository").Observe(time.Since(start).Seconds())
	}()

	// Make the API call
	repository, resp, err := c.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		c.logger.Error("Failed to get repository: %v", err)
		c.metrics.Errors.WithLabelValues("GetRepository").Inc()
		return nil, err
	}

	// Update rate limits based on response
	if resp.Rate.Remaining > 0 {
		c.rateLimit.UpdateLimits(resp.Rate.Remaining, resp.Rate.Reset.Time)
	}

	return repository, nil
}

// GetIssues gets repository issues with rate limiting
func (c *Client) GetIssues(ctx context.Context, owner, repo string, opts *github.IssueListByRepoOptions) ([]*github.Issue, error) {
	// Wait if necessary to comply with API rate limits
	if err := c.rateLimit.Wait(ctx); err != nil {
		return nil, err
	}

	// Record metrics
	c.metrics.APIRequests.WithLabelValues("GetIssues").Inc()
	start := time.Now()
	defer func() {
		c.metrics.APILatency.WithLabelValues("GetIssues").Observe(time.Since(start).Seconds())
	}()

	// Make the API call
	issues, resp, err := c.client.Issues.ListByRepo(ctx, owner, repo, opts)
	if err != nil {
		c.logger.Error("Failed to get issues: %v", err)
		c.metrics.Errors.WithLabelValues("GetIssues").Inc()
		return nil, err
	}

	// Update rate limits based on response
	if resp.Rate.Remaining > 0 {
		c.rateLimit.UpdateLimits(resp.Rate.Remaining, resp.Rate.Reset.Time)
	}

	return issues, nil
}

// GetPullRequests gets repository pull requests with rate limiting
func (c *Client) GetPullRequests(ctx context.Context, owner, repo string, opts *github.PullRequestListOptions) ([]*github.PullRequest, error) {
	// Wait if necessary to comply with API rate limits
	if err := c.rateLimit.Wait(ctx); err != nil {
		return nil, err
	}

	// Record metrics
	c.metrics.APIRequests.WithLabelValues("GetPullRequests").Inc()
	start := time.Now()
	defer func() {
		c.metrics.APILatency.WithLabelValues("GetPullRequests").Observe(time.Since(start).Seconds())
	}()

	// Make the API call
	prs, resp, err := c.client.PullRequests.List(ctx, owner, repo, opts)
	if err != nil {
		c.logger.Error("Failed to get pull requests: %v", err)
		c.metrics.Errors.WithLabelValues("GetPullRequests").Inc()
		return nil, err
	}

	// Update rate limits based on response
	if resp.Rate.Remaining > 0 {
		c.rateLimit.UpdateLimits(resp.Rate.Remaining, resp.Rate.Reset.Time)
	}

	return prs, nil
}

// GetUser gets a user with rate limiting
func (c *Client) GetUser(ctx context.Context, username string) (*github.User, error) {
	// Wait if necessary to comply with API rate limits
	if err := c.rateLimit.Wait(ctx); err != nil {
		return nil, err
	}

	// Record metrics
	c.metrics.APIRequests.WithLabelValues("GetUser").Inc()
	start := time.Now()
	defer func() {
		c.metrics.APILatency.WithLabelValues("GetUser").Observe(time.Since(start).Seconds())
	}()

	// Make the API call
	user, resp, err := c.client.Users.Get(ctx, username)
	if err != nil {
		c.logger.Error("Failed to get user: %v", err)
		c.metrics.Errors.WithLabelValues("GetUser").Inc()
		return nil, err
	}

	// Update rate limits based on response
	if resp.Rate.Remaining > 0 {
		c.rateLimit.UpdateLimits(resp.Rate.Remaining, resp.Rate.Reset.Time)
	}

	return user, nil
}

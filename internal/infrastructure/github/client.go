package github

import (
	"context"
	"github.com/google/go-github/v39/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type Client struct {
	client *github.Client
	logger *zap.Logger
}

func NewGithubClient(token string, logger *zap.Logger) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		client: github.NewClient(tc),
		logger: logger,
	}
}

func (c *Client) GetClient() *github.Client {
	return c.client
}

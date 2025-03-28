package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Dhoini/GitHub_Parser/internal/domain/entity"
	"github.com/Dhoini/GitHub_Parser/internal/domain/repository"
	domainService "github.com/Dhoini/GitHub_Parser/internal/domain/service"
	"github.com/Dhoini/GitHub_Parser/internal/infrastructure/metrics"
	"github.com/Dhoini/GitHub_Parser/pkg/utils/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

type JobInfo struct {
	ID           string
	Status       string // "pending", "in_progress", "completed", "failed"
	Progress     int    // 0-100
	ErrorMessage string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Params       domainService.ParsingJobParams
}

type ParserServiceImpl struct {
	githubService domainService.GithubService
	repoRepo      repository.RepositoryRepository
	issueRepo     repository.IssueRepository
	prRepo        repository.PullRequestRepository
	userRepo      repository.UserRepository
	logger        *logger.Logger
	metrics       *metrics.Metrics
	mongoClient   *mongo.Client // Added for transaction support
	// Map for storing active parsing jobs
	jobs map[string]*JobInfo
}

func NewParserService(
	githubService domainService.GithubService,
	repoRepo repository.RepositoryRepository,
	issueRepo repository.IssueRepository,
	prRepo repository.PullRequestRepository,
	userRepo repository.UserRepository,
	mongoClient *mongo.Client,
	metrics *metrics.Metrics,
	logger *logger.Logger,
) *ParserServiceImpl {
	return &ParserServiceImpl{
		githubService: githubService,
		repoRepo:      repoRepo,
		issueRepo:     issueRepo,
		prRepo:        prRepo,
		userRepo:      userRepo,
		mongoClient:   mongoClient,
		metrics:       metrics,
		logger:        logger,
		jobs:          make(map[string]*JobInfo),
	}
}

func (s *ParserServiceImpl) ParseRepository(ctx context.Context, owner, name string) (*entity.Repository, error) {
	repo, err := s.githubService.GetRepository(ctx, owner, name)
	if err != nil {
		s.logger.Error("Failed to get repository from GitHub API: %v", err)
		return nil, err
	}

	if err := s.repoRepo.Save(ctx, repo); err != nil {
		s.logger.Error("Error saving repository: %v", err)
		return nil, err
	}

	// Increment metrics after successful parsing and saving
	if s.metrics != nil {
		s.metrics.ParsedRepositories.Inc()
		s.metrics.DBOperations.WithLabelValues("save", "repository").Inc()
	}

	return repo, nil
}

func (s *ParserServiceImpl) ParseIssues(ctx context.Context, owner, repo string) ([]*entity.Issue, error) {
	// Get repository to ensure it exists and we have its ID
	repository, err := s.githubService.GetRepository(ctx, owner, repo)
	if err != nil {
		s.logger.Error("Failed to get repository for issues parsing: %v", err)
		return nil, err
	}

	// Get issues from GitHub API (first page, 100 issues)
	issues, err := s.githubService.GetIssues(ctx, owner, repo, 1, 100)
	if err != nil {
		s.logger.Error("Failed to get issues from GitHub API: %v", err)
		return nil, err
	}

	// Save each issue to the database
	for _, issue := range issues {
		// Make sure the issue is linked to the correct repository
		issue.RepositoryID = repository.ID

		if err := s.issueRepo.Save(ctx, issue); err != nil {
			s.logger.Error("Error saving issue #%d: %v", issue.Number, err)
			// Continue even if there's an error saving one issue
		}
	}

	// Increment metrics
	if s.metrics != nil {
		s.metrics.ParsedIssues.Add(float64(len(issues)))
		s.metrics.DBOperations.WithLabelValues("save", "issue").Add(float64(len(issues)))
	}

	return issues, nil
}

func (s *ParserServiceImpl) ParsePullRequests(ctx context.Context, owner, repo string) ([]*entity.PullRequest, error) {
	// Get repository to ensure it exists and we have its ID
	repository, err := s.githubService.GetRepository(ctx, owner, repo)
	if err != nil {
		s.logger.Error("Failed to get repository for PR parsing: %v", err)
		return nil, err
	}

	// Get PRs from GitHub API (first page, 100 PRs)
	prs, err := s.githubService.GetPullRequests(ctx, owner, repo, 1, 100)
	if err != nil {
		s.logger.Error("Failed to get pull requests from GitHub API: %v", err)
		return nil, err
	}

	// Save each PR to the database
	for _, pr := range prs {
		// Make sure the PR is linked to the correct repository
		pr.RepositoryID = repository.ID

		if err := s.prRepo.Save(ctx, pr); err != nil {
			s.logger.Error("Error saving PR #%d: %v", pr.Number, err)
			// Continue even if there's an error saving one PR
		}
	}

	// Increment metrics
	if s.metrics != nil {
		s.metrics.ParsedPullRequests.Add(float64(len(prs)))
		s.metrics.DBOperations.WithLabelValues("save", "pull_request").Add(float64(len(prs)))
	}

	return prs, nil
}

func (s *ParserServiceImpl) ParseUser(ctx context.Context, username string) (*entity.User, error) {
	// Get user from GitHub API
	user, err := s.githubService.GetUser(ctx, username)
	if err != nil {
		s.logger.Error("Failed to get user from GitHub API: %v", err)
		return nil, err
	}

	// Save user to the database
	if err := s.userRepo.Save(ctx, user); err != nil {
		s.logger.Error("Error saving user %s: %v", username, err)
		return nil, err
	}

	// Increment metrics
	if s.metrics != nil {
		s.metrics.ParsedUsers.Inc()
		s.metrics.DBOperations.WithLabelValues("save", "user").Inc()
	}

	return user, nil
}

func (s *ParserServiceImpl) StartParsingJob(ctx context.Context, params domainService.ParsingJobParams) (string, error) {
	// Generate a unique job ID
	jobID := fmt.Sprintf("job-%d", time.Now().Unix())

	// Create job info
	job := &JobInfo{
		ID:        jobID,
		Status:    "pending",
		Progress:  0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Params:    params,
	}

	// Save the job
	s.jobs[jobID] = job

	// Start the job in a separate goroutine
	go s.processParsingJob(context.Background(), jobID)

	// Increment job metrics
	if s.metrics != nil {
		s.metrics.ParsingJobs.WithLabelValues("pending").Inc()
		s.metrics.ParsingJobsTotal.Inc()
	}

	return jobID, nil
}

func (s *ParserServiceImpl) GetParsingJobStatus(ctx context.Context, jobID string) (*domainService.ParsingJobStatus, error) {
	job, exists := s.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	return &domainService.ParsingJobStatus{
		ID:           job.ID,
		Status:       job.Status,
		Progress:     job.Progress,
		ErrorMessage: job.ErrorMessage,
		CreatedAt:    job.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    job.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *ParserServiceImpl) processParsingJob(ctx context.Context, jobID string) {
	job, exists := s.jobs[jobID]
	if !exists {
		s.logger.Error("Job not found: %s", jobID)
		return
	}

	// Update job status
	job.Status = "in_progress"
	job.UpdatedAt = time.Now()

	// Update metrics
	if s.metrics != nil {
		s.metrics.ParsingJobs.WithLabelValues("pending").Dec()
		s.metrics.ParsingJobs.WithLabelValues("in_progress").Inc()
	}

	// Create a context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// Start with parsing the repository
	repo, err := s.ParseRepository(timeoutCtx, job.Params.OwnerName, job.Params.RepoName)
	if err != nil {
		job.Status = "failed"
		job.ErrorMessage = fmt.Sprintf("failed to parse repository: %v", err)
		job.UpdatedAt = time.Now()
		s.logger.Error("Job %s failed: %v", jobID, err)

		// Update metrics
		if s.metrics != nil {
			s.metrics.ParsingJobs.WithLabelValues("in_progress").Dec()
			s.metrics.ParsingJobs.WithLabelValues("failed").Inc()
			s.metrics.ParsingJobsErrors.Inc()
		}

		return
	}

	job.Progress = 20
	job.UpdatedAt = time.Now()

	// If we need to parse issues
	if job.Params.ParseIssues {
		_, err := s.ParseIssues(timeoutCtx, job.Params.OwnerName, job.Params.RepoName)
		if err != nil {
			job.Status = "failed"
			job.ErrorMessage = fmt.Sprintf("failed to parse issues: %v", err)
			job.UpdatedAt = time.Now()
			s.logger.Error("Job %s failed at issues parsing: %v", jobID, err)

			// Update metrics
			if s.metrics != nil {
				s.metrics.ParsingJobs.WithLabelValues("in_progress").Dec()
				s.metrics.ParsingJobs.WithLabelValues("failed").Inc()
				s.metrics.ParsingJobsErrors.Inc()
			}

			return
		}

		job.Progress = 50
		job.UpdatedAt = time.Now()
	}

	// If we need to parse pull requests
	if job.Params.ParsePRs {
		_, err := s.ParsePullRequests(timeoutCtx, job.Params.OwnerName, job.Params.RepoName)
		if err != nil {
			job.Status = "failed"
			job.ErrorMessage = fmt.Sprintf("failed to parse pull requests: %v", err)
			job.UpdatedAt = time.Now()
			s.logger.Error("Job %s failed at PRs parsing: %v", jobID, err)

			// Update metrics
			if s.metrics != nil {
				s.metrics.ParsingJobs.WithLabelValues("in_progress").Dec()
				s.metrics.ParsingJobs.WithLabelValues("failed").Inc()
				s.metrics.ParsingJobsErrors.Inc()
			}

			return
		}

		job.Progress = 80
		job.UpdatedAt = time.Now()
	}

	// If we need to parse users
	if job.Params.ParseUsers {
		// Here we just parse the repository owner
		// In a real application, you might want to parse other contributors as well
		_, err := s.ParseUser(timeoutCtx, repo.OwnerLogin)
		if err != nil {
			job.Status = "failed"
			job.ErrorMessage = fmt.Sprintf("failed to parse owner: %v", err)
			job.UpdatedAt = time.Now()
			s.logger.Error("Job %s failed at owner parsing: %v", jobID, err)

			// Update metrics
			if s.metrics != nil {
				s.metrics.ParsingJobs.WithLabelValues("in_progress").Dec()
				s.metrics.ParsingJobs.WithLabelValues("failed").Inc()
				s.metrics.ParsingJobsErrors.Inc()
			}

			return
		}
	}

	// Job completed successfully
	job.Status = "completed"
	job.Progress = 100
	job.UpdatedAt = time.Now()

	// Update metrics
	if s.metrics != nil {
		s.metrics.ParsingJobs.WithLabelValues("in_progress").Dec()
		s.metrics.ParsingJobs.WithLabelValues("completed").Inc()
	}

	s.logger.Info("Job %s completed successfully", jobID)
}

func (s *ParserServiceImpl) ParseRepositoryWithDetails(ctx context.Context, owner, name string, parseIssues, parsePRs bool) (*entity.Repository, error) {
	if s.mongoClient == nil {
		return nil, fmt.Errorf("mongo client is not initialized")
	}

	// Start a transaction
	session, err := s.mongoClient.StartSession()
	if err != nil {
		s.logger.Error("Failed to start MongoDB session: %v", err)
		return nil, err
	}
	defer session.EndSession(ctx)

	var repo *entity.Repository

	err = mongo.WithSession(ctx, session, func(sessCtx mongo.SessionContext) error {
		if err := session.StartTransaction(); err != nil {
			return err
		}

		// Get and save repository
		repo, err = s.githubService.GetRepository(sessCtx, owner, name)
		if err != nil {
			return err
		}

		if err := s.repoRepo.Save(sessCtx, repo); err != nil {
			return err
		}

		// If we need to parse issues
		if parseIssues {
			issues, err := s.githubService.GetIssues(sessCtx, owner, name, 1, 100)
			if err != nil {
				return err
			}

			for _, issue := range issues {
				issue.RepositoryID = repo.ID
				if err := s.issueRepo.Save(sessCtx, issue); err != nil {
					return err
				}
			}

			// Update metrics
			if s.metrics != nil {
				s.metrics.ParsedIssues.Add(float64(len(issues)))
			}
		}

		// Similarly for PRs
		if parsePRs {
			prs, err := s.githubService.GetPullRequests(sessCtx, owner, name, 1, 100)
			if err != nil {
				return err
			}

			for _, pr := range prs {
				pr.RepositoryID = repo.ID
				if err := s.prRepo.Save(sessCtx, pr); err != nil {
					return err
				}
			}

			// Update metrics
			if s.metrics != nil {
				s.metrics.ParsedPullRequests.Add(float64(len(prs)))
			}
		}

		// Commit the transaction
		if err := session.CommitTransaction(sessCtx); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Transaction failed: %v", err)
		return nil, err
	}

	// Update metrics
	if s.metrics != nil {
		s.metrics.ParsedRepositories.Inc()
	}

	return repo, nil
}

package service

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"time"

	"github.com/Dhoini/GitHub_Parser/internal/domain/entity"
	"github.com/Dhoini/GitHub_Parser/internal/domain/repository"
	domainService "github.com/Dhoini/GitHub_Parser/internal/domain/service"
	"github.com/Dhoini/GitHub_Parser/pkg/utils/logger" // Замените на ваш логгер
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
	logger        *logger.Logger // Замените на ваш логгер
	// Map для хранения активных задач парсинга
	jobs map[string]*JobInfo
}

func NewParserService(
	githubService domainService.GithubService,
	repoRepo repository.RepositoryRepository,
	issueRepo repository.IssueRepository,
	prRepo repository.PullRequestRepository,
	userRepo repository.UserRepository,
	logger *logger.Logger, // Замените на ваш логгер
) *ParserServiceImpl {
	return &ParserServiceImpl{
		githubService: githubService,
		repoRepo:      repoRepo,
		issueRepo:     issueRepo,
		prRepo:        prRepo,
		userRepo:      userRepo,
		logger:        logger, // Замените на ваш логгер
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

	// Инкрементируем метрику после успешного парсинга и сохранения
	s.metrics.ParsedRepositories.Inc()
	s.metrics.DBOperations.WithLabelValues("save", "repository").Inc()

	return repo, nil
}

// Остальные методы аналогично...

func (s *ParserServiceImpl) ParseIssues(ctx context.Context, owner, repo string) ([]*entity.Issue, error) {
	// Получаем репозиторий, чтобы убедиться, что он существует и у нас есть его ID
	repository, err := s.githubService.GetRepository(ctx, owner, repo)
	if err != nil {
		s.logger.Error("Failed to get repository for issues parsing: %v", err)
		return nil, err
	}

	// Получаем issues с GitHub API (первая страница, 100 issues)
	issues, err := s.githubService.GetIssues(ctx, owner, repo, 1, 100)
	if err != nil {
		s.logger.Error("Failed to get issues from GitHub API: %v", err)
		return nil, err
	}

	// Сохраняем каждый issue в БД
	for _, issue := range issues {
		// Убедимся, что issue привязан к правильному репозиторию
		issue.RepositoryID = repository.ID

		if err := s.issueRepo.Save(ctx, issue); err != nil {
			s.logger.Error("Error saving issue #%d: %v", issue.Number, err)
			// Продолжаем даже при ошибке сохранения одного issue
		}
	}

	return issues, nil
}

func (s *ParserServiceImpl) ParsePullRequests(ctx context.Context, owner, repo string) ([]*entity.PullRequest, error) {
	// Получаем репозиторий, чтобы убедиться, что он существует и у нас есть его ID
	repository, err := s.githubService.GetRepository(ctx, owner, repo)
	if err != nil {
		s.logger.Error("Failed to get repository for PR parsing: %v", err)
		return nil, err
	}

	// Получаем PRs с GitHub API (первая страница, 100 PRs)
	prs, err := s.githubService.GetPullRequests(ctx, owner, repo, 1, 100)
	if err != nil {
		s.logger.Error("Failed to get pull requests from GitHub API: %v", err)
		return nil, err
	}

	// Сохраняем каждый PR в БД
	for _, pr := range prs {
		// Убедимся, что PR привязан к правильному репозиторию
		pr.RepositoryID = repository.ID

		if err := s.prRepo.Save(ctx, pr); err != nil {
			s.logger.Error("Error saving PR #%d: %v", pr.Number, err)
			// Продолжаем даже при ошибке сохранения одного PR
		}
	}

	return prs, nil
}

func (s *ParserServiceImpl) ParseUser(ctx context.Context, username string) (*entity.User, error) {
	// Получаем пользователя из GitHub API
	user, err := s.githubService.GetUser(ctx, username)
	if err != nil {
		s.logger.Error("Failed to get user from GitHub API: %v", err)
		return nil, err
	}

	// Сохраняем пользователя в БД
	if err := s.userRepo.Save(ctx, user); err != nil {
		s.logger.Error("Error saving user %s: %v", username, err)
		return nil, err
	}

	return user, nil
}

func (s *ParserServiceImpl) StartParsingJob(ctx context.Context, params domainService.ParsingJobParams) (string, error) {
	// Генерируем уникальный идентификатор задачи
	jobID := fmt.Sprintf("job-%d", time.Now().Unix())

	// Создаем информацию о задаче
	job := &JobInfo{
		ID:        jobID,
		Status:    "pending",
		Progress:  0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Params:    params,
	}

	// Сохраняем задачу
	s.jobs[jobID] = job

	// Запускаем задачу в отдельной горутине
	go s.processParsingJob(ctx, jobID)

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

	// Обновляем статус задачи
	job.Status = "in_progress"
	job.UpdatedAt = time.Now()

	// Создаем контекст с таймаутом
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Начинаем с парсинга репозитория
	repo, err := s.ParseRepository(timeoutCtx, job.Params.OwnerName, job.Params.RepoName)
	if err != nil {
		job.Status = "failed"
		job.ErrorMessage = fmt.Sprintf("failed to parse repository: %v", err)
		job.UpdatedAt = time.Now()
		s.logger.Error("Job %s failed: %v", jobID, err)
		return
	}

	job.Progress = 20
	job.UpdatedAt = time.Now()

	// Если нужно парсить issues
	if job.Params.ParseIssues {
		_, err := s.ParseIssues(timeoutCtx, job.Params.OwnerName, job.Params.RepoName)
		if err != nil {
			job.Status = "failed"
			job.ErrorMessage = fmt.Sprintf("failed to parse issues: %v", err)
			job.UpdatedAt = time.Now()
			s.logger.Error("Job %s failed at issues parsing: %v", jobID, err)
			return
		}

		job.Progress = 50
		job.UpdatedAt = time.Now()
	}

	// Если нужно парсить pull requests
	if job.Params.ParsePRs {
		_, err := s.ParsePullRequests(timeoutCtx, job.Params.OwnerName, job.Params.RepoName)
		if err != nil {
			job.Status = "failed"
			job.ErrorMessage = fmt.Sprintf("failed to parse pull requests: %v", err)
			job.UpdatedAt = time.Now()
			s.logger.Error("Job %s failed at PRs parsing: %v", jobID, err)
			return
		}

		job.Progress = 80
		job.UpdatedAt = time.Now()
	}

	// Если нужно парсить пользователей
	if job.Params.ParseUsers {
		// Здесь мы парсим только владельца репозитория
		// В реальном приложении можно парсить и других участников
		_, err := s.ParseUser(timeoutCtx, repo.OwnerLogin)
		if err != nil {
			job.Status = "failed"
			job.ErrorMessage = fmt.Sprintf("failed to parse owner: %v", err)
			job.UpdatedAt = time.Now()
			s.logger.Error("Job %s failed at owner parsing: %v", jobID, err)
			return
		}
	}

	// Задача успешно завершена
	job.Status = "completed"
	job.Progress = 100
	job.UpdatedAt = time.Now()
	s.logger.Info("Job %s completed successfully", jobID)
}

func (s *ParserServiceImpl) ParseRepositoryWithDetails(ctx context.Context, owner, name string, parseIssues, parsePRs bool) (*entity.Repository, error) {
	// Запускаем транзакцию
	session, err := s.mongoClient.StartSession()
	if err != nil {
		s.logger.Error("Failed to start MongoDB session: %v", err)
		return nil, err
	}
	defer session.EndSession(ctx)

	var repo *entity.Repository

	err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) error {
		// Получаем и сохраняем репозиторий
		repo, err = s.githubService.GetRepository(sessCtx, owner, name)
		if err != nil {
			return err
		}

		if err := s.repoRepo.Save(sessCtx, repo); err != nil {
			return err
		}

		// Если нужно парсить issues
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

			// Обновляем метрики
			s.metrics.ParsedIssues.Add(float64(len(issues)))
		}

		// Аналогично для PRs
		if parsePRs {
			// Реализация
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Обновляем метрики
	s.metrics.ParsedRepositories.Inc()

	return repo, nil
}

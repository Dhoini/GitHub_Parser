package grpc

import (
	"context"
	"time"

	"github.com/Dhoini/GitHub_Parser/internal/domain/repository"
	"github.com/Dhoini/GitHub_Parser/internal/domain/service"
	pb "github.com/Dhoini/GitHub_Parser/internal/infrastructure/api/proto"
	"github.com/Dhoini/GitHub_Parser/pkg/utils/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	pb.UnimplementedGithubParserServiceServer
	parserService service.ParserService
	logger        *logger.Logger
}

func NewHandler(parserService service.ParserService, logger *logger.Logger) *Handler {
	return &Handler{
		parserService: parserService,
		logger:        logger,
	}
}

// ParseRepository парсит репозиторий GitHub
func (h *Handler) ParseRepository(ctx context.Context, req *pb.ParseRepositoryRequest) (*pb.ParseRepositoryResponse, error) {
	if req.Owner == "" || req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "owner and name are required")
	}

	repo, err := h.parserService.ParseRepository(ctx, req.Owner, req.Name)
	if err != nil {
		h.logger.Error("Failed to parse repository: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to parse repository: %v", err)
	}

	return &pb.ParseRepositoryResponse{
		Repository: &pb.Repository{
			Id:              repo.ID,
			Name:            repo.Name,
			FullName:        repo.FullName,
			Description:     repo.Description,
			IsPrivate:       repo.IsPrivate,
			OwnerLogin:      repo.OwnerLogin,
			Language:        repo.Language,
			StarsCount:      int32(repo.StarsCount),
			ForksCount:      int32(repo.ForksCount),
			OpenIssuesCount: int32(repo.OpenIssuesCount),
			CreatedAt:       repo.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       repo.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// ListRepositories возвращает список репозиториев
func (h *Handler) ListRepositories(ctx context.Context, req *pb.ListRepositoriesRequest) (*pb.ListRepositoriesResponse, error) {
	filter := repository.RepositoryFilter{
		OwnerLogin: req.OwnerLogin,
		Language:   req.Language,
		MinStars:   int(req.MinStars),
		Limit:      int(req.Limit),
		Offset:     int(req.Offset),
	}

	// Здесь требуется реализация метода List репозитория
	// Для примера заглушка
	h.logger.Error("ListRepositories method not implemented")
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

// ParseIssues парсит issues репозитория
func (h *Handler) ParseIssues(ctx context.Context, req *pb.ParseIssuesRequest) (*pb.ParseIssuesResponse, error) {
	if req.Owner == "" || req.Repo == "" {
		return nil, status.Errorf(codes.InvalidArgument, "owner and repo are required")
	}

	issues, err := h.parserService.ParseIssues(ctx, req.Owner, req.Repo)
	if err != nil {
		h.logger.Error("Failed to parse issues: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to parse issues: %v", err)
	}

	var pbIssues []*pb.Issue
	for _, issue := range issues {
		pbIssue := &pb.Issue{
			Id:           issue.ID,
			Number:       int32(issue.Number),
			Title:        issue.Title,
			Body:         issue.Body,
			State:        issue.State,
			AuthorLogin:  issue.AuthorLogin,
			RepositoryId: issue.RepositoryID,
			CreatedAt:    issue.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    issue.UpdatedAt.Format(time.RFC3339),
		}

		if issue.ClosedAt != nil {
			pbIssue.ClosedAt = issue.ClosedAt.Format(time.RFC3339)
		}

		pbIssues = append(pbIssues, pbIssue)
	}

	return &pb.ParseIssuesResponse{
		Issues: pbIssues,
	}, nil
}

// ListIssues возвращает список issues
func (h *Handler) ListIssues(ctx context.Context, req *pb.ListIssuesRequest) (*pb.ListIssuesResponse, error) {
	filter := repository.IssueFilter{
		RepositoryID: req.RepositoryId,
		State:        req.State,
		Limit:        int(req.Limit),
		Offset:       int(req.Offset),
	}

	// Здесь требуется реализация метода List репозитория
	// Для примера заглушка
	h.logger.Error("ListIssues method not implemented")
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

// ParsePullRequests парсит pull requests репозитория
func (h *Handler) ParsePullRequests(ctx context.Context, req *pb.ParsePullRequestsRequest) (*pb.ParsePullRequestsResponse, error) {
	if req.Owner == "" || req.Repo == "" {
		return nil, status.Errorf(codes.InvalidArgument, "owner and repo are required")
	}

	prs, err := h.parserService.ParsePullRequests(ctx, req.Owner, req.Repo)
	if err != nil {
		h.logger.Error("Failed to parse pull requests: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to parse pull requests: %v", err)
	}

	var pbPRs []*pb.PullRequest
	for _, pr := range prs {
		pbPR := &pb.PullRequest{
			Id:           pr.ID,
			Number:       int32(pr.Number),
			Title:        pr.Title,
			Body:         pr.Body,
			State:        pr.State,
			AuthorLogin:  pr.AuthorLogin,
			RepositoryId: pr.RepositoryID,
			CreatedAt:    pr.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    pr.UpdatedAt.Format(time.RFC3339),
		}

		if pr.MergedAt != nil {
			pbPR.MergedAt = pr.MergedAt.Format(time.RFC3339)
		}

		if pr.ClosedAt != nil {
			pbPR.ClosedAt = pr.ClosedAt.Format(time.RFC3339)
		}

		pbPRs = append(pbPRs, pbPR)
	}

	return &pb.ParsePullRequestsResponse{
		PullRequests: pbPRs,
	}, nil
}

// ListPullRequests возвращает список pull requests
func (h *Handler) ListPullRequests(ctx context.Context, req *pb.ListPullRequestsRequest) (*pb.ListPullRequestsResponse, error) {
	filter := repository.PullRequestFilter{
		RepositoryID: req.RepositoryId,
		State:        req.State,
		Limit:        int(req.Limit),
		Offset:       int(req.Offset),
	}

	// Здесь требуется реализация метода List репозитория
	// Для примера заглушка
	h.logger.Error("ListPullRequests method not implemented")
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

// ParseUser парсит пользователя GitHub
func (h *Handler) ParseUser(ctx context.Context, req *pb.ParseUserRequest) (*pb.ParseUserResponse, error) {
	if req.Username == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username is required")
	}

	user, err := h.parserService.ParseUser(ctx, req.Username)
	if err != nil {
		h.logger.Error("Failed to parse user: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to parse user: %v", err)
	}

	return &pb.ParseUserResponse{
		User: &pb.User{
			Id:        user.ID,
			Login:     user.Login,
			Name:      user.Name,
			Email:     user.Email,
			AvatarUrl: user.AvatarURL,
			Bio:       user.Bio,
			Company:   user.Company,
			Location:  user.Location,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
			UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// ListUsers возвращает список пользователей
func (h *Handler) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	filter := repository.UserFilter{
		Login:  req.Login,
		Limit:  int(req.Limit),
		Offset: int(req.Offset),
	}

	// Здесь требуется реализация метода List репозитория
	// Для примера заглушка
	h.logger.Error("ListUsers method not implemented")
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

// StartParsingJob запускает асинхронную задачу парсинга
func (h *Handler) StartParsingJob(ctx context.Context, req *pb.StartParsingJobRequest) (*pb.StartParsingJobResponse, error) {
	if req.OwnerName == "" || req.RepoName == "" {
		return nil, status.Errorf(codes.InvalidArgument, "owner_name and repo_name are required")
	}

	params := service.ParsingJobParams{
		OwnerName:   req.OwnerName,
		RepoName:    req.RepoName,
		ParseIssues: req.ParseIssues,
		ParsePRs:    req.ParsePullRequests,
		ParseUsers:  req.ParseUsers,
	}

	jobID, err := h.parserService.StartParsingJob(ctx, params)
	if err != nil {
		h.logger.Error("Failed to start parsing job: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to start parsing job: %v", err)
	}

	return &pb.StartParsingJobResponse{
		JobId: jobID,
	}, nil
}

// GetParsingJobStatus возвращает статус асинхронной задачи парсинга
func (h *Handler) GetParsingJobStatus(ctx context.Context, req *pb.GetParsingJobStatusRequest) (*pb.GetParsingJobStatusResponse, error) {
	if req.JobId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "job_id is required")
	}

	jobStatus, err := h.parserService.GetParsingJobStatus(ctx, req.JobId)
	if err != nil {
		h.logger.Error("Failed to get parsing job status: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get parsing job status: %v", err)
	}

	return &pb.GetParsingJobStatusResponse{
		Id:           jobStatus.ID,
		Status:       jobStatus.Status,
		Progress:     int32(jobStatus.Progress),
		ErrorMessage: jobStatus.ErrorMessage,
		CreatedAt:    jobStatus.CreatedAt,
		UpdatedAt:    jobStatus.UpdatedAt,
	}, nil
}

package grpc

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Dhoini/GitHub_Parser/internal/domain/service"
	pb "github.com/Dhoini/GitHub_Parser/internal/infrastructure/api/proto"
)

type Handler struct {
	pb.UnimplementedGithubParserServiceServer
	parserService service.ParserService
	logger        *zap.Logger
}

func NewHandler(parserService service.ParserService, logger *zap.Logger) *Handler {
	return &Handler{
		parserService: parserService,
		logger:        logger,
	}
}

func (h *Handler) ParseRepository(ctx context.Context, req *pb.ParseRepositoryRequest) (*pb.ParseRepositoryResponse, error) {
	if req.Owner == "" || req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "owner and name are required")
	}

	repo, err := h.parserService.ParseRepository(ctx, req.Owner, req.Name)
	if err != nil {
		h.logger.Error("Failed to parse repository",
			zap.String("owner", req.Owner),
			zap.String("name", req.Name),
			zap.Error(err))
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

// Аналогично реализуйте остальные методы обработчика

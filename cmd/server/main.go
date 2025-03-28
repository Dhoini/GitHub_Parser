package main

import (
	"fmt"
	"github.com/Dhoini/GitHub_Parser/internal/application/service"
	"github.com/Dhoini/GitHub_Parser/internal/config"
	"github.com/Dhoini/GitHub_Parser/internal/infrastructure/github"
	"github.com/Dhoini/GitHub_Parser/internal/infrastructure/persistence/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Инициализация логгера
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// Подключение к MongoDB
	mongoClient, err := connectMongoDB(cfg.MongoDB.URI)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}

	db := mongoClient.Database(cfg.MongoDB.Database)

	// Инициализация репозиториев
	repoRepo := mongodb.NewRepositoryRepository(db, logger)
	issueRepo := mongodb.NewIssueRepository(db, logger)
	prRepo := mongodb.NewPullRequestRepository(db, logger)
	userRepo := mongodb.NewUserRepository(db, logger)

	// Инициализация клиента GitHub
	githubClient := github.NewGithubClient(cfg.GitHub.Token, logger)

	// Инициализация сервисов
	githubService := service.NewGithubService(githubClient.GetClient(), logger)
	parserService := service.NewParserService(githubService, repoRepo, issueRepo, prRepo, userRepo, logger)

	// Инициализация gRPC сервера
	server := grpc.NewServer()
	handler := grpcHandler.NewHandler(parserService, logger)
	pb.RegisterGithubParserServiceServer(server, handler)

	// Запуск gRPC сервера
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	go func() {
		logger.Info("Starting gRPC server", zap.Int("port", cfg.Server.Port))
		if err := server.Serve(lis); err != nil {
			logger.Fatal("Failed to serve", zap.Error(err))
		}
	}()

	// Настройка graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	server.GracefulStop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := mongoClient.Disconnect(ctx); err != nil {
		logger.Error("Error disconnecting from MongoDB", zap.Error(err))
	}

	logger.Info("Server stopped")
}

func connectMongoDB(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

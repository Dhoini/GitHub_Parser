package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Dhoini/GitHub_Parser/internal/application/service"
	"github.com/Dhoini/GitHub_Parser/internal/config"
	"github.com/Dhoini/GitHub_Parser/internal/infrastructure/api/proto"
	"github.com/Dhoini/GitHub_Parser/internal/infrastructure/github"
	"github.com/Dhoini/GitHub_Parser/internal/infrastructure/metrics"
	"github.com/Dhoini/GitHub_Parser/internal/infrastructure/persistence/mongodb"
	grpcHandler "github.com/Dhoini/GitHub_Parser/internal/interfaces/grpc"
	"github.com/Dhoini/GitHub_Parser/pkg/utils/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

func main() {
	// Инициализация логгера
	customLogger := logger.New(logger.DEBUG)
	customLogger.Info("Starting GitHub Parser service")

	// Инициализация метрик
	appMetrics := metrics.NewMetrics(customLogger)
	metrics.StartMetricsServer(":9090", customLogger)
	customLogger.Info("Metrics server started on :9090")

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		customLogger.Fatal("Failed to load config: %v", err)
	}

	// Подключение к MongoDB
	mongoClient, err := connectMongoDB(cfg.MongoDB.URI)
	if err != nil {
		customLogger.Fatal("Failed to connect to MongoDB: %v", err)
	}

	// Обновляем метрику открытых соединений
	appMetrics.DBConnectionsOpen.Set(float64(mongoClient.NumberSessionsInProgress()))

	db := mongoClient.Database(cfg.MongoDB.Database)

	// Инициализация репозиториев
	repoRepo := mongodb.NewRepositoryRepository(db, customLogger)
	issueRepo := mongodb.NewIssueRepository(db, customLogger)
	prRepo := mongodb.NewPullRequestRepository(db, customLogger)
	userRepo := mongodb.NewUserRepository(db, customLogger)

	// Инициализация клиента GitHub
	githubClient := github.NewGithubClient(cfg.GitHub.Token, appMetrics, customLogger)

	// Инициализация сервисов
	githubService := service.NewGithubService(githubClient.GetClient(), customLogger)
	parserService := service.NewParserService(githubService, repoRepo, issueRepo, prRepo, userRepo, customLogger)

	// Инициализация gRPC сервера
	server := grpc.NewServer()
	handler := grpcHandler.NewHandler(
		parserService,
		repoRepo,
		issueRepo,
		prRepo,
		userRepo,
		customLogger,
	)
	proto.RegisterGithubParserServiceServer(server, handler)

	// Запуск gRPC сервера
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		customLogger.Fatal("Failed to listen: %v", err)
	}

	go func() {
		customLogger.Info("Starting gRPC server on port %d", cfg.Server.Port)
		if err := server.Serve(lis); err != nil {
			customLogger.Fatal("Failed to serve: %v", err)
		}
	}()

	// Настройка graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	customLogger.Info("Shutting down server...")

	server.GracefulStop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := mongoClient.Disconnect(ctx); err != nil {
		customLogger.Error("Error disconnecting from MongoDB: %v", err)
	}

	customLogger.Info("Server stopped")
}

func connectMongoDB(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)

	// Add connection pooling settings
	clientOptions.SetMinPoolSize(5)
	clientOptions.SetMaxPoolSize(100)
	clientOptions.SetMaxConnIdleTime(30 * time.Minute)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

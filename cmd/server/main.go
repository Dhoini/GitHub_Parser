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
	// Initialize logger
	customLogger := logger.New(logger.DEBUG)
	customLogger.Info("Starting GitHub Parser service")

	// Initialize metrics
	appMetrics := metrics.NewMetrics(customLogger)
	metrics.StartMetricsServer(":9090", customLogger)
	customLogger.Info("Metrics server started on :9090")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		customLogger.Fatal("Failed to load config: %v", err)
	}

	// Connect to MongoDB
	mongoClient, err := connectMongoDB(cfg.MongoDB.URI)
	if err != nil {
		customLogger.Fatal("Failed to connect to MongoDB: %v", err)
	}

	// Update metrics for open connections
	appMetrics.DBConnectionsOpen.Set(float64(mongoClient.NumberSessionsInProgress()))

	db := mongoClient.Database(cfg.MongoDB.Database)

	// Initialize repositories
	repoRepo := mongodb.NewRepositoryRepository(db, customLogger)
	issueRepo := mongodb.NewIssueRepository(db, customLogger)
	prRepo := mongodb.NewPullRequestRepository(db, customLogger)
	userRepo := mongodb.NewUserRepository(db, customLogger)

	// Initialize GitHub client
	githubClient := github.NewGithubClient(cfg.GitHub.Token, appMetrics, customLogger)

	// Initialize services
	githubService := service.NewGithubService(githubClient.GetClient(), customLogger)
	parserService := service.NewParserService(
		githubService,
		repoRepo,
		issueRepo,
		prRepo,
		userRepo,
		mongoClient, // Pass mongo client for transaction support
		appMetrics,  // Pass metrics
		customLogger,
	)

	// Initialize gRPC server
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

	// Start gRPC server
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

	// Setup graceful shutdown
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

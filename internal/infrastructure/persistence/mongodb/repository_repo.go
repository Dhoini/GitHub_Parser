package mongodb

import (
	"context"
	"github.com/Dhoini/GitHub_Parser/internal/domain/entity"
	"github.com/Dhoini/GitHub_Parser/internal/domain/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type RepositoryRepositoryMongo struct {
	collection *mongo.Collection
	logger     *zap.Logger
}

func NewRepositoryRepository(db *mongo.Database, logger *zap.Logger) repository.RepositoryRepository {
	return &RepositoryRepositoryMongo{
		collection: db.Collection("repositories"),
		logger:     logger,
	}
}

func (r *RepositoryRepositoryMongo) Save(ctx context.Context, repo *entity.Repository) error {
	filter := bson.M{"id": repo.ID}
	update := bson.M{"$set": bson.M{
		"id":              repo.ID,
		"name":            repo.Name,
		"fullName":        repo.FullName,
		"description":     repo.Description,
		"isPrivate":       repo.IsPrivate,
		"ownerLogin":      repo.OwnerLogin,
		"language":        repo.Language,
		"starsCount":      repo.StarsCount,
		"forksCount":      repo.ForksCount,
		"openIssuesCount": repo.OpenIssuesCount,
		"createdAt":       repo.CreatedAt,
		"updatedAt":       repo.UpdatedAt,
	}}

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// Аналогично реализуйте остальные методы интерфейса

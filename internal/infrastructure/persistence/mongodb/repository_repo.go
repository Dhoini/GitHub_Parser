package mongodb

import (
	"context"
	"github.com/Dhoini/GitHub_Parser/internal/domain/entity"
	"github.com/Dhoini/GitHub_Parser/internal/domain/repository"
	"github.com/Dhoini/GitHub_Parser/pkg/utils/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RepositoryRepositoryMongo struct {
	collection *mongo.Collection
	logger     *logger.Logger
}

func NewRepositoryRepository(db *mongo.Database, logger *logger.Logger) repository.RepositoryRepository {
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
	if err != nil {
		r.logger.Error("Failed to save repository: %v", err)
		return err
	}

	return nil
}

func (r *RepositoryRepositoryMongo) FindByID(ctx context.Context, id int64) (*entity.Repository, error) {
	filter := bson.M{"id": id}

	var repo entity.Repository
	err := r.collection.FindOne(ctx, filter).Decode(&repo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			r.logger.Info("Repository not found: %d", id)
			return nil, nil
		}
		r.logger.Error("Failed to find repository by ID: %v", err)
		return nil, err
	}

	return &repo, nil
}

func (r *RepositoryRepositoryMongo) FindByOwnerAndName(ctx context.Context, owner, name string) (*entity.Repository, error) {
	filter := bson.M{
		"ownerLogin": owner,
		"name":       name,
	}

	var repo entity.Repository
	err := r.collection.FindOne(ctx, filter).Decode(&repo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			r.logger.Info("Repository not found: %s/%s", owner, name)
			return nil, nil
		}
		r.logger.Error("Failed to find repository by owner and name: %v", err)
		return nil, err
	}

	return &repo, nil
}

func (r *RepositoryRepositoryMongo) List(ctx context.Context, filter repository.RepositoryFilter) ([]*entity.Repository, error) {
	findFilter := bson.M{}

	if filter.OwnerLogin != "" {
		findFilter["ownerLogin"] = filter.OwnerLogin
	}

	if filter.Language != "" {
		findFilter["language"] = filter.Language
	}

	if filter.MinStars > 0 {
		findFilter["starsCount"] = bson.M{"$gte": filter.MinStars}
	}

	// Настройка пагинации
	findOptions := options.Find()
	if filter.Limit > 0 {
		findOptions.SetLimit(int64(filter.Limit))
	}
	if filter.Offset > 0 {
		findOptions.SetSkip(int64(filter.Offset))
	}

	// Сортировка по количеству звезд (сначала популярные)
	findOptions.SetSort(bson.M{"starsCount": -1})

	cursor, err := r.collection.Find(ctx, findFilter, findOptions)
	if err != nil {
		r.logger.Error("Failed to list repositories: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var repos []*entity.Repository
	if err := cursor.All(ctx, &repos); err != nil {
		r.logger.Error("Failed to decode repositories: %v", err)
		return nil, err
	}

	return repos, nil
}

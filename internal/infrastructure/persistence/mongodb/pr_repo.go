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

type PullRequestRepositoryMongo struct {
	collection *mongo.Collection
	logger     *logger.Logger
}

func NewPullRequestRepository(db *mongo.Database, logger *logger.Logger) repository.PullRequestRepository {
	return &PullRequestRepositoryMongo{
		collection: db.Collection("pull_requests"),
		logger:     logger,
	}
}

func (r *PullRequestRepositoryMongo) Save(ctx context.Context, pr *entity.PullRequest) error {
	filter := bson.M{"id": pr.ID}
	update := bson.M{"$set": bson.M{
		"id":           pr.ID,
		"number":       pr.Number,
		"title":        pr.Title,
		"body":         pr.Body,
		"state":        pr.State,
		"authorLogin":  pr.AuthorLogin,
		"repositoryID": pr.RepositoryID,
		"createdAt":    pr.CreatedAt,
		"updatedAt":    pr.UpdatedAt,
		"mergedAt":     pr.MergedAt,
		"closedAt":     pr.ClosedAt,
	}}

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		r.logger.Error("Failed to save pull request: %v", err)
		return err
	}

	return nil
}

func (r *PullRequestRepositoryMongo) FindByID(ctx context.Context, id int64) (*entity.PullRequest, error) {
	filter := bson.M{"id": id}

	var pr entity.PullRequest
	err := r.collection.FindOne(ctx, filter).Decode(&pr)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			r.logger.Info("Pull request not found: %d", id)
			return nil, nil
		}
		r.logger.Error("Failed to find pull request by ID: %v", err)
		return nil, err
	}

	return &pr, nil
}

func (r *PullRequestRepositoryMongo) FindByNumber(ctx context.Context, repoID int64, number int) (*entity.PullRequest, error) {
	filter := bson.M{
		"repositoryID": repoID,
		"number":       number,
	}

	var pr entity.PullRequest
	err := r.collection.FindOne(ctx, filter).Decode(&pr)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			r.logger.Info("Pull request not found: repo=%d, number=%d", repoID, number)
			return nil, nil
		}
		r.logger.Error("Failed to find pull request by number: %v", err)
		return nil, err
	}

	return &pr, nil
}

func (r *PullRequestRepositoryMongo) List(ctx context.Context, filter repository.PullRequestFilter) ([]*entity.PullRequest, error) {
	findFilter := bson.M{}

	if filter.RepositoryID != 0 {
		findFilter["repositoryID"] = filter.RepositoryID
	}

	if filter.State != "" {
		findFilter["state"] = filter.State
	}

	// Настройка пагинации
	findOptions := options.Find()
	if filter.Limit > 0 {
		findOptions.SetLimit(int64(filter.Limit))
	}
	if filter.Offset > 0 {
		findOptions.SetSkip(int64(filter.Offset))
	}

	// Сортировка по времени создания (сначала новые)
	findOptions.SetSort(bson.M{"createdAt": -1})

	cursor, err := r.collection.Find(ctx, findFilter, findOptions)
	if err != nil {
		r.logger.Error("Failed to list pull requests: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var prs []*entity.PullRequest
	if err := cursor.All(ctx, &prs); err != nil {
		r.logger.Error("Failed to decode pull requests: %v", err)
		return nil, err
	}

	return prs, nil
}

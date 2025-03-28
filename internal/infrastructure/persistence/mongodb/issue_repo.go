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

type IssueRepositoryMongo struct {
	collection *mongo.Collection
	logger     *logger.Logger
}

func NewIssueRepository(db *mongo.Database, logger *logger.Logger) repository.IssueRepository {
	return &IssueRepositoryMongo{
		collection: db.Collection("issues"),
		logger:     logger,
	}
}

func (r *IssueRepositoryMongo) Save(ctx context.Context, issue *entity.Issue) error {
	filter := bson.M{"id": issue.ID}
	update := bson.M{"$set": bson.M{
		"id":           issue.ID,
		"number":       issue.Number,
		"title":        issue.Title,
		"body":         issue.Body,
		"state":        issue.State,
		"authorLogin":  issue.AuthorLogin,
		"repositoryID": issue.RepositoryID,
		"createdAt":    issue.CreatedAt,
		"updatedAt":    issue.UpdatedAt,
		"closedAt":     issue.ClosedAt,
	}}

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		r.logger.Error("Failed to save issue: %v", err)
		return err
	}

	return nil
}

func (r *IssueRepositoryMongo) FindByID(ctx context.Context, id int64) (*entity.Issue, error) {
	filter := bson.M{"id": id}

	var issue entity.Issue
	err := r.collection.FindOne(ctx, filter).Decode(&issue)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			r.logger.Info("Issue not found: %d", id)
			return nil, nil
		}
		r.logger.Error("Failed to find issue by ID: %v", err)
		return nil, err
	}

	return &issue, nil
}

func (r *IssueRepositoryMongo) List(ctx context.Context, filter repository.IssueFilter) ([]*entity.Issue, error) {
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
		r.logger.Error("Failed to list issues: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var issues []*entity.Issue
	if err := cursor.All(ctx, &issues); err != nil {
		r.logger.Error("Failed to decode issues: %v", err)
		return nil, err
	}

	return issues, nil
}

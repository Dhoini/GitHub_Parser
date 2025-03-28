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

type UserRepositoryMongo struct {
	collection *mongo.Collection
	logger     *logger.Logger
}

func NewUserRepository(db *mongo.Database, logger *logger.Logger) repository.UserRepository {
	return &UserRepositoryMongo{
		collection: db.Collection("users"),
		logger:     logger,
	}
}

func (r *UserRepositoryMongo) Save(ctx context.Context, user *entity.User) error {
	filter := bson.M{"id": user.ID}
	update := bson.M{"$set": bson.M{
		"id":        user.ID,
		"login":     user.Login,
		"name":      user.Name,
		"email":     user.Email,
		"avatarURL": user.AvatarURL,
		"bio":       user.Bio,
		"company":   user.Company,
		"location":  user.Location,
		"createdAt": user.CreatedAt,
		"updatedAt": user.UpdatedAt,
	}}

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		r.logger.Error("Failed to save user: %v", err)
		return err
	}

	return nil
}

func (r *UserRepositoryMongo) FindByID(ctx context.Context, id int64) (*entity.User, error) {
	filter := bson.M{"id": id}

	var user entity.User
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			r.logger.Info("User not found: %d", id)
			return nil, nil
		}
		r.logger.Error("Failed to find user by ID: %v", err)
		return nil, err
	}

	return &user, nil
}

func (r *UserRepositoryMongo) FindByLogin(ctx context.Context, login string) (*entity.User, error) {
	filter := bson.M{"login": login}

	var user entity.User
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			r.logger.Info("User not found: %s", login)
			return nil, nil
		}
		r.logger.Error("Failed to find user by login: %v", err)
		return nil, err
	}

	return &user, nil
}

func (r *UserRepositoryMongo) List(ctx context.Context, filter repository.UserFilter) ([]*entity.User, error) {
	findFilter := bson.M{}

	if filter.Login != "" {
		// Используем регулярное выражение для частичного совпадения логина
		findFilter["login"] = bson.M{"$regex": filter.Login, "$options": "i"}
	}

	// Настройка пагинации
	findOptions := options.Find()
	if filter.Limit > 0 {
		findOptions.SetLimit(int64(filter.Limit))
	}
	if filter.Offset > 0 {
		findOptions.SetSkip(int64(filter.Offset))
	}

	// Сортировка по логину
	findOptions.SetSort(bson.M{"login": 1})

	cursor, err := r.collection.Find(ctx, findFilter, findOptions)
	if err != nil {
		r.logger.Error("Failed to list users: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*entity.User
	if err := cursor.All(ctx, &users); err != nil {
		r.logger.Error("Failed to decode users: %v", err)
		return nil, err
	}

	return users, nil
}

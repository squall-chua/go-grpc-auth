package repository

import (
	"context"
	"errors"
	"time"

	"github.com/squall-chua/gmqb"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, namespace, email string) (*domain.User, error)
	GetByUsername(ctx context.Context, namespace, username string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, namespace string, offset, limit int) ([]*domain.User, int64, error)
}

type mongoUserRepository struct {
	collection *gmqb.Collection[domain.User]
	f          func(fieldPath string) string
}

func NewUserRepository(db *mongo.Database) UserRepository {
	return &mongoUserRepository{
		collection: gmqb.Wrap[domain.User](db.Collection("users")),
		f:          gmqb.Field[domain.User],
	}
}

func (r *mongoUserRepository) Create(ctx context.Context, user *domain.User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, user)
	return err
}

func (r *mongoUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	user, err := r.collection.FindOne(ctx, gmqb.Eq(r.f("ID"), id))
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *mongoUserRepository) GetByEmail(ctx context.Context, namespace, email string) (*domain.User, error) {
	user, err := r.collection.FindOne(ctx, gmqb.And(
		gmqb.Eq(r.f("Namespace"), namespace),
		gmqb.Eq(r.f("Email"), email),
	))
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *mongoUserRepository) GetByUsername(ctx context.Context, namespace, username string) (*domain.User, error) {
	user, err := r.collection.FindOne(ctx, gmqb.And(
		gmqb.Eq(r.f("Namespace"), namespace),
		gmqb.Eq(r.f("Username"), username),
	))
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *mongoUserRepository) Update(ctx context.Context, user *domain.User) error {
	user.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx, gmqb.Eq(r.f("ID"), user.ID), user)
	return err
}

func (r *mongoUserRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, gmqb.Eq(r.f("ID"), id))
	return err
}

func (r *mongoUserRepository) List(ctx context.Context, namespace string, offset, limit int) ([]*domain.User, int64, error) {
	filter := gmqb.Eq(r.f("Namespace"), namespace)
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	users, err := r.collection.Find(ctx, filter,
		gmqb.WithSkip(int64(offset)),
		gmqb.WithLimit(int64(limit)),
	)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*domain.User, len(users))
	for i := range users {
		result[i] = &users[i]
	}

	return result, count, nil
}

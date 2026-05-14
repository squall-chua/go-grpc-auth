package repository

import (
	"context"
	"errors"
	"time"

	"github.com/squall-chua/gmqb"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var ErrTokenNotFound = errors.New("token not found")

type TokenRepository interface {
	Create(ctx context.Context, token *domain.Token) error
	GetByHash(ctx context.Context, hash string) (*domain.Token, error)
	DeleteByHash(ctx context.Context, hash string) error
	DeleteByUserID(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
}

type mongoTokenRepository struct {
	collection *gmqb.Collection[domain.Token]
	f          func(fieldPath string) string
}

func NewTokenRepository(db *mongo.Database) TokenRepository {
	return &mongoTokenRepository{
		collection: gmqb.Wrap[domain.Token](db.Collection("tokens")),
		f:          gmqb.Field[domain.Token],
	}
}

func (r *mongoTokenRepository) Create(ctx context.Context, token *domain.Token) error {
	token.CreatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, token)
	return err
}

func (r *mongoTokenRepository) GetByHash(ctx context.Context, hash string) (*domain.Token, error) {
	token, err := r.collection.FindOne(ctx, gmqb.Eq(r.f("TokenHash"), hash))
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrTokenNotFound
		}
		return nil, err
	}
	return token, nil
}

func (r *mongoTokenRepository) DeleteByHash(ctx context.Context, hash string) error {
	_, err := r.collection.DeleteOne(ctx, gmqb.Eq(r.f("TokenHash"), hash))
	return err
}

func (r *mongoTokenRepository) DeleteByUserID(ctx context.Context, userID string) error {
	_, err := r.collection.DeleteMany(ctx, gmqb.Eq(r.f("UserID"), userID))
	return err
}

func (r *mongoTokenRepository) DeleteExpired(ctx context.Context) error {
	_, err := r.collection.DeleteMany(ctx, gmqb.Lt(r.f("ExpiresAt"), time.Now()))
	return err
}

package repository

import (
	"context"
	"time"

	"github.com/squall-chua/gmqb"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type AuthCodeRepository interface {
	Save(ctx context.Context, code *domain.AuthCode) error
	Get(ctx context.Context, code string) (*domain.AuthCode, error)
	Delete(ctx context.Context, code string) error
}

type authCodeRepository struct {
	collection *gmqb.Collection[domain.AuthCode]
	f          func(fieldPath string) string
}

func NewAuthCodeRepository(db *mongo.Database) AuthCodeRepository {
	return &authCodeRepository{
		collection: gmqb.Wrap[domain.AuthCode](db.Collection("auth_codes")),
		f:          gmqb.Field[domain.AuthCode],
	}
}

func (r *authCodeRepository) Save(ctx context.Context, code *domain.AuthCode) error {
	_, err := r.collection.InsertOne(ctx, code)
	return err
}

func (r *authCodeRepository) Get(ctx context.Context, code string) (*domain.AuthCode, error) {
	ac, err := r.collection.FindOne(ctx, gmqb.Eq(r.f("Code"), code))
	if err != nil {
		return nil, err
	}
	if ac.ExpiresAt.Before(time.Now()) {
		return nil, mongo.ErrNoDocuments
	}
	return ac, nil
}

func (r *authCodeRepository) Delete(ctx context.Context, code string) error {
	_, err := r.collection.DeleteOne(ctx, gmqb.Eq(r.f("Code"), code))
	return err
}

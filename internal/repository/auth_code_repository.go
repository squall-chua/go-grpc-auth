package repository

import (
	"context"
	"time"

	"github.com/squall-chua/gmqb"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
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
	if code.ID == bson.NilObjectID {
		code.ID = bson.NewObjectID()
	}
	_, err := r.collection.InsertOne(ctx, code)
	return err
}

func (r *authCodeRepository) Get(ctx context.Context, code string) (*domain.AuthCode, error) {
	filter := gmqb.And(
		gmqb.Eq(r.f("Code"), code),
		gmqb.Gt(r.f("ExpiresAt"), time.Now().UTC()),
	)
	return r.collection.FindOne(ctx, filter)
}

func (r *authCodeRepository) Delete(ctx context.Context, code string) error {
	_, err := r.collection.DeleteOne(ctx, gmqb.Eq(r.f("Code"), code))
	return err
}

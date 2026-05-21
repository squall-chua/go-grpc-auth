package repository

import (
	"context"

	"github.com/squall-chua/gmqb"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MFARepository interface {
	UpsertSecret(ctx context.Context, secret *domain.MFASecret) error
	GetSecret(ctx context.Context, userID string, method domain.MFAMethod) (*domain.MFASecret, error)
	DeleteSecret(ctx context.Context, userID string, method domain.MFAMethod) error
	ListEnrolledMethods(ctx context.Context, userID string) ([]domain.MFAMethod, error)

	CreateToken(ctx context.Context, token *domain.MFAToken) error
	GetToken(ctx context.Context, tokenStr string) (*domain.MFAToken, error)
	DeleteToken(ctx context.Context, tokenStr string) error
}

type mongoMFARepository struct {
	secrets *gmqb.Collection[domain.MFASecret]
	tokens  *gmqb.Collection[domain.MFAToken]
	fs      func(fieldPath string) string
	ft      func(fieldPath string) string
}

func NewMFARepository(db *mongo.Database) MFARepository {
	return &mongoMFARepository{
		secrets: gmqb.Wrap[domain.MFASecret](db.Collection("mfa_secrets")),
		tokens:  gmqb.Wrap[domain.MFAToken](db.Collection("mfa_tokens")),
		fs:      gmqb.Field[domain.MFASecret],
		ft:      gmqb.Field[domain.MFAToken],
	}
}

func (r *mongoMFARepository) UpsertSecret(ctx context.Context, secret *domain.MFASecret) error {
	_, err := r.secrets.ReplaceOne(ctx,
		gmqb.And(
			gmqb.Eq(r.fs("UserID"), secret.UserID),
			gmqb.Eq(r.fs("Method"), secret.Method),
		),
		secret,
		gmqb.WithUpsertReplace(true),
	)
	return err
}

func (r *mongoMFARepository) GetSecret(ctx context.Context, userID string, method domain.MFAMethod) (*domain.MFASecret, error) {
	return r.secrets.FindOne(ctx, gmqb.And(
		gmqb.Eq(r.fs("UserID"), userID),
		gmqb.Eq(r.fs("Method"), method),
	))
}

func (r *mongoMFARepository) DeleteSecret(ctx context.Context, userID string, method domain.MFAMethod) error {
	_, err := r.secrets.DeleteOne(ctx, gmqb.And(
		gmqb.Eq(r.fs("UserID"), userID),
		gmqb.Eq(r.fs("Method"), method),
	))
	return err
}

func (r *mongoMFARepository) ListEnrolledMethods(ctx context.Context, userID string) ([]domain.MFAMethod, error) {
	secrets, err := r.secrets.Find(ctx, gmqb.And(
		gmqb.Eq(r.fs("UserID"), userID),
		gmqb.Eq(r.fs("Confirmed"), true),
	))
	if err != nil {
		return nil, err
	}
	methods := make([]domain.MFAMethod, 0, len(secrets))
	for _, s := range secrets {
		methods = append(methods, s.Method)
	}
	return methods, nil
}

func (r *mongoMFARepository) CreateToken(ctx context.Context, token *domain.MFAToken) error {
	if token.ID == bson.NilObjectID {
		token.ID = bson.NewObjectID()
	}
	_, err := r.tokens.InsertOne(ctx, token)
	return err
}

func (r *mongoMFARepository) GetToken(ctx context.Context, tokenStr string) (*domain.MFAToken, error) {
	return r.tokens.FindOne(ctx, gmqb.Eq(r.ft("Token"), tokenStr))
}

func (r *mongoMFARepository) DeleteToken(ctx context.Context, tokenStr string) error {
	_, err := r.tokens.DeleteOne(ctx, gmqb.Eq(r.ft("Token"), tokenStr))
	return err
}

package repository

import (
	"context"

	"github.com/squall-chua/gmqb"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ConsentRepository interface {
	Save(ctx context.Context, consent *domain.Consent) error
	Get(ctx context.Context, userID, clientID string) (*domain.Consent, error)
	Delete(ctx context.Context, userID, clientID string) error
}

type consentRepository struct {
	collection *gmqb.Collection[domain.Consent]
	f          func(fieldPath string) string
}

func NewConsentRepository(db *mongo.Database) ConsentRepository {
	return &consentRepository{
		collection: gmqb.Wrap[domain.Consent](db.Collection("consents")),
		f:          gmqb.Field[domain.Consent],
	}
}

func (r *consentRepository) Save(ctx context.Context, consent *domain.Consent) error {
	if consent.ID == bson.NilObjectID {
		consent.ID = bson.NewObjectID()
	}
	_, err := r.collection.ReplaceOne(ctx, gmqb.And(
		gmqb.Eq(r.f("UserID"), consent.UserID),
		gmqb.Eq(r.f("ClientID"), consent.ClientID),
	), consent, gmqb.WithUpsertReplace(true))
	return err
}

func (r *consentRepository) Get(ctx context.Context, userID, clientID string) (*domain.Consent, error) {
	return r.collection.FindOne(ctx, gmqb.And(
		gmqb.Eq(r.f("UserID"), userID),
		gmqb.Eq(r.f("ClientID"), clientID),
	))
}

func (r *consentRepository) Delete(ctx context.Context, userID, clientID string) error {
	_, err := r.collection.DeleteOne(ctx, gmqb.And(
		gmqb.Eq(r.f("UserID"), userID),
		gmqb.Eq(r.f("ClientID"), clientID),
	))
	return err
}

package repository

import (
	"context"
	"time"

	"github.com/squall-chua/gmqb"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	Get(ctx context.Context, id string) (*domain.Session, error)
	Delete(ctx context.Context, id string) error
	DeleteByUserID(ctx context.Context, userID string) error
}

type sessionRepository struct {
	collection *gmqb.Collection[domain.Session]
	f          func(fieldPath string) string
}

func NewSessionRepository(db *mongo.Database) SessionRepository {
	return &sessionRepository{
		collection: gmqb.Wrap[domain.Session](db.Collection("sessions")),
		f:          gmqb.Field[domain.Session],
	}
}

func (r *sessionRepository) Create(ctx context.Context, session *domain.Session) error {
	if session.ID == bson.NilObjectID {
		session.ID = bson.NewObjectID()
	}
	_, err := r.collection.InsertOne(ctx, session)
	return err
}

func (r *sessionRepository) Get(ctx context.Context, id string) (*domain.Session, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	filter := gmqb.And(
		gmqb.Eq(r.f("ID"), oid),
		gmqb.Gt(r.f("ExpiresAt"), time.Now().UTC()),
	)
	return r.collection.FindOne(ctx, filter)
}

func (r *sessionRepository) Delete(ctx context.Context, id string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection.DeleteOne(ctx, gmqb.Eq(r.f("ID"), oid))
	return err
}

func (r *sessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	_, err := r.collection.DeleteMany(ctx, gmqb.Eq(r.f("UserID"), userID))
	return err
}

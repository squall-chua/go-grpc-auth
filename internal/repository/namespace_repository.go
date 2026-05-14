package repository

import (
	"context"
	"errors"

	"github.com/squall-chua/gmqb"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var ErrNamespaceNotFound = errors.New("namespace not found")

type NamespaceRepository interface {
	Create(ctx context.Context, ns *domain.Namespace) error
	GetByID(ctx context.Context, id string) (*domain.Namespace, error)
	GetByName(ctx context.Context, name string) (*domain.Namespace, error)
	List(ctx context.Context, offset, limit int) ([]*domain.Namespace, int64, error)
	Update(ctx context.Context, ns *domain.Namespace) error
	Delete(ctx context.Context, id string) error
}

type mongoNamespaceRepository struct {
	collection *gmqb.Collection[domain.Namespace]
	f          func(fieldPath string) string
}

func NewNamespaceRepository(db *mongo.Database) NamespaceRepository {
	return &mongoNamespaceRepository{
		collection: gmqb.Wrap[domain.Namespace](db.Collection("namespaces")),
		f:          gmqb.Field[domain.Namespace],
	}
}

func (r *mongoNamespaceRepository) Create(ctx context.Context, ns *domain.Namespace) error {
	_, err := r.collection.InsertOne(ctx, ns)
	return err
}

func (r *mongoNamespaceRepository) GetByID(ctx context.Context, id string) (*domain.Namespace, error) {
	ns, err := r.collection.FindOne(ctx, gmqb.Eq(r.f("ID"), id))
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNamespaceNotFound
		}
		return nil, err
	}
	return ns, nil
}

func (r *mongoNamespaceRepository) GetByName(ctx context.Context, name string) (*domain.Namespace, error) {
	ns, err := r.collection.FindOne(ctx, gmqb.Eq(r.f("Name"), name))
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNamespaceNotFound
		}
		return nil, err
	}
	return ns, nil
}

func (r *mongoNamespaceRepository) List(ctx context.Context, offset, limit int) ([]*domain.Namespace, int64, error) {
	count, err := r.collection.CountDocuments(ctx, gmqb.NewFilter())
	if err != nil {
		return nil, 0, err
	}

	namespaces, err := r.collection.Find(ctx, gmqb.NewFilter(),
		gmqb.WithSkip(int64(offset)),
		gmqb.WithLimit(int64(limit)),
	)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*domain.Namespace, len(namespaces))
	for i := range namespaces {
		result[i] = &namespaces[i]
	}
	return result, count, nil
}

func (r *mongoNamespaceRepository) Update(ctx context.Context, ns *domain.Namespace) error {
	_, err := r.collection.ReplaceOne(ctx, gmqb.Eq(r.f("ID"), ns.ID), ns)
	return err
}

func (r *mongoNamespaceRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, gmqb.Eq(r.f("ID"), id))
	return err
}

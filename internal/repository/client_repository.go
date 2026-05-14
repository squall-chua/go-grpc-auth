package repository

import (
	"context"

	"github.com/squall-chua/gmqb"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ClientRepository interface {
	Create(ctx context.Context, client *domain.OIDCClient) error
	GetByID(ctx context.Context, clientID string) (*domain.OIDCClient, error)
	List(ctx context.Context, namespace string) ([]*domain.OIDCClient, error)
	Update(ctx context.Context, client *domain.OIDCClient) error
	Delete(ctx context.Context, clientID string) error
}

type mongoClientRepository struct {
	collection *gmqb.Collection[domain.OIDCClient]
	f          func(fieldPath string) string
}

func NewClientRepository(db *mongo.Database) ClientRepository {
	return &mongoClientRepository{
		collection: gmqb.Wrap[domain.OIDCClient](db.Collection("oidc_clients")),
		f:          gmqb.Field[domain.OIDCClient],
	}
}

func (r *mongoClientRepository) Create(ctx context.Context, client *domain.OIDCClient) error {
	_, err := r.collection.InsertOne(ctx, client)
	return err
}

func (r *mongoClientRepository) GetByID(ctx context.Context, clientID string) (*domain.OIDCClient, error) {
	return r.collection.FindOne(ctx, gmqb.Eq(r.f("ClientID"), clientID))
}

func (r *mongoClientRepository) List(ctx context.Context, namespace string) ([]*domain.OIDCClient, error) {
	filter := gmqb.NewFilter()
	if namespace != "" {
		filter.Eq(r.f("Namespace"), namespace)
	}
	clients, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.OIDCClient, len(clients))
	for i := range clients {
		result[i] = &clients[i]
	}
	return result, nil
}

func (r *mongoClientRepository) Update(ctx context.Context, client *domain.OIDCClient) error {
	_, err := r.collection.ReplaceOne(ctx, gmqb.Eq(r.f("ClientID"), client.ClientID), client)
	return err
}

func (r *mongoClientRepository) Delete(ctx context.Context, clientID string) error {
	_, err := r.collection.DeleteOne(ctx, gmqb.Eq(r.f("ClientID"), clientID))
	return err
}

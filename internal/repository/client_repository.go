package repository

import (
	"context"
	"time"

	"github.com/squall-chua/gmqb"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ClientUpdateFields struct {
	Name          string
	RedirectURIs  []string
	AllowedScopes []string
	SkipConsent   bool
}

type ClientRepository interface {
	Create(ctx context.Context, client *domain.OIDCClient) error
	GetByID(ctx context.Context, clientID string) (*domain.OIDCClient, error)
	List(ctx context.Context, namespace, query string, offset, limit int) ([]*domain.OIDCClient, int64, error)
	Update(ctx context.Context, clientID string, fields ClientUpdateFields) (*domain.OIDCClient, error)
	UpdateSecret(ctx context.Context, clientID string, hashedSecret string) error
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
	if client.ID == bson.NilObjectID {
		client.ID = bson.NewObjectID()
	}
	_, err := r.collection.InsertOne(ctx, client)
	return err
}

func (r *mongoClientRepository) GetByID(ctx context.Context, clientID string) (*domain.OIDCClient, error) {
	return r.collection.FindOne(ctx, gmqb.Eq(r.f("ClientID"), clientID))
}

func (r *mongoClientRepository) List(ctx context.Context, namespace, query string, offset, limit int) ([]*domain.OIDCClient, int64, error) {
	conditions := []gmqb.Filter{}
	if namespace != "" {
		conditions = append(conditions, gmqb.Eq(r.f("Namespace"), namespace))
	}
	if query != "" {
		conditions = append(conditions, gmqb.Or(
			gmqb.Regex(r.f("ClientID"), query, "i"),
			gmqb.Regex(r.f("Name"), query, "i"),
		))
	}

	var filter gmqb.Filter
	if len(conditions) > 0 {
		filter = gmqb.And(conditions...)
	} else {
		filter = gmqb.NewFilter()
	}

	pipeline := gmqb.NewPipeline().
		Match(filter).
		Facet(map[string]gmqb.Pipeline{
			"data": gmqb.NewPipeline().
				Skip(int64(offset)).
				Limit(int64(limit)),
			"metadata": gmqb.NewPipeline().
				Count("total"),
		})

	type resultDoc struct {
		Data     []domain.OIDCClient `bson:"data"`
		Metadata []struct {
			Total int64 `bson:"total"`
		} `bson:"metadata"`
	}

	results, err := gmqb.Aggregate[resultDoc](r.collection, ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}

	if len(results) == 0 {
		return nil, 0, nil
	}

	res := results[0]
	total := int64(0)
	if len(res.Metadata) > 0 {
		total = res.Metadata[0].Total
	}

	clients := make([]*domain.OIDCClient, len(res.Data))
	for i := range res.Data {
		clients[i] = &res.Data[i]
	}

	return clients, total, nil
}

func (r *mongoClientRepository) Update(ctx context.Context, clientID string, fields ClientUpdateFields) (*domain.OIDCClient, error) {
	return r.collection.FindOneAndUpdate(ctx,
		gmqb.Eq(r.f("ClientID"), clientID),
		gmqb.NewUpdate().
			Set(r.f("Name"), fields.Name).
			Set(r.f("RedirectURIs"), fields.RedirectURIs).
			Set(r.f("AllowedScopes"), fields.AllowedScopes).
			Set(r.f("SkipConsent"), fields.SkipConsent).
			Set(r.f("UpdatedAt"), time.Now().UTC()),
		gmqb.WithReturnDocument(options.After),
	)
}

func (r *mongoClientRepository) UpdateSecret(ctx context.Context, clientID string, hashedSecret string) error {
	_, err := r.collection.UpdateOne(ctx, gmqb.Eq(r.f("ClientID"), clientID),
		gmqb.NewUpdate().
			Set(r.f("ClientSecret"), hashedSecret).
			Set(r.f("UpdatedAt"), time.Now().UTC()),
	)
	return err
}

func (r *mongoClientRepository) Delete(ctx context.Context, clientID string) error {
	_, err := r.collection.DeleteOne(ctx, gmqb.Eq(r.f("ClientID"), clientID))
	return err
}

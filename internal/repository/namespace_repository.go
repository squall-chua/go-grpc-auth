package repository

import (
	"context"
	"errors"

	"github.com/squall-chua/gmqb"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
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
	if ns.ID == bson.NilObjectID {
		ns.ID = bson.NewObjectID()
	}
	_, err := r.collection.InsertOne(ctx, ns)
	return err
}

func (r *mongoNamespaceRepository) GetByID(ctx context.Context, id string) (*domain.Namespace, error) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	ns, err := r.collection.FindOne(ctx, gmqb.Eq(r.f("ID"), objID))
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
	filter := gmqb.NewFilter()

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
		Data     []domain.Namespace `bson:"data"`
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

	namespaces := make([]*domain.Namespace, len(res.Data))
	for i := range res.Data {
		namespaces[i] = &res.Data[i]
	}

	return namespaces, total, nil
}

func (r *mongoNamespaceRepository) Update(ctx context.Context, ns *domain.Namespace) error {
	_, err := r.collection.UpdateOne(ctx, gmqb.Eq(r.f("ID"), ns.ID),
		gmqb.NewUpdate().
			Set(r.f("Name"), ns.Name).
			Set(r.f("Config"), ns.Config),
	)
	return err
}

func (r *mongoNamespaceRepository) Delete(ctx context.Context, id string) error {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection.DeleteOne(ctx, gmqb.Eq(r.f("ID"), objID))
	return err
}

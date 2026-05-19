package repository

import (
	"context"

	"github.com/squall-chua/gmqb"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type PermissionRepository interface {
	Create(ctx context.Context, perm *domain.Permission) error
	GetByName(ctx context.Context, namespace, name string) (*domain.Permission, error)
	List(ctx context.Context, query string, offset, limit int) ([]*domain.Permission, int64, error)
	Delete(ctx context.Context, id string) error
}

type mongoPermissionRepository struct {
	collection *gmqb.Collection[domain.Permission]
	f          func(fieldPath string) string
}

func NewPermissionRepository(db *mongo.Database) PermissionRepository {
	return &mongoPermissionRepository{
		collection: gmqb.Wrap[domain.Permission](db.Collection("permissions")),
		f:          gmqb.Field[domain.Permission],
	}
}

func (r *mongoPermissionRepository) GetByName(ctx context.Context, namespace, name string) (*domain.Permission, error) {
	return r.collection.FindOne(ctx, gmqb.And(
		gmqb.Eq(r.f("Namespace"), namespace),
		gmqb.Eq(r.f("Name"), name),
	))
}

func (r *mongoPermissionRepository) Create(ctx context.Context, perm *domain.Permission) error {
	if perm.ID == bson.NilObjectID {
		perm.ID = bson.NewObjectID()
	}

	filter := gmqb.And(
		gmqb.Eq(r.f("Namespace"), perm.Namespace),
		gmqb.Eq(r.f("Name"), perm.Name),
	)

	update := gmqb.NewUpdate().
		SetOnInsert(r.f("ID"), perm.ID).
		SetOnInsert(r.f("Name"), perm.Name).
		SetOnInsert(r.f("Namespace"), perm.Namespace).
		SetOnInsert(r.f("Description"), perm.Description).
		SetOnInsert(r.f("CreatedAt"), perm.CreatedAt).
		SetOnInsert(r.f("UpdatedAt"), perm.UpdatedAt)

	result, err := r.collection.UpsertOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.UpsertedCount == 0 {
		existing, err := r.GetByName(ctx, perm.Namespace, perm.Name)
		if err != nil {
			return err
		}
		*perm = *existing
	}

	return nil
}

func (r *mongoPermissionRepository) List(ctx context.Context, query string, offset, limit int) ([]*domain.Permission, int64, error) {
	pipeline := gmqb.NewPipeline()
	if query != "" {
		pipeline = pipeline.Match(gmqb.Or(
			gmqb.Regex(r.f("Name"), query, "i"),
			gmqb.Regex(r.f("Namespace"), query, "i"),
		))
	}

	pipeline = pipeline.
		Facet(map[string]gmqb.Pipeline{
			"data": gmqb.NewPipeline().
				Sort(gmqb.Asc(r.f("Name"))).
				Skip(int64(offset)).
				Limit(int64(limit)),
			"metadata": gmqb.NewPipeline().
				Count("total"),
		})

	type resultDoc struct {
		Data     []domain.Permission `bson:"data"`
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

	perms := make([]*domain.Permission, len(res.Data))
	for i := range res.Data {
		perms[i] = &res.Data[i]
	}

	return perms, total, nil
}

func (r *mongoPermissionRepository) Delete(ctx context.Context, id string) error {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection.DeleteOne(ctx, gmqb.Eq(r.f("ID"), objID))
	return err
}

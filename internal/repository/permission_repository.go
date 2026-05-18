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
	List(ctx context.Context, namespace string, offset, limit int) ([]*domain.Permission, int64, error)
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

func (r *mongoPermissionRepository) Create(ctx context.Context, perm *domain.Permission) error {
	if perm.ID == bson.NilObjectID {
		perm.ID = bson.NewObjectID()
	}
	_, err := r.collection.InsertOne(ctx, perm)
	return err
}

func (r *mongoPermissionRepository) List(ctx context.Context, namespace string, offset, limit int) ([]*domain.Permission, int64, error) {
	filter := gmqb.Eq(r.f("Namespace"), namespace)

	pipeline := gmqb.NewPipeline().
		Match(filter).
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

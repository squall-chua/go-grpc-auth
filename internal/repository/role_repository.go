package repository

import (
	"context"

	"github.com/squall-chua/gmqb"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type RoleRepository interface {
	Create(ctx context.Context, role *domain.Role) error
	GetByID(ctx context.Context, id string) (*domain.Role, error)
	GetByName(ctx context.Context, namespace, name string) (*domain.Role, error)
	List(ctx context.Context, query string, offset, limit int) ([]*domain.Role, int64, error)
	Update(ctx context.Context, role *domain.Role) error
	Delete(ctx context.Context, id string) error
}

type mongoRoleRepository struct {
	collection *gmqb.Collection[domain.Role]
	f          func(fieldPath string) string
}

func NewRoleRepository(db *mongo.Database) RoleRepository {
	return &mongoRoleRepository{
		collection: gmqb.Wrap[domain.Role](db.Collection("roles")),
		f:          gmqb.Field[domain.Role],
	}
}

func (r *mongoRoleRepository) Create(ctx context.Context, role *domain.Role) error {
	if role.ID == bson.NilObjectID {
		role.ID = bson.NewObjectID()
	}

	filter := gmqb.And(
		gmqb.Eq(r.f("Namespace"), role.Namespace),
		gmqb.Eq(r.f("Name"), role.Name),
	)

	update := gmqb.NewUpdate().
		SetOnInsert(r.f("ID"), role.ID).
		SetOnInsert(r.f("Name"), role.Name).
		SetOnInsert(r.f("Namespace"), role.Namespace).
		SetOnInsert(r.f("Permissions"), role.Permissions).
		SetOnInsert(r.f("CreatedAt"), role.CreatedAt).
		SetOnInsert(r.f("UpdatedAt"), role.UpdatedAt)

	result, err := r.collection.UpsertOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.UpsertedCount == 0 {
		existing, err := r.GetByName(ctx, role.Namespace, role.Name)
		if err != nil {
			return err
		}
		*role = *existing
	}

	return nil
}

func (r *mongoRoleRepository) GetByID(ctx context.Context, id string) (*domain.Role, error) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return r.collection.FindOne(ctx, gmqb.Eq(r.f("ID"), objID))
}

func (r *mongoRoleRepository) GetByName(ctx context.Context, namespace, name string) (*domain.Role, error) {
	return r.collection.FindOne(ctx, gmqb.And(
		gmqb.Eq(r.f("Namespace"), namespace),
		gmqb.Eq(r.f("Name"), name),
	))
}

func (r *mongoRoleRepository) List(ctx context.Context, query string, offset, limit int) ([]*domain.Role, int64, error) {
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
		Data     []domain.Role `bson:"data"`
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

	roles := make([]*domain.Role, len(res.Data))
	for i := range res.Data {
		roles[i] = &res.Data[i]
	}

	return roles, total, nil
}

func (r *mongoRoleRepository) Update(ctx context.Context, role *domain.Role) error {
	_, err := r.collection.ReplaceOne(ctx, gmqb.Eq(r.f("ID"), role.ID), role)
	return err
}

func (r *mongoRoleRepository) Delete(ctx context.Context, id string) error {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection.DeleteOne(ctx, gmqb.Eq(r.f("ID"), objID))
	return err
}

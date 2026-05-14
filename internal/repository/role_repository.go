package repository

import (
	"context"

	"github.com/squall-chua/gmqb"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type RoleRepository interface {
	Create(ctx context.Context, role *domain.Role) error
	GetByID(ctx context.Context, id string) (*domain.Role, error)
	GetByName(ctx context.Context, namespace, name string) (*domain.Role, error)
	List(ctx context.Context, namespace string) ([]*domain.Role, error)
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
	_, err := r.collection.InsertOne(ctx, role)
	return err
}

func (r *mongoRoleRepository) GetByID(ctx context.Context, id string) (*domain.Role, error) {
	return r.collection.FindOne(ctx, gmqb.Eq(r.f("ID"), id))
}

func (r *mongoRoleRepository) GetByName(ctx context.Context, namespace, name string) (*domain.Role, error) {
	return r.collection.FindOne(ctx, gmqb.And(
		gmqb.Eq(r.f("Namespace"), namespace),
		gmqb.Eq(r.f("Name"), name),
	))
}

func (r *mongoRoleRepository) List(ctx context.Context, namespace string) ([]*domain.Role, error) {
	roles, err := r.collection.Find(ctx, gmqb.Eq(r.f("Namespace"), namespace))
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Role, len(roles))
	for i := range roles {
		result[i] = &roles[i]
	}
	return result, nil
}

func (r *mongoRoleRepository) Update(ctx context.Context, role *domain.Role) error {
	_, err := r.collection.ReplaceOne(ctx, gmqb.Eq(r.f("ID"), role.ID), role)
	return err
}

func (r *mongoRoleRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, gmqb.Eq(r.f("ID"), id))
	return err
}

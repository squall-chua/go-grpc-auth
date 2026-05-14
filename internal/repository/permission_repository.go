package repository

import (
	"context"

	"github.com/squall-chua/gmqb"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type PermissionRepository interface {
	Create(ctx context.Context, perm *domain.Permission) error
	List(ctx context.Context, namespace string) ([]*domain.Permission, error)
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
	_, err := r.collection.InsertOne(ctx, perm)
	return err
}

func (r *mongoPermissionRepository) List(ctx context.Context, namespace string) ([]*domain.Permission, error) {
	perms, err := r.collection.Find(ctx, gmqb.Eq(r.f("Namespace"), namespace))
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Permission, len(perms))
	for i := range perms {
		result[i] = &perms[i]
	}
	return result, nil
}

func (r *mongoPermissionRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, gmqb.Eq(r.f("ID"), id))
	return err
}

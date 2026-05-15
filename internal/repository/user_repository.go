package repository

import (
	"context"
	"errors"
	"time"

	"github.com/squall-chua/gmqb"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, namespace, email string) (*domain.User, error)
	GetByUsername(ctx context.Context, namespace, username string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, namespace string, offset, limit int) ([]*domain.User, int64, error)
}

type mongoUserRepository struct {
	collection *gmqb.Collection[domain.User]
	f          func(fieldPath string) string
}

func NewUserRepository(db *mongo.Database) UserRepository {
	return &mongoUserRepository{
		collection: gmqb.Wrap[domain.User](db.Collection("users")),
		f:          gmqb.Field[domain.User],
	}
}

func (r *mongoUserRepository) Create(ctx context.Context, user *domain.User) error {
	if user.ID == bson.NilObjectID {
		user.ID = bson.NewObjectID()
	}
	user.CreatedAt = time.Now().UTC()
	user.UpdatedAt = time.Now().UTC()
	_, err := r.collection.InsertOne(ctx, user)
	return err
}

func (r *mongoUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	user, err := r.collection.FindOne(ctx, gmqb.Eq(r.f("ID"), objID))
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *mongoUserRepository) GetByEmail(ctx context.Context, namespace, email string) (*domain.User, error) {
	user, err := r.collection.FindOne(ctx, gmqb.And(
		gmqb.Eq(r.f("Namespace"), namespace),
		gmqb.Eq(r.f("Email"), email),
	))
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *mongoUserRepository) GetByUsername(ctx context.Context, namespace, username string) (*domain.User, error) {
	user, err := r.collection.FindOne(ctx, gmqb.And(
		gmqb.Eq(r.f("Namespace"), namespace),
		gmqb.Eq(r.f("Username"), username),
	))
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *mongoUserRepository) Update(ctx context.Context, user *domain.User) error {
	user.UpdatedAt = time.Now().UTC()
	_, err := r.collection.UpdateOne(ctx, gmqb.Eq(r.f("ID"), user.ID),
		gmqb.NewUpdate().
			Set(r.f("Email"), user.Email).
			Set(r.f("Username"), user.Username).
			Set(r.f("PasswordHash"), user.PasswordHash).
			Set(r.f("Status"), user.Status).
			Set(r.f("Roles"), user.Roles).
			Set(r.f("Permissions"), user.Permissions).
			Set(r.f("PasswordHistory"), user.PasswordHistory).
			Set(r.f("SocialIdentities"), user.SocialIdentities).
			Set(r.f("UpdatedAt"), user.UpdatedAt),
	)
	return err
}

func (r *mongoUserRepository) Delete(ctx context.Context, id string) error {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection.DeleteOne(ctx, gmqb.Eq(r.f("ID"), objID))
	return err
}

func (r *mongoUserRepository) List(ctx context.Context, namespace string, offset, limit int) ([]*domain.User, int64, error) {
	filter := gmqb.Eq(r.f("Namespace"), namespace)

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
		Data     []domain.User `bson:"data"`
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

	users := make([]*domain.User, len(res.Data))
	for i := range res.Data {
		users[i] = &res.Data[i]
	}

	return users, total, nil
}

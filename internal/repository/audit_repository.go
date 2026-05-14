package repository

import (
	"context"

	"github.com/squall-chua/gmqb"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type AuditRepository interface {
	Create(ctx context.Context, log *domain.AuditLog) error
	List(ctx context.Context, filter gmqb.Filter, offset, limit int) ([]*domain.AuditLog, int64, error)
}

type mongoAuditRepository struct {
	collection *gmqb.Collection[domain.AuditLog]
}

func NewAuditRepository(db *mongo.Database) AuditRepository {
	return &mongoAuditRepository{
		collection: gmqb.Wrap[domain.AuditLog](db.Collection("audit_logs")),
	}
}

func (r *mongoAuditRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	_, err := r.collection.InsertOne(ctx, log)
	return err
}

func (r *mongoAuditRepository) List(ctx context.Context, filter gmqb.Filter, offset, limit int) ([]*domain.AuditLog, int64, error) {
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	logs, err := r.collection.Find(ctx, filter,
		gmqb.WithSort(gmqb.Desc("timestamp")),
		gmqb.WithSkip(int64(offset)),
		gmqb.WithLimit(int64(limit)),
	)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*domain.AuditLog, len(logs))
	for i := range logs {
		result[i] = &logs[i]
	}

	return result, count, nil
}

package repository

import (
	"context"

	"github.com/squall-chua/gmqb"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
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
	if log.ID == bson.NilObjectID {
		log.ID = bson.NewObjectID()
	}
	_, err := r.collection.InsertOne(ctx, log)
	return err
}

func (r *mongoAuditRepository) List(ctx context.Context, filter gmqb.Filter, offset, limit int) ([]*domain.AuditLog, int64, error) {
	f := gmqb.Field[domain.AuditLog]

	pipeline := gmqb.NewPipeline().
		Match(filter).
		Facet(map[string]gmqb.Pipeline{
			"data": gmqb.NewPipeline().
				Sort(gmqb.Desc(f("ID"))).
				Skip(int64(offset)).
				Limit(int64(limit)),
			"metadata": gmqb.NewPipeline().
				Count("total"),
		})

	type resultDoc struct {
		Data     []domain.AuditLog `bson:"data"`
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

	auditLogs := make([]*domain.AuditLog, len(res.Data))
	for i := range res.Data {
		auditLogs[i] = &res.Data[i]
	}

	return auditLogs, total, nil
}

package audit

import (
	"context"
	"time"

	"github.com/squall-chua/gmqb"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"github.com/squall-chua/go-grpc-auth/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
)

type AuditListFilter struct {
	Namespace string
	Event     string
	UserID    string
	From      time.Time
	To        time.Time
}

type AuditService interface {
	Log(ctx context.Context, event domain.AuditEvent, userID, namespace, ip, ua string, metadata any)
	List(ctx context.Context, filter AuditListFilter, page, pageSize int) ([]*domain.AuditLog, int64, error)
}

type auditService struct {
	repo repository.AuditRepository
	f    func(fieldPath string) string
}

func NewAuditService(repo repository.AuditRepository) AuditService {
	return &auditService{
		repo: repo,
		f:    gmqb.Field[domain.AuditLog],
	}
}

func (s *auditService) Log(ctx context.Context, event domain.AuditEvent, userID, namespace, ip, ua string, metadata any) {
	log := &domain.AuditLog{
		Event:     event,
		UserID:    userID,
		Namespace: namespace,
		IP:        ip,
		UserAgent: ua,
		Metadata:  metadata,
		Timestamp: time.Now().UTC(),
	}

	// We fire and forget audit logs in a goroutine to not block the main flow
	// In production, consider using a worker pool or message queue
	go func() {
		// Use a new context for the background operation
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.repo.Create(bgCtx, log); err != nil {
			zap.L().Error("Failed to create audit log", zap.Error(err), zap.String("user_id", log.UserID), zap.String("event", string(log.Event)))
		}
	}()
}

func (s *auditService) List(ctx context.Context, listFilter AuditListFilter, page, pageSize int) ([]*domain.AuditLog, int64, error) {
	conditions := []gmqb.Filter{}
	if listFilter.Namespace != "" {
		conditions = append(conditions, gmqb.Eq(s.f("Namespace"), listFilter.Namespace))
	}
	if listFilter.Event != "" {
		conditions = append(conditions, gmqb.Eq(s.f("Event"), listFilter.Event))
	}
	if listFilter.UserID != "" {
		conditions = append(conditions, gmqb.Eq(s.f("UserID"), listFilter.UserID))
	}
	if !listFilter.From.IsZero() {
		conditions = append(conditions, gmqb.Gte(s.f("Timestamp"), listFilter.From))
	}
	if !listFilter.To.IsZero() {
		conditions = append(conditions, gmqb.Lt(s.f("Timestamp"), listFilter.To))
	}

	var filter gmqb.Filter
	if len(conditions) > 0 {
		filter = gmqb.And(conditions...)
	} else {
		filter = gmqb.Raw(bson.D{})
	}

	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filter, offset, pageSize)
}

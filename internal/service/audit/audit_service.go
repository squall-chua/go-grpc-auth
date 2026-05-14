package audit

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/squall-chua/gmqb"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"github.com/squall-chua/go-grpc-auth/internal/repository"
)

type AuditService interface {
	Log(ctx context.Context, event domain.AuditEvent, userID, namespace, ip, ua string, metadata any)
	List(ctx context.Context, namespace string, offset, limit int) ([]*domain.AuditLog, int64, error)
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
		ID:        uuid.New().String(),
		Event:     event,
		UserID:    userID,
		Namespace: namespace,
		IP:        ip,
		UserAgent: ua,
		Metadata:  metadata,
		Timestamp: time.Now(),
	}

	// We fire and forget audit logs in a goroutine to not block the main flow
	// In production, consider using a worker pool or message queue
	go func() {
		// Use a new context for the background operation
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.repo.Create(bgCtx, log)
	}()
}

func (s *auditService) List(ctx context.Context, namespace string, offset, limit int) ([]*domain.AuditLog, int64, error) {
	filter := gmqb.NewFilter()
	if namespace != "" {
		filter.Eq(s.f("Namespace"), namespace)
	}
	return s.repo.List(ctx, filter, offset, limit)
}

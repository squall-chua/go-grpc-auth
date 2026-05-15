package audit

import (
	"context"
	"time"

	"github.com/squall-chua/gmqb"
	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"github.com/squall-chua/go-grpc-auth/internal/repository"
	"go.uber.org/zap"
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

func (s *auditService) List(ctx context.Context, namespace string, offset, limit int) ([]*domain.AuditLog, int64, error) {
	filter := gmqb.NewFilter()
	if namespace != "" {
		filter.Eq(s.f("Namespace"), namespace)
	}
	return s.repo.List(ctx, filter, offset, limit)
}

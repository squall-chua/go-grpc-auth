package admin

import (
	"context"

	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"github.com/squall-chua/go-grpc-auth/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NamespaceService interface {
	CreateNamespace(ctx context.Context, ns *domain.Namespace) (*domain.Namespace, error)
	GetNamespace(ctx context.Context, id string) (*domain.Namespace, error)
	ListNamespaces(ctx context.Context, query string, page, pageSize int) ([]*domain.Namespace, int64, error)
	UpdateNamespace(ctx context.Context, ns *domain.Namespace) error
	DeleteNamespace(ctx context.Context, id string) error
}

type namespaceService struct {
	nsRepo repository.NamespaceRepository
}

func NewNamespaceService(nsRepo repository.NamespaceRepository) NamespaceService {
	return &namespaceService{
		nsRepo: nsRepo,
	}
}

func (s *namespaceService) CreateNamespace(ctx context.Context, ns *domain.Namespace) (*domain.Namespace, error) {
	if err := s.nsRepo.Create(ctx, ns); err != nil {
		return nil, status.Error(codes.Internal, "failed to create namespace")
	}
	return ns, nil
}

func (s *namespaceService) GetNamespace(ctx context.Context, id string) (*domain.Namespace, error) {
	ns, err := s.nsRepo.GetByID(ctx, id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "namespace not found")
	}
	return ns, nil
}

func (s *namespaceService) ListNamespaces(ctx context.Context, query string, page, pageSize int) ([]*domain.Namespace, int64, error) {
	offset := (page - 1) * pageSize
	return s.nsRepo.List(ctx, query, offset, pageSize)
}

func (s *namespaceService) UpdateNamespace(ctx context.Context, ns *domain.Namespace) error {
	return s.nsRepo.Update(ctx, ns)
}

func (s *namespaceService) DeleteNamespace(ctx context.Context, id string) error {
	return s.nsRepo.Delete(ctx, id)
}

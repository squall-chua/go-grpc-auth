package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/squall-chua/go-grpc-auth/internal/domain"
	"github.com/squall-chua/go-grpc-auth/internal/repository"
	"go.uber.org/zap"
)

type WebhookService interface {
	Notify(ctx context.Context, namespace string, event domain.AuditEvent, payload any)
}

type webhookService struct {
	client *http.Client
	nsRepo repository.NamespaceRepository
}

func NewWebhookService(nsRepo repository.NamespaceRepository) WebhookService {
	return &webhookService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		nsRepo: nsRepo,
	}
}

func (s *webhookService) Notify(ctx context.Context, namespace string, event domain.AuditEvent, payload any) {
	go func() {
		ns, err := s.nsRepo.GetByName(context.Background(), namespace)
		if err != nil || ns.Config.WebhookURL == "" {
			return
		}

		body := map[string]any{
			"event":     event,
			"namespace": namespace,
			"payload":   payload,
			"timestamp": time.Now().UTC(),
		}

		data, err := json.Marshal(body)
		if err != nil {
			zap.L().Error("webhook: failed to marshal payload", zap.Error(err))
			return
		}

		req, err := http.NewRequest("POST", ns.Config.WebhookURL, bytes.NewBuffer(data))
		if err != nil {
			zap.L().Error("webhook: failed to create request", zap.Error(err))
			return
		}
		req.Header.Set("Content-Type", "application/json")

		if ns.Config.WebhookSecret != "" {
			sig := computeHMAC(data, ns.Config.WebhookSecret)
			req.Header.Set("X-Webhook-Signature", sig)
		}

		resp, err := s.client.Do(req)
		if err != nil {
			zap.L().Warn("webhook: delivery failed", zap.String("namespace", namespace), zap.Error(err))
			return
		}
		resp.Body.Close()

		if resp.StatusCode >= 400 {
			zap.L().Warn("webhook: non-success response", zap.String("namespace", namespace), zap.Int("status", resp.StatusCode))
		}
	}()
}

func computeHMAC(data []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

package webhook

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/squall-chua/go-grpc-auth/internal/domain"
)

type WebhookService interface {
	Notify(ctx context.Context, namespace string, event domain.AuditEvent, payload any)
}

type webhookService struct {
	client *http.Client
}

func NewWebhookService() WebhookService {
	return &webhookService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *webhookService) Notify(ctx context.Context, namespace string, event domain.AuditEvent, payload any) {
	// In a real system, we'd fetch the webhook URL from the namespace config
	// For this phase, we'll assume a placeholder URL or skip if not configured
	
	// We'll run this in a goroutine to not block
	go func() {
		// Mock: check if namespace has webhook configured
		// URL := ns.Config.WebhookURL
		
		// For demo, we'll just log it
		fmt.Printf("[Webhook] Event %s in namespace %s: %+v\n", event, namespace, payload)
		
		/*
		data, _ := json.Marshal(map[string]any{
			"event":     event,
			"namespace": namespace,
			"payload":   payload,
			"timestamp": time.Now().UTC(),
		})
		
		req, _ := http.NewRequest("POST", URL, bytes.NewBuffer(data))
		req.Header.Set("Content-Type", "application/json")
		// Add signature header for security
		
		resp, err := s.client.Do(req)
		if err == nil {
			resp.Body.Close()
		}
		*/
	}()
}

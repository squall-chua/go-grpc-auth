package notification

import (
	"fmt"

	"go.uber.org/zap"
)

// Registry holds the available providers keyed by name plus defaults.
//
// Register* and SetDefault* must be called during startup, before the
// Registry is shared with goroutines that call Pick*/Email/SMS. After
// that point the Registry is treated as read-only and is safe for
// concurrent reads without further synchronization.
type Registry struct {
	email        map[string]EmailSender
	sms          map[string]SMSSender
	defaultEmail string
	defaultSMS   string
}

func NewRegistry() *Registry {
	return &Registry{
		email: make(map[string]EmailSender),
		sms:   make(map[string]SMSSender),
	}
}

func (r *Registry) RegisterEmail(name string, s EmailSender) {
	r.email[name] = s
}

func (r *Registry) RegisterSMS(name string, s SMSSender) {
	r.sms[name] = s
}

func (r *Registry) SetDefaultEmail(name string) { r.defaultEmail = name }
func (r *Registry) SetDefaultSMS(name string)   { r.defaultSMS = name }

// Email returns a registered email sender by name.
func (r *Registry) Email(name string) (EmailSender, bool) {
	s, ok := r.email[name]
	return s, ok
}

// SMS returns a registered SMS sender by name.
func (r *Registry) SMS(name string) (SMSSender, bool) {
	s, ok := r.sms[name]
	return s, ok
}

// PickEmail resolves a sender for the given preferred name. Empty or
// unregistered names fall back to the default. Returns the sender and
// the resolved provider name.
func (r *Registry) PickEmail(preferred string) (EmailSender, string, error) {
	if preferred != "" {
		if s, ok := r.email[preferred]; ok {
			return s, preferred, nil
		}
		zap.L().Warn("notification: email provider not registered, using default",
			zap.String("preferred", preferred), zap.String("default", r.defaultEmail))
	}
	if r.defaultEmail == "" {
		return nil, "", fmt.Errorf("%w: no default email provider", ErrProviderUnavailable)
	}
	s, ok := r.email[r.defaultEmail]
	if !ok {
		return nil, "", fmt.Errorf("%w: default email provider %q not registered", ErrProviderUnavailable, r.defaultEmail)
	}
	return s, r.defaultEmail, nil
}

// PickSMS resolves an SMS sender (symmetric to PickEmail).
func (r *Registry) PickSMS(preferred string) (SMSSender, string, error) {
	if preferred != "" {
		if s, ok := r.sms[preferred]; ok {
			return s, preferred, nil
		}
		zap.L().Warn("notification: sms provider not registered, using default",
			zap.String("preferred", preferred), zap.String("default", r.defaultSMS))
	}
	if r.defaultSMS == "" {
		return nil, "", fmt.Errorf("%w: no default sms provider", ErrProviderUnavailable)
	}
	s, ok := r.sms[r.defaultSMS]
	if !ok {
		return nil, "", fmt.Errorf("%w: default sms provider %q not registered", ErrProviderUnavailable, r.defaultSMS)
	}
	return s, r.defaultSMS, nil
}

// EmailProviderNames returns the names of all registered email providers.
func (r *Registry) EmailProviderNames() []string {
	names := make([]string, 0, len(r.email))
	for name := range r.email {
		names = append(names, name)
	}
	return names
}

// SMSProviderNames returns the names of all registered SMS providers.
func (r *Registry) SMSProviderNames() []string {
	names := make([]string, 0, len(r.sms))
	for name := range r.sms {
		names = append(names, name)
	}
	return names
}

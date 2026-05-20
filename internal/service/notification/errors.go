package notification

import "errors"

var (
	ErrTemplateNotFound    = errors.New("notification: template not found")
	ErrInvalidRecipient    = errors.New("notification: invalid recipient")
	ErrProviderUnavailable = errors.New("notification: no provider available")
	ErrRateLimited         = errors.New("notification: rate limited by provider")
)

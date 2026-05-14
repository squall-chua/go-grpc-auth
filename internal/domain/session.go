package domain

import "time"

type Session struct {
	ID        string
	UserID    string
	Namespace string
	IP        string
	UserAgent string
	CreatedAt time.Time
	ExpiresAt time.Time
}

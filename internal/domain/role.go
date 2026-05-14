package domain

import "time"

type Role struct {
	ID          string    `bson:"id"`
	Name        string    `bson:"name"`
	Namespace   string    `bson:"namespace"`
	Permissions []string  `bson:"permissions"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
}

type Permission struct {
	ID          string    `bson:"id"`
	Name        string    `bson:"name"` // e.g. "users:read"
	Namespace   string    `bson:"namespace"`
	Description string    `bson:"description"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
}

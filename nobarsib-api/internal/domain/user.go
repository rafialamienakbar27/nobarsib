package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Peran pengguna (§4.1). Hanya dua: pengunjung tidak punya akun sama sekali.
const (
	RoleAdmin      = "admin"
	RoleVenueOwner = "venue_owner"
)

type User struct {
	ID           uuid.UUID
	Email        string
	Phone        string
	PasswordHash string
	FullName     string
	Role         string
	IsActive     bool
	LastLoginAt  *time.Time
	CreatedAt    time.Time
}

func (u User) IsAdmin() bool { return u.Role == RoleAdmin }

type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	TouchLogin(ctx context.Context, id uuid.UUID) error
}

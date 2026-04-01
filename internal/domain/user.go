package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Role      UserRole  `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

func NewUser(email string, role UserRole) (*User, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}

	if role != RoleAdmin && role != RoleUser {
		return nil, errors.New("invalid role")
	}

	return &User{
		ID:        uuid.New(),
		Email:     email,
		Role:      role,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) IsUser() bool {
	return u.Role == RoleUser
}

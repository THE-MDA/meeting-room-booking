package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Room struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Capacity    int       `json:"capacity"`
	CreatedAt   time.Time `json:"created_at"`
}

func NewRoom(name string, description string, capacity int) (*Room, error) {
	if name == "" {
		return nil, errors.New("room name is required")
	}

	if capacity <= 0 {
		return nil, errors.New("capacity must be positive")
	}

	return &Room{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Capacity:    capacity,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

func (r *Room) Validate() error {
	if r.Name == "" {
		return errors.New("room name is required")
	}

	if r.Capacity <= 0 {
		return errors.New("capacity must be positive")
	}
	return nil
}

package domain

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByRole(ctx context.Context, role UserRole) ([]*User, error)
}

type RoomRepository interface {
	Create(ctx context.Context, room *Room) error
	GetByID(ctx context.Context, id uuid.UUID) (*Room, error)
	GetAll(ctx context.Context) ([]*Room, error)
}

type ScheduleRepository interface {
	Create(ctx context.Context, schedule *Schedule) error
	GetByRoomID(ctx context.Context, roomID uuid.UUID) ([]*Schedule, error)
	Exists(ctx context.Context, roomID uuid.UUID, dayOfWeek DayOfWeek) (bool, error)
}

type BookingRepository interface {
	Create(ctx context.Context, booking *Booking) error
	GetByID(ctx context.Context, id uuid.UUID) (*Booking, error)
	GetBySlotID(ctx context.Context, slotID string) (*Booking, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, onlyFuture bool) ([]*Booking, error)
	GetAll(ctx context.Context, limit, offset int) ([]*Booking, error)
	Update(ctx context.Context, booking *Booking) error
	Cancel(ctx context.Context, id uuid.UUID) error
	IsSlotBooked(ctx context.Context, slotID string) (bool, error)
	GetBookedSlots(ctx context.Context, slotIDs []string) (map[string]bool, error)
}

package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type BookingStatus string

const (
	BookingStatusActive    BookingStatus = "active"
	BookingStatusCancelled BookingStatus = "cancelled"
)

type Booking struct {
	ID          uuid.UUID     `json:"id"`
	SlotID      string        `json:"slot_id"`
	RoomID      uuid.UUID     `json:"room_id"`
	UserID      uuid.UUID     `json:"user_id"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Status      BookingStatus `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	CancelledAt *time.Time    `json:"cancelled_at,omitempty"`
}

func NewBooking(slotID string, roomID, userID uuid.UUID, startTime, endTime time.Time) (*Booking, error) {
	if slotID == "" {
		return nil, errors.New("slot ID is required")
	}

	if roomID == uuid.Nil {
		return nil, errors.New("room ID is required")
	}

	if userID == uuid.Nil {
		return nil, errors.New("user ID is required")
	}

	if startTime.IsZero() || endTime.IsZero() {
		return nil, errors.New("start time and end time are required")
	}

	if !endTime.After(startTime) {
		return nil, errors.New("end time must be after start time")
	}

	if endTime.Sub(startTime) != 30*time.Minute {
		return nil, errors.New("slot duration must be exactly 30 minutes")
	}

	if startTime.Before(time.Now().UTC()) {
		return nil, errors.New("cannot book slots in the past")
	}

	return &Booking{
		ID:        uuid.New(),
		SlotID:    slotID,
		RoomID:    roomID,
		UserID:    userID,
		StartTime: startTime.UTC(),
		EndTime:   endTime.UTC(),
		Status:    BookingStatusActive,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (b *Booking) Cancel() {
	now := time.Now().UTC()
	b.Status = BookingStatusCancelled
	b.CancelledAt = &now
}

func (b *Booking) IsActive() bool {
	return b.Status == BookingStatusActive
}

func (b *Booking) IsCancelled() bool {
	return b.Status == BookingStatusCancelled
}

func (b *Booking) IsInFuture() bool {
	return b.StartTime.After(time.Now().UTC())
}

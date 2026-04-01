package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Slot struct {
	ID        string     `json:"id"`
	RoomID    uuid.UUID  `json:"room_id"`
	StartTime time.Time  `json:"start_time"`
	EndTime   time.Time  `json:"end_time"`
	IsBooked  bool       `json:"is_booked"`
	BookingID *uuid.UUID `json:"booking_id,omitempty"`
}

func GenerateSlotID(roomID uuid.UUID, startTime time.Time) string {
	namespace := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

	name := fmt.Sprintf("%s:%d", roomID.String(), startTime.Unix())

	slotUUID := uuid.NewSHA1(namespace, []byte(name))

	return slotUUID.String()
}

func NewSlot(roomID uuid.UUID, startTime, endTime time.Time) *Slot {
	return &Slot{
		ID:        GenerateSlotID(roomID, startTime),
		RoomID:    roomID,
		StartTime: startTime.UTC(),
		EndTime:   endTime.UTC(),
		IsBooked:  false,
	}
}

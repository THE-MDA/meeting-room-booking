package domain

import (
	"time"

	"github.com/google/uuid"
)

type TimeSlot struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

func (ts TimeSlot) ToSlot(roomID uuid.UUID) *Slot {
	return &Slot{
		ID:        GenerateSlotID(roomID, ts.StartTime),
		RoomID:    roomID,
		StartTime: ts.StartTime,
		EndTime:   ts.EndTime,
		IsBooked:  false,
	}
}

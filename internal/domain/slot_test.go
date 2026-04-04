package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGenerateSlotID(t *testing.T) {
	roomID := uuid.New()
	time1 := time.Date(2026, 4, 10, 9, 0, 0, 0, time.UTC)
	time2 := time.Date(2026, 4, 10, 9, 30, 0, 0, time.UTC)

	id1 := GenerateSlotID(roomID, time1)
	id2 := GenerateSlotID(roomID, time1)
	id3 := GenerateSlotID(roomID, time2)

	if id1 != id2 {
		t.Error("GenerateSlotID() not deterministic for same input")
	}

	if id1 == id3 {
		t.Error("GenerateSlotID() produced same ID for different times")
	}

	roomID2 := uuid.New()
	id4 := GenerateSlotID(roomID2, time1)
	if id1 == id4 {
		t.Error("GenerateSlotID() produced same ID for different rooms")
	}
}

func TestNewSlot(t *testing.T) {
	roomID := uuid.New()
	startTime := time.Date(2026, 4, 10, 9, 0, 0, 0, time.UTC)
	endTime := startTime.Add(30 * time.Minute)

	slot := NewSlot(roomID, startTime, endTime)

	if slot == nil {
		t.Fatal("NewSlot() returned nil slot")
	}

	if slot.ID == "" {
		t.Error("Slot.ID should not be empty")
	}

	if slot.RoomID != roomID {
		t.Errorf("Slot.RoomID = %v, want %v", slot.RoomID, roomID)
	}

	if !slot.StartTime.Equal(startTime) {
		t.Errorf("Slot.StartTime = %v, want %v", slot.StartTime, startTime)
	}

	if slot.IsBooked {
		t.Error("New slot should not be booked")
	}
}

func TestTimeSlotToSlot(t *testing.T) {
	roomID := uuid.New()
	startTime := time.Date(2026, 4, 10, 9, 0, 0, 0, time.UTC)
	endTime := startTime.Add(30 * time.Minute)

	timeSlot := TimeSlot{
		StartTime: startTime,
		EndTime:   endTime,
	}

	slot := timeSlot.ToSlot(roomID)

	if slot.ID == "" {
		t.Error("ToSlot() should generate ID")
	}

	if !slot.StartTime.Equal(startTime) {
		t.Errorf("ToSlot() StartTime = %v, want %v", slot.StartTime, startTime)
	}

	if !slot.EndTime.Equal(endTime) {
		t.Errorf("ToSlot() EndTime = %v, want %v", slot.EndTime, endTime)
	}
}

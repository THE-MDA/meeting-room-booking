package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewBooking(t *testing.T) {
	roomID := uuid.New()
	userID := uuid.New()
	slotID := "test-slot-id"
	now := time.Now().UTC()
	futureTime := now.Add(24 * time.Hour)

	tests := []struct {
		name      string
		slotID    string
		roomID    uuid.UUID
		userID    uuid.UUID
		startTime time.Time
		endTime   time.Time
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid booking",
			slotID:    slotID,
			roomID:    roomID,
			userID:    userID,
			startTime: futureTime,
			endTime:   futureTime.Add(30 * time.Minute),
			wantErr:   false,
		},
		{
			name:      "empty slot ID",
			slotID:    "",
			roomID:    roomID,
			userID:    userID,
			startTime: futureTime,
			endTime:   futureTime.Add(30 * time.Minute),
			wantErr:   true,
			errMsg:    "slot ID is required",
		},
		{
			name:      "wrong duration (60 min)",
			slotID:    slotID,
			roomID:    roomID,
			userID:    userID,
			startTime: futureTime,
			endTime:   futureTime.Add(60 * time.Minute),
			wantErr:   true,
			errMsg:    "slot duration must be exactly 30 minutes",
		},
		{
			name:      "past slot",
			slotID:    slotID,
			roomID:    roomID,
			userID:    userID,
			startTime: now.Add(-24 * time.Hour),
			endTime:   now.Add(-24*time.Hour + 30*time.Minute),
			wantErr:   true,
			errMsg:    "cannot book slots in the past",
		},
		{
			name:      "zero room ID",
			slotID:    slotID,
			roomID:    uuid.Nil,
			userID:    userID,
			startTime: futureTime,
			endTime:   futureTime.Add(30 * time.Minute),
			wantErr:   true,
			errMsg:    "room ID is required",
		},
		{
			name:      "zero user ID",
			slotID:    slotID,
			roomID:    roomID,
			userID:    uuid.Nil,
			startTime: futureTime,
			endTime:   futureTime.Add(30 * time.Minute),
			wantErr:   true,
			errMsg:    "user ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			booking, err := NewBooking(tt.slotID, tt.roomID, tt.userID, tt.startTime, tt.endTime)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewBooking() expected error but got nil")
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("NewBooking() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewBooking() unexpected error: %v", err)
			}

			if booking == nil {
				t.Fatal("NewBooking() returned nil booking")
			}

			if booking.Status != BookingStatusActive {
				t.Errorf("Booking.Status = %v, want %v", booking.Status, BookingStatusActive)
			}
		})
	}
}

func TestBookingCancel(t *testing.T) {
	roomID := uuid.New()
	userID := uuid.New()
	futureTime := time.Now().UTC().Add(24 * time.Hour)

	booking, _ := NewBooking("slot-id", roomID, userID, futureTime, futureTime.Add(30*time.Minute))

	if !booking.IsActive() {
		t.Error("New booking should be active")
	}

	booking.Cancel()

	if !booking.IsCancelled() {
		t.Error("Booking should be cancelled after Cancel()")
	}

	if booking.IsActive() {
		t.Error("Booking should not be active after Cancel()")
	}

	if booking.CancelledAt == nil {
		t.Error("CancelledAt should be set after Cancel()")
	}
}

func TestBookingIsInFuture(t *testing.T) {
	roomID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC()

	futureBooking, err := NewBooking("slot-id", roomID, userID, now.Add(24*time.Hour), now.Add(24*time.Hour+30*time.Minute))
	if err != nil {
		t.Fatalf("Failed to create future booking: %v", err)
	}

	if !futureBooking.IsInFuture() {
		t.Error("Future booking should return true for IsInFuture()")
	}

	pastBooking, err := NewBooking("slot-id", roomID, userID, now.Add(-48*time.Hour), now.Add(-47*time.Hour+30*time.Minute))
	if err == nil {
		if pastBooking.IsInFuture() {
			t.Error("Past booking should return false for IsInFuture()")
		}
	} else {
		t.Logf("Past booking correctly rejected: %v", err)
	}
}

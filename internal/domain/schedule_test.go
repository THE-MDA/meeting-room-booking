package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewSchedule(t *testing.T) {
	roomID := uuid.New()

	tests := []struct {
		name      string
		roomID    uuid.UUID
		dayOfWeek DayOfWeek
		startTime string
		endTime   string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid schedule 9-17",
			roomID:    roomID,
			dayOfWeek: Monday,
			startTime: "09:00:00",
			endTime:   "17:00:00",
			wantErr:   false,
		},
		{
			name:      "valid schedule 9-17:30 (multiple of 30)",
			roomID:    roomID,
			dayOfWeek: Tuesday,
			startTime: "09:00:00",
			endTime:   "17:30:00",
			wantErr:   false,
		},
		{
			name:      "invalid - not multiple of 30",
			roomID:    roomID,
			dayOfWeek: Monday,
			startTime: "09:00:00",
			endTime:   "17:15:00",
			wantErr:   true,
			errMsg:    "time range must be a multiple of 30 minutes",
		},
		{
			name:      "invalid - end before start",
			roomID:    roomID,
			dayOfWeek: Monday,
			startTime: "17:00:00",
			endTime:   "09:00:00",
			wantErr:   true,
			errMsg:    "end time must be after start time",
		},
		{
			name:      "invalid - wrong time format",
			roomID:    roomID,
			dayOfWeek: Monday,
			startTime: "09:00",
			endTime:   "17:00",
			wantErr:   true,
			errMsg:    "invalid start time format, expected HH:MM:SS",
		},
		{
			name:      "invalid - zero room ID",
			roomID:    uuid.Nil,
			dayOfWeek: Monday,
			startTime: "09:00:00",
			endTime:   "17:00:00",
			wantErr:   true,
			errMsg:    "room ID is required",
		},
		{
			name:      "invalid - duration less than 30 min",
			roomID:    roomID,
			dayOfWeek: Monday,
			startTime: "09:00:00",
			endTime:   "09:15:00",
			wantErr:   true,
			errMsg:    "time range must be a multiple of 30 minutes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schedule, err := NewSchedule(tt.roomID, tt.dayOfWeek, tt.startTime, tt.endTime)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewSchedule() expected error but got nil")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("NewSchedule() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewSchedule() unexpected error: %v", err)
				return
			}

			if schedule == nil {
				t.Fatal("NewSchedule() returned nil schedule")
			}
		})
	}
}

func TestGenerateSlotsForDate(t *testing.T) {
	roomID := uuid.New()
	schedule, err := NewSchedule(roomID, Monday, "09:00:00", "17:00:00")
	if err != nil {
		t.Fatalf("Failed to create schedule: %v", err)
	}

	now := time.Now().UTC()
	daysUntilMonday := (time.Monday - now.Weekday() + 7) % 7
	if daysUntilMonday == 0 {
		daysUntilMonday = 7
	}
	nextMonday := now.AddDate(0, 0, int(daysUntilMonday))

	date := time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day(), 0, 0, 0, 0, time.UTC)

	slots := schedule.GenerateSlotsForDate(date)

	expectedSlots := 16
	if len(slots) != expectedSlots {
		t.Errorf("GenerateSlotsForDate() returned %d slots, expected %d", len(slots), expectedSlots)
	}

	if len(slots) > 0 {
		expectedStart := time.Date(date.Year(), date.Month(), date.Day(), 9, 0, 0, 0, time.UTC)
		if !slots[0].StartTime.Equal(expectedStart) {
			t.Errorf("First slot start time = %v, expected %v", slots[0].StartTime, expectedStart)
		}

		expectedEnd := time.Date(date.Year(), date.Month(), date.Day(), 9, 30, 0, 0, time.UTC)
		if !slots[0].EndTime.Equal(expectedEnd) {
			t.Errorf("First slot end time = %v, expected %v", slots[0].EndTime, expectedEnd)
		}
	}
}

func TestGenerateSlotsForDateRange(t *testing.T) {
	roomID := uuid.New()
	schedule, err := NewSchedule(roomID, Monday, "09:00:00", "17:00:00")
	if err != nil {
		t.Fatalf("Failed to create schedule: %v", err)
	}

	startDate := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 4, 12, 0, 0, 0, 0, time.UTC)

	slots := schedule.GenerateSlotsForDateRange(startDate, endDate)

	expectedSlots := 16
	if len(slots) != expectedSlots {
		t.Errorf("GenerateSlotsForDateRange() returned %d slots, expected %d", len(slots), expectedSlots)
	}
}

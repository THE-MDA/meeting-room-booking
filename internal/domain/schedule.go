package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type DayOfWeek int

const (
	Monday    DayOfWeek = 0
	Tuesday   DayOfWeek = 1
	Wednesday DayOfWeek = 2
	Thursday  DayOfWeek = 3
	Friday    DayOfWeek = 4
	Saturday  DayOfWeek = 5
	Sunday    DayOfWeek = 6
)

type Schedule struct {
	ID        uuid.UUID `json:"id"`
	RoomID    uuid.UUID `json:"room_id"`
	DayOfWeek DayOfWeek `json:"day_of_week"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time"`
	CreatedAt time.Time `json:"created_at"`
}

func NewSchedule(roomID uuid.UUID, dayOfWeek DayOfWeek, startTime, endTime string) (*Schedule, error) {
	if roomID == uuid.Nil {
		return nil, errors.New("room ID is required")
	}

	if dayOfWeek < Monday || dayOfWeek > Sunday {
		return nil, errors.New("invalid day of week")
	}

	start, err := time.Parse("15:04:05", startTime)
	if err != nil {
		return nil, errors.New("invalid start time format, expected HH:MM:SS")
	}

	end, err := time.Parse("15:04:05", endTime)
	if err != nil {
		return nil, errors.New("invalid end time format, expected HH:MM:SS")
	}

	if !end.After(start) {
		return nil, errors.New("end time must be after start time")
	}

	duration := end.Sub(start)
	if duration%30 != 0 {
		return nil, errors.New("time range must be a multiple of 30 minutes")
	}

	if duration < 30*time.Minute {
		return nil, errors.New("time range must be at least 30 minutes")
	}

	return &Schedule{
		ID:        uuid.New(),
		RoomID:    roomID,
		DayOfWeek: dayOfWeek,
		StartTime: startTime,
		EndTime:   endTime,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (s *Schedule) GenerateSlotsForDate(date time.Time) []TimeSlot {
	expectedWeekday := (int(s.DayOfWeek) + 1) % 7
	if int(date.Weekday()) != expectedWeekday {
		return nil
	}

	start, _ := time.Parse("15:04:05", s.StartTime)
	end, _ := time.Parse("15:04:05", s.EndTime)

	var slots []TimeSlot

	current := time.Date(date.Year(), date.Month(), date.Day(),
		start.Hour(), start.Minute(), 0, 0, time.UTC)

	endTime := time.Date(date.Year(), date.Month(), date.Day(),
		end.Hour(), end.Minute(), 0, 0, time.UTC)

	slotDuration := 30 * time.Minute

	for current.Add(slotDuration).Before(endTime) || current.Add(slotDuration).Equal(endTime) {
		slotStart := current
		slotEnd := current.Add(slotDuration)

		slots = append(slots, TimeSlot{
			StartTime: slotStart,
			EndTime:   slotEnd,
		})

		current = current.Add(slotDuration)
	}

	return slots
}

func (s *Schedule) GenerateSlotsForDateRange(startDate, endDate time.Time) []TimeSlot {
	var allSlots []TimeSlot

	for date := startDate; date.Before(endDate) || date.Equal(endDate); date = date.AddDate(0, 0, 1) {
		slots := s.GenerateSlotsForDate(date)
		allSlots = append(allSlots, slots...)
	}

	return allSlots
}

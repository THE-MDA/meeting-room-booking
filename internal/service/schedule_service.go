package service

import (
	"context"
	"errors"
	"meeting-room-booking/internal/domain"

	"github.com/google/uuid"
)

type ScheduleService struct {
	scheduleRepo domain.ScheduleRepository
	roomRepo     domain.RoomRepository
}

func NewScheduleService(scheduleRepo domain.ScheduleRepository, roomRepo domain.RoomRepository) *ScheduleService {
	return &ScheduleService{
		scheduleRepo: scheduleRepo,
		roomRepo:     roomRepo,
	}
}

func (s *ScheduleService) CreateSchedule(ctx context.Context, roomID uuid.UUID,
	dayOfWeek domain.DayOfWeek, startTime string, endTime string) (*domain.Schedule, error) {
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	if room == nil {
		return nil, errors.New("room not found")
	}

	exists, err := s.scheduleRepo.Exists(ctx, roomID, dayOfWeek)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("schedule already exists for this room on this day")
	}

	schedule, err := domain.NewSchedule(roomID, dayOfWeek, startTime, endTime)
	if err != nil {
		return nil, err
	}

	if err := s.scheduleRepo.Create(ctx, schedule); err != nil {
		return nil, err
	}

	return schedule, nil
}

func (s *ScheduleService) GetSchedulesByRoom(ctx context.Context, roomID uuid.UUID) ([]*domain.Schedule, error) {
	return s.scheduleRepo.GetByRoomID(ctx, roomID)
}

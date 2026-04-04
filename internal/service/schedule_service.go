package service

import (
	"context"
	"errors"
	"log/slog"
	"meeting-room-booking/internal/domain"

	"github.com/google/uuid"
)

type ScheduleService struct {
	scheduleRepo domain.ScheduleRepository
	roomRepo     domain.RoomRepository
	slotRepo     domain.SlotRepository
}

func NewScheduleService(scheduleRepo domain.ScheduleRepository, roomRepo domain.RoomRepository, slotRepo domain.SlotRepository) *ScheduleService {
	return &ScheduleService{
		scheduleRepo: scheduleRepo,
		roomRepo:     roomRepo,
		slotRepo:     slotRepo,
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

	// Получаем все расписания для комнаты
	schedules, err := s.scheduleRepo.GetByRoomID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	// Генерируем слоты
	if err := s.slotRepo.GenerateAndSaveSlotsForRoom(ctx, roomID, schedules, 90); err != nil {
		// Логируем ошибку, но не отменяем создание расписания
		slog.Error("Failed to generate slots", "error", err)
	}

	return schedule, nil
}

func (s *ScheduleService) GetSchedulesByRoom(ctx context.Context, roomID uuid.UUID) ([]*domain.Schedule, error) {
	return s.scheduleRepo.GetByRoomID(ctx, roomID)
}

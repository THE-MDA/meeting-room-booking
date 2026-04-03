package service

import (
	"context"
	"errors"
	"meeting-room-booking/internal/domain"

	"github.com/google/uuid"
)

type RoomService struct {
	roomRepo domain.RoomRepository
}

func NewRoomService(roomRepo domain.RoomRepository) *RoomService {
	return &RoomService{
		roomRepo: roomRepo,
	}
}

func (s *RoomService) CreateRoom(ctx context.Context, name string, description string, capacity int) (*domain.Room, error) {
	room, err := domain.NewRoom(name, description, capacity)
	if err != nil {
		return nil, err
	}

	if err := s.roomRepo.Create(ctx, room); err != nil {
		return nil, err
	}

	return room, nil
}

func (s *RoomService) GetRoomByID(ctx context.Context, id uuid.UUID) (*domain.Room, error) {
	room, err := s.roomRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if room == nil {
		return nil, errors.New("room not found")
	}

	return room, nil
}

func (s *RoomService) GetAllRooms(ctx context.Context) ([]*domain.Room, error) {
	return s.roomRepo.GetAll(ctx)
}

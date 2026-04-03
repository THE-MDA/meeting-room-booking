package service

import (
	"context"
	"errors"
	"meeting-room-booking/internal/domain"
	"meeting-room-booking/internal/repository"
	"time"

	"github.com/google/uuid"
)

type BookingService struct {
	bookingRepo  *repository.BookingRepository
	roomRepo     *repository.RoomRepository
	scheduleRepo *repository.ScheduleRepository
}

func NewBookingService(bookingRepo *repository.BookingRepository, roomRepo *repository.RoomRepository,
	scheduleRepo *repository.ScheduleRepository) *BookingService {
	return &BookingService{
		bookingRepo:  bookingRepo,
		roomRepo:     roomRepo,
		scheduleRepo: scheduleRepo,
	}
}

func (s *BookingService) GetAvailableSlots(ctx context.Context, roomID uuid.UUID,
	date time.Time) ([]*domain.Slot, error) {
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	if room == nil {
		return nil, errors.New("room not found")
	}

	schedules, err := s.scheduleRepo.GetByRoomID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	if len(schedules) == 0 {
		return []*domain.Slot{}, nil
	}

	var allTimeSlots []domain.TimeSlot
	for _, schedule := range schedules {
		slots := schedule.GenerateSlotsForDate(date)
		allTimeSlots = append(allTimeSlots, slots...)
	}

	if len(allTimeSlots) == 0 {
		return []*domain.Slot{}, nil
	}

	slots := make([]*domain.Slot, len(allTimeSlots))
	slotIDs := make([]string, len(allTimeSlots))
	for i, ts := range allTimeSlots {
		slot := ts.ToSlot(roomID)
		slots[i] = slot
		slotIDs[i] = slot.ID
	}

	bookedSlots, err := s.bookingRepo.GetBookedSlots(ctx, slotIDs)
	if err != nil {
		return nil, err
	}

	for _, slot := range slots {
		if bookedSlots[slot.ID] {
			slot.IsBooked = true
		}
	}

	return slots, nil
}

func (s *BookingService) CreateBooking(ctx context.Context, userID uuid.UUID, slotID string,
	roomID uuid.UUID, startTime time.Time, endTime time.Time) (*domain.Booking, error) {
	if startTime.Before(time.Now().UTC()) {
		return nil, errors.New("cannot book slots in the past")
	}

	isBooked, err := s.bookingRepo.IsSlotBooked(ctx, slotID)
	if err != nil {
		return nil, err
	}

	if isBooked {
		return nil, errors.New("slot is already booked")
	}

	booking, err := domain.NewBooking(slotID, roomID, userID, startTime, endTime)
	if err != nil {
		return nil, err
	}

	return booking, nil
}

func (s *BookingService) CanceledBooking(ctx context.Context, bookingID uuid.UUID) error {
	return s.bookingRepo.Cancel(ctx, bookingID)
}

func (s *BookingService) GetUserBookings(ctx context.Context, userID uuid.UUID) ([]*domain.Booking, error) {
	return s.bookingRepo.GetByUserID(ctx, userID, true)
}

func (s *BookingService) GetAllBookings(ctx context.Context, limit, offset int) ([]*domain.Booking, error) {
	if limit <= 0 {
		limit = 20
	}

	if offset < 0 {
		offset = 0
	}

	return s.bookingRepo.GetAll(ctx, limit, offset)
}

func (s *BookingService) GetBookingByID(ctx context.Context, bookingID uuid.UUID) (*domain.Booking, error) {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	if booking == nil {
		return nil, errors.New("booking not found")
	}

	return booking, nil
}

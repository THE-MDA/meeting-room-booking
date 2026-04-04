package service

import (
	"context"
	"errors"
	"meeting-room-booking/internal/domain"
	"time"

	"github.com/google/uuid"
)

type BookingService struct {
	bookingRepo  domain.BookingRepository
	slotRepo     domain.SlotRepository
	roomRepo     domain.RoomRepository
	scheduleRepo domain.ScheduleRepository
}

func NewBookingService(bookingRepo domain.BookingRepository, slotRepo domain.SlotRepository, roomRepo domain.RoomRepository, scheduleRepo domain.ScheduleRepository) *BookingService{
	return &BookingService{
		bookingRepo:  bookingRepo,
		slotRepo:     slotRepo,
		roomRepo:     roomRepo,
		scheduleRepo: scheduleRepo,
	}
}

func (s *BookingService) GetAvailableSlots(ctx context.Context, roomID uuid.UUID, date time.Time) ([]*domain.Slot, error) {
	return s.slotRepo.GetAvailableSlots(ctx, roomID, date)
}

func (s *BookingService) CreateBooking(ctx context.Context, userID uuid.UUID, slotID string) (*domain.Booking, error) {
	slotUUID, err := uuid.Parse(slotID)
	if err != nil {
		return nil, errors.New("invalid slot id")
	}
	slot, err := s.slotRepo.GetByID(ctx, slotUUID)
	if err != nil {
		return nil, err
	}
	if slot == nil {
		return nil, errors.New("slot not found")
	}
	if slot.StartTime.Before(time.Now().UTC()) {
		return nil, errors.New("cannot book slots in the past")
	}
	isBooked, err := s.bookingRepo.IsSlotBooked(ctx, slotID)
	if err != nil {
		return nil, err
	}
	if isBooked {
		return nil, errors.New("slot is already booked")
	}
	booking, err := domain.NewBooking(slotID, slot.RoomID, userID, slot.StartTime, slot.EndTime)
	if err != nil {
		return nil, err
	}
	if err := s.bookingRepo.Create(ctx, booking); err != nil {
		return nil, err
	}
	return booking, nil
}

func (s *BookingService) CancelBooking(ctx context.Context, bookingID uuid.UUID) error {
	return s.bookingRepo.Cancel(ctx, bookingID)
}

func (s *BookingService) GetUserBookings(ctx context.Context, userID uuid.UUID) ([]*domain.Booking, error) {
    bookings, err := s.bookingRepo.GetByUserID(ctx, userID, false)
    if err != nil {
        return nil, err
    }
    
    var futureActiveBookings []*domain.Booking
    now := time.Now().UTC()
    for _, booking := range bookings {
        if booking.Status == domain.BookingStatusActive && booking.StartTime.After(now) {
            futureActiveBookings = append(futureActiveBookings, booking)
        }
    }
    
    return futureActiveBookings, nil
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

func (s *BookingService) CountBookings(ctx context.Context) (int, error) {
    return s.bookingRepo.Count(ctx)
}
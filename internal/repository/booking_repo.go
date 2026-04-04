package repository

import (
	"context"
	"database/sql"
	"fmt"
	"meeting-room-booking/internal/domain"
	"time"

	"github.com/google/uuid"
)

type BookingRepository struct {
	db *DB
}

func NewBookingRepository(db *DB) *BookingRepository {
	return &BookingRepository{
		db: db,
	}
}

func (r *BookingRepository) Create(ctx context.Context, booking *domain.Booking) error {
	query := `INSERT INTO bookings (id, slot_id, room_id, user_id, start_time, end_time, status, created_at)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query,
		booking.ID, booking.SlotID, booking.RoomID, booking.UserID,
		booking.StartTime, booking.EndTime, booking.Status, booking.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create booking: %w", err)
	}

	return nil
}

func (r *BookingRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error) {
	query := `SELECT id, slot_id, room_id, user_id, start_time, end_time, status, created_at, cancelled_at
    FROM bookings WHERE id = $1`

	var booking domain.Booking
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&booking.ID,
		&booking.SlotID,
		&booking.RoomID,
		&booking.UserID,
		&booking.StartTime,
		&booking.EndTime,
		&booking.Status,
		&booking.CreatedAt,
		&booking.CancelledAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		fmt.Printf("DEBUG GetByID: scan error=%v\n", err)
		return nil, fmt.Errorf("failed to get booking by ID: %w", err)
	}

	return &booking, nil
}

func (r *BookingRepository) GetBySlotID(ctx context.Context, slotID string) (*domain.Booking, error) {
	query := `SELECT id, slot_id, room_id, user_id, start_time, end_time, status, created_at, cancelled_at
	FROM bookings WHERE slot_id = $1 AND status = 'active'`

	var booking domain.Booking
	err := r.db.QueryRowContext(ctx, query, slotID).Scan(
		&booking.ID,
		&booking.SlotID,
		&booking.RoomID,
		&booking.UserID,
		&booking.StartTime,
		&booking.EndTime,
		&booking.Status,
		&booking.CreatedAt,
		&booking.CancelledAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to get booking by slot ID: %w", err)
	}

	return &booking, nil
}

func (r *BookingRepository) GetByUserID(ctx context.Context, userID uuid.UUID, onlyFuture bool) ([]*domain.Booking, error) {
	query := `SELECT id, slot_id, room_id, user_id, start_time, end_time, status, created_at, cancelled_at
	FROM bookings WHERE user_id = $1`

	args := []interface{}{userID}

	if onlyFuture {
		query += ` AND start_time > $2`
		args = append(args, time.Now().UTC())
	}

	query += ` ORDER BY start_time ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get bookings by user ID: %w", err)
	}
	defer rows.Close()

	var bookings []*domain.Booking
	for rows.Next() {
		var booking domain.Booking
		if err := rows.Scan(&booking.ID,
			&booking.SlotID, &booking.RoomID, &booking.UserID, &booking.StartTime,
			&booking.EndTime, &booking.Status, &booking.CreatedAt, &booking.CancelledAt); err != nil {
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}
		bookings = append(bookings, &booking)
	}
	return bookings, rows.Err()
}

func (r *BookingRepository) GetAll(ctx context.Context, limit int, offset int) ([]*domain.Booking, error) {
	query := `SELECT id, slot_id, room_id, user_id, start_time, end_time, status, created_at, cancelled_at
	FROM bookings ORDER BY created_at DESC
	LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get all bookings: %w", err)
	}
	defer rows.Close()

	var bookings []*domain.Booking
	for rows.Next() {
		var booking domain.Booking
		if err := rows.Scan(&booking.ID,
			&booking.SlotID, &booking.RoomID, &booking.UserID, &booking.StartTime,
			&booking.EndTime, &booking.Status, &booking.CreatedAt, &booking.CancelledAt); err != nil {
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}
		bookings = append(bookings, &booking)
	}
	return bookings, rows.Err()
}

func (r *BookingRepository) Update(ctx context.Context, booking *domain.Booking) error {
	query := `UPDATE bookings SET status = $1, cancelled_at = $2 WHERE id = $3`

	_, err := r.db.ExecContext(ctx, query, booking.Status, booking.CancelledAt, booking.ID)
	if err != nil {
		return fmt.Errorf("failed to update booking: %w", err)
	}

	return nil
}

func (r *BookingRepository) Cancel(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE bookings SET status = 'cancelled', cancelled_at = $1
	WHERE id = $2 AND status = 'active'`

	result, err := r.db.ExecContext(ctx, query, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to cancel booking: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil
	}

	return nil
}

func (r *BookingRepository) IsSlotBooked(ctx context.Context, slotID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM bookings WHERE slot_id = $1 AND status = 'active')`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, slotID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check slot booked: %w", err)
	}
	return exists, nil
}

func (r *BookingRepository) GetBookedSlots(ctx context.Context, slotIDs []string) (map[string]bool, error) {
	if len(slotIDs) == 0 {
		return make(map[string]bool), nil
	}

	query := `SELECT slot_id FROM bookings WHERE slot_id = ANY($1) AND status = 'active'`

	rows, err := r.db.QueryContext(ctx, query, slotIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get booked slots: %w", err)
	}
	defer rows.Close()

	booked := make(map[string]bool)
	for rows.Next() {
		var slotID string
		if err := rows.Scan(&slotID); err != nil {
			return nil, fmt.Errorf("failed to scan booked slot: %w", err)
		}
		booked[slotID] = true
	}

	return booked, rows.Err()
}

func (r *BookingRepository) Count(ctx context.Context) (int, error) {
    var count int
    err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM bookings").Scan(&count)
    return count, err
}
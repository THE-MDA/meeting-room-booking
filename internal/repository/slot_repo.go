package repository

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"meeting-room-booking/internal/domain"

	"github.com/google/uuid"
)

type SlotRepository struct {
	db *DB
}

func NewSlotRepository(db *DB) *SlotRepository {
	return &SlotRepository{db: db}
}

func (r *SlotRepository) Create(ctx context.Context, slot *domain.Slot) error {
	query := `INSERT INTO slots (id, room_id, start_time, end_time) VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, slot.ID, slot.RoomID, slot.StartTime, slot.EndTime)
	return err
}

func (r *SlotRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Slot, error) {
	query := `SELECT id, room_id, start_time, end_time FROM slots WHERE id = $1`
	var slot domain.Slot
	err := r.db.QueryRowContext(ctx, query, id).Scan(&slot.ID, &slot.RoomID, &slot.StartTime, &slot.EndTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &slot, nil
}

func (r *SlotRepository) GetAvailableSlots(ctx context.Context, roomID uuid.UUID, date time.Time) ([]*domain.Slot, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	query := `
        SELECT s.id, s.room_id, s.start_time, s.end_time
        FROM slots s
        LEFT JOIN bookings b ON b.slot_id = s.id::text AND b.status = 'active'
        WHERE s.room_id = $1 AND s.start_time >= $2 AND s.start_time < $3
          AND b.id IS NULL
        ORDER BY s.start_time
    `
	rows, err := r.db.QueryContext(ctx, query, roomID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var slots []*domain.Slot
	for rows.Next() {
		var slot domain.Slot
		if err := rows.Scan(&slot.ID, &slot.RoomID, &slot.StartTime, &slot.EndTime); err != nil {
			return nil, err
		}
		slots = append(slots, &slot)
	}
	return slots, nil
}

func (r *SlotRepository) GenerateAndSaveSlotsForRoom(ctx context.Context, roomID uuid.UUID, schedules []*domain.Schedule, daysForward int) error {
	if len(schedules) == 0 {
		slog.Warn("No schedules to generate slots from")
		return nil
	}

	now := time.Now().UTC()
	endDate := now.AddDate(0, 0, daysForward)
	var slots []*domain.Slot

	for _, schedule := range schedules {
		for date := now; date.Before(endDate); date = date.AddDate(0, 0, 1) {
			timeSlots := schedule.GenerateSlotsForDate(date)
			for _, ts := range timeSlots {
				slot := ts.ToSlot(roomID)
				slots = append(slots, slot)
			}
		}
	}

	slog.Info("Generating slots", "count", len(slots), "room_id", roomID)
	inserted := 0
    for _, slot := range slots {
        query := `INSERT INTO slots (id, room_id, start_time, end_time) VALUES ($1, $2, $3, $4) ON CONFLICT (id) DO NOTHING`
        result, err := r.db.ExecContext(ctx, query, slot.ID, slot.RoomID, slot.StartTime, slot.EndTime)
        if err != nil {
            return err
        }
        rowsAffected, _ := result.RowsAffected()
        if rowsAffected > 0 {
            inserted++
        }
    }
    slog.Info("Slots inserted", "inserted", inserted, "total", len(slots))
    return nil
}

package repository

import (
	"context"
	"fmt"
	"meeting-room-booking/internal/domain"
	"time"

	"github.com/google/uuid"
)

type ScheduleRepository struct {
	db *DB
}

func NewScheduleRepository(db *DB) *ScheduleRepository {
	return &ScheduleRepository{
		db: db,
	}
}

func (r *ScheduleRepository) Create(ctx context.Context, schedule *domain.Schedule) error {
	query := `INSERT INTO schedules (id, room_id, day_of_week, start_time, end_time, created_at) 
    VALUES ($1, $2, $3, $4, $5, $6)`

	startTime, _ := time.Parse("15:04:05", schedule.StartTime)
	endTime, _ := time.Parse("15:04:05", schedule.EndTime)

	_, err := r.db.ExecContext(ctx, query,
		schedule.ID,
		schedule.RoomID,
		schedule.DayOfWeek,
		startTime,
		endTime,
		schedule.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create schedule: %w", err)
	}

	return nil
}

func (r *ScheduleRepository) GetByRoomID(ctx context.Context, roomID uuid.UUID) ([]*domain.Schedule, error) {
	query := `SELECT id, room_id, day_of_week, start_time, end_time, created_at 
    FROM schedules WHERE room_id = $1 ORDER BY day_of_week ASC`

	rows, err := r.db.QueryContext(ctx, query, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedules by room ID: %w", err)
	}
	defer rows.Close()

	var schedules []*domain.Schedule

	for rows.Next() {
		var schedule domain.Schedule
		var startTime, endTime time.Time

		err := rows.Scan(
			&schedule.ID,
			&schedule.RoomID,
			&schedule.DayOfWeek,
			&startTime,
			&endTime,
			&schedule.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schedule: %w", err)
		}

		schedule.StartTime = startTime.Format("15:04:05")
		schedule.EndTime = endTime.Format("15:04:05")

		schedules = append(schedules, &schedule)
	}

	return schedules, rows.Err()
}

func (r *ScheduleRepository) Exists(ctx context.Context, roomID uuid.UUID, dayOfWeek domain.DayOfWeek) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM schedules WHERE room_id = $1 AND day_of_week = $2)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, roomID, dayOfWeek).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check schedule existence: %w", err)
	}

	return exists, nil
}

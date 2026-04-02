package repository

import (
	"context"
	"database/sql"
	"fmt"
	"meeting-room-booking/internal/domain"

	"github.com/google/uuid"
)

type RoomRepository struct {
	db *DB
}

func NewRoomRepository(db *DB) *RoomRepository {
	return &RoomRepository{
		db: db,
	}
}

func (r *RoomRepository) Create(ctx context.Context, room *domain.Room) error {
	query := `INSERT INTO rooms (id, name, description, capacity, created_at) VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.ExecContext(ctx, query, room.ID, room.Name, room.Description, room.Capacity, room.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create room: %w", err)
	}

	return nil
}

func (r *RoomRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Room, error) {
	query := `SELECT id, name, description, capacity, created_at FROM rooms WHERE id = $1`

	var room domain.Room

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&room.ID,
		&room.Name,
		&room.Description,
		&room.Capacity,
		&room.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to get room by ID: %w", err)
	}

	return &room, nil
}

func (r *RoomRepository) GetAll(ctx context.Context) ([]*domain.Room, error) {
	query := `SELECT id, name, description, capacity, created_at FROM rooms ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all rooms: %w", err)
	}
	defer rows.Close()

	var rooms []*domain.Room

	for rows.Next() {
		var room domain.Room
		if err := rows.Scan(&room.ID, &room.Name, &room.Description, &room.Capacity, &room.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan room: %w", err)
		}

		rooms = append(rooms, &room)
	}

	return rooms, nil
}

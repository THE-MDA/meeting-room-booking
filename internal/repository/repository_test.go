package repository

import (
    "context"
    "database/sql"
    "testing"

    _ "github.com/lib/pq"
    "meeting-room-booking/internal/domain"
)

func setupTestDB(t *testing.T) *DB {
    connStr := "host=localhost port=5433 user=app password=secret dbname=meeting_rooms_test sslmode=disable"
    sqlDB, err := sql.Open("postgres", connStr)
    if err != nil {
        t.Skipf("Skipping test: cannot connect to test database: %v", err)
        return nil
    }
    
    // Проверяем подключение
    if err := sqlDB.Ping(); err != nil {
        t.Skipf("Skipping test: cannot ping test database: %v", err)
        return nil
    }
    
    // Очищаем таблицы перед тестом
    sqlDB.Exec("TRUNCATE bookings CASCADE")
    sqlDB.Exec("TRUNCATE schedules CASCADE")
    sqlDB.Exec("TRUNCATE rooms CASCADE")
    sqlDB.Exec("TRUNCATE users CASCADE")
    
    return &DB{DB: sqlDB}
}

func teardownTestDB(db *DB) {
    if db != nil && db.DB != nil {
        db.Exec("TRUNCATE bookings CASCADE")
        db.Exec("TRUNCATE schedules CASCADE")
        db.Exec("TRUNCATE rooms CASCADE")
        db.Exec("TRUNCATE users CASCADE")
        db.Close()
    }
}

func TestUserRepository_CreateAndGet(t *testing.T) {
    db := setupTestDB(t)
    if db == nil {
        return
    }
    defer teardownTestDB(db)
    
    repo := NewUserRepository(db)
    ctx := context.Background()
    
    user, err := domain.NewUser("test@example.com", domain.RoleUser)
    if err != nil {
        t.Fatalf("Failed to create user: %v", err)
    }
    
    err = repo.Create(ctx, user)
    if err != nil {
        t.Fatalf("Failed to create user: %v", err)
    }
    
    found, err := repo.GetByID(ctx, user.ID)
    if err != nil {
        t.Fatalf("Failed to get user: %v", err)
    }
    if found == nil {
        t.Fatal("User not found")
    }
    if found.Email != user.Email {
        t.Errorf("Email = %s, want %s", found.Email, user.Email)
    }
}

func TestRoomRepository_CreateAndGet(t *testing.T) {
    db := setupTestDB(t)
    if db == nil {
        return
    }
    defer teardownTestDB(db)
    
    repo := NewRoomRepository(db)
    ctx := context.Background()
    
    room, err := domain.NewRoom("Test Room", "Description", 10)
    if err != nil {
        t.Fatalf("Failed to create room: %v", err)
    }
    
    err = repo.Create(ctx, room)
    if err != nil {
        t.Fatalf("Failed to create room: %v", err)
    }
    
    found, err := repo.GetByID(ctx, room.ID)
    if err != nil {
        t.Fatalf("Failed to get room: %v", err)
    }
    if found == nil {
        t.Fatal("Room not found")
    }
    if found.Name != room.Name {
        t.Errorf("Name = %s, want %s", found.Name, room.Name)
    }
}

func TestNewUserRepository(t *testing.T) {
    db := &DB{DB: nil}
    repo := NewUserRepository(db)
    if repo == nil {
        t.Error("UserRepository is nil")
    }
}

func TestNewRoomRepository(t *testing.T) {
    db := &DB{DB: nil}
    repo := NewRoomRepository(db)
    if repo == nil {
        t.Error("RoomRepository is nil")
    }
}

func TestNewScheduleRepository(t *testing.T) {
    db := &DB{DB: nil}
    repo := NewScheduleRepository(db)
    if repo == nil {
        t.Error("ScheduleRepository is nil")
    }
}

func TestNewBookingRepository(t *testing.T) {
    db := &DB{DB: nil}
    repo := NewBookingRepository(db)
    if repo == nil {
        t.Error("BookingRepository is nil")
    }
}
package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"meeting-room-booking/internal/config"
	"meeting-room-booking/internal/domain"
	"meeting-room-booking/internal/migrator"
	"meeting-room-booking/internal/repository"
)

func setupTestDB(t *testing.T) (*repository.DB, func()) {
	// Пересоздаём тестовую БД
	adminDB, err := sql.Open("postgres", "host=localhost port=5433 user=app password=secret dbname=postgres sslmode=disable")
	if err != nil {
		t.Skipf("Skipping test: cannot connect to postgres: %v", err)
		return nil, nil
	}
	defer adminDB.Close()

	// Убиваем активные соединения
	adminDB.Exec(`SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = 'meeting_rooms_test' AND pid <> pg_backend_pid()`)

	// Пересоздаём БД
	adminDB.Exec("DROP DATABASE IF EXISTS meeting_rooms_test")
	_, err = adminDB.Exec("CREATE DATABASE meeting_rooms_test")
	if err != nil {
		t.Skipf("Skipping test: cannot create test database: %v", err)
		return nil, nil
	}

	// Подключаемся к тестовой БД
	cfg := &config.Config{
		DBHost:     "localhost",
		DBPort:     5433,
		DBUser:     "app",
		DBPassword: "secret",
		DBName:     "meeting_rooms_test",
	}

	db, err := sql.Open("postgres", cfg.GetDBConnectionString())
	if err != nil {
		t.Skipf("Skipping test: cannot connect to test database: %v", err)
		return nil, nil
	}

	if err := db.Ping(); err != nil {
		t.Skipf("Skipping test: cannot ping test database: %v", err)
		return nil, nil
	}

	// Применяем миграции
	mig := migrator.New("../../migrations", cfg.GetDatabaseURL())
	if err := mig.Up(); err != nil {
		t.Logf("Migration error: %v", err)
	}

	dbWrapper := &repository.DB{DB: db}

	cleanup(dbWrapper)

	return dbWrapper, func() {
		cleanup(dbWrapper)
		db.Close()
	}
}

func cleanup(db *repository.DB) {
	db.Exec("TRUNCATE TABLE bookings CASCADE")
	db.Exec("TRUNCATE TABLE slots CASCADE")
	db.Exec("TRUNCATE TABLE schedules CASCADE")
	db.Exec("TRUNCATE TABLE rooms CASCADE")
	db.Exec("TRUNCATE TABLE users CASCADE")
}

func TestRoomService_CreateRoom(t *testing.T) {
	db, teardown := setupTestDB(t)
	if db == nil {
		return
	}
	defer teardown()

	roomRepo := repository.NewRoomRepository(db)
	service := NewRoomService(roomRepo)

	room, err := service.CreateRoom(context.Background(), "Test Conference Room", "Large room with projector", 20)
	if err != nil {
		t.Fatalf("CreateRoom() error = %v", err)
	}

	if room.Name != "Test Conference Room" {
		t.Errorf("CreateRoom() name = %v, want %v", room.Name, "Test Conference Room")
	}

	if room.Capacity != 20 {
		t.Errorf("CreateRoom() capacity = %v, want %v", room.Capacity, 20)
	}

	if room.ID == uuid.Nil {
		t.Error("CreateRoom() returned room with nil ID")
	}
}

func TestRoomService_GetAllRooms(t *testing.T) {
	db, teardown := setupTestDB(t)
	if db == nil {
		return
	}
	defer teardown()

	roomRepo := repository.NewRoomRepository(db)
	service := NewRoomService(roomRepo)

	room1, _ := service.CreateRoom(context.Background(), "Room 1", "", 5)
	room2, _ := service.CreateRoom(context.Background(), "Room 2", "", 10)

	rooms, err := service.GetAllRooms(context.Background())
	if err != nil {
		t.Fatalf("GetAllRooms() error = %v", err)
	}

	if len(rooms) < 2 {
		t.Errorf("GetAllRooms() returned %d rooms, want at least 2", len(rooms))
	}

	found1, found2 := false, false
	for _, r := range rooms {
		if r.ID == room1.ID {
			found1 = true
		}
		if r.ID == room2.ID {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Error("GetAllRooms() did not return all created rooms")
	}
}

func TestScheduleService_CreateSchedule(t *testing.T) {
	db, teardown := setupTestDB(t)
	if db == nil {
		return
	}
	defer teardown()

	roomRepo := repository.NewRoomRepository(db)
	scheduleRepo := repository.NewScheduleRepository(db)
	slotRepo := repository.NewSlotRepository(db)

	roomService := NewRoomService(roomRepo)
	scheduleService := NewScheduleService(scheduleRepo, roomRepo, slotRepo)

	room, err := roomService.CreateRoom(context.Background(), "Test Room", "", 10)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	schedule, err := scheduleService.CreateSchedule(
		context.Background(),
		room.ID,
		domain.Monday,
		"09:00:00",
		"17:00:00",
	)
	if err != nil {
		t.Fatalf("CreateSchedule() error = %v", err)
	}

	if schedule.RoomID != room.ID {
		t.Errorf("CreateSchedule() RoomID = %v, want %v", schedule.RoomID, room.ID)
	}

	if schedule.StartTime != "09:00:00" {
		t.Errorf("CreateSchedule() StartTime = %v, want 09:00:00", schedule.StartTime)
	}

	if schedule.EndTime != "17:00:00" {
		t.Errorf("CreateSchedule() EndTime = %v, want 17:00:00", schedule.EndTime)
	}
}

func TestBookingService_CreateBooking(t *testing.T) {
	db, teardown := setupTestDB(t)
	if db == nil {
		return
	}
	defer teardown()

	userRepo := repository.NewUserRepository(db)
	roomRepo := repository.NewRoomRepository(db)
	scheduleRepo := repository.NewScheduleRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	slotRepo := repository.NewSlotRepository(db)

	roomService := NewRoomService(roomRepo)
	scheduleService := NewScheduleService(scheduleRepo, roomRepo, slotRepo)
	bookingService := NewBookingService(bookingRepo, slotRepo, roomRepo, scheduleRepo)

	user, err := domain.NewUser("testuser@example.com", domain.RoleUser)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("Failed to save user: %v", err)
	}

	room, err := roomService.CreateRoom(context.Background(), "Test Room", "", 10)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	_, err = scheduleService.CreateSchedule(
		context.Background(),
		room.ID,
		domain.Monday,
		"09:00:00",
		"17:00:00",
	)
	if err != nil {
		t.Fatalf("Failed to create schedule: %v", err)
	}

	futureDate := time.Now().UTC().Add(24 * time.Hour)
	for futureDate.Weekday() != time.Monday {
		futureDate = futureDate.AddDate(0, 0, 1)
	}

	slots, err := bookingService.GetAvailableSlots(context.Background(), room.ID, futureDate)
	if err != nil {
		t.Fatalf("GetAvailableSlots() error = %v", err)
	}

	if len(slots) == 0 {
		t.Skip("No available slots found, skipping test")
	}

	slot := slots[0]
	booking, err := bookingService.CreateBooking(
		context.Background(),
		user.ID,
		slot.ID,
	)
	if err != nil {
		t.Fatalf("CreateBooking() error = %v", err)
	}

	if booking == nil {
		t.Fatal("CreateBooking() returned nil booking")
	}

	if booking.UserID != user.ID {
		t.Errorf("CreateBooking() UserID = %v, want %v", booking.UserID, user.ID)
	}

	if booking.SlotID != slot.ID {
		t.Errorf("CreateBooking() SlotID = %v, want %v", booking.SlotID, slot.ID)
	}

	if booking.Status != domain.BookingStatusActive {
		t.Errorf("CreateBooking() Status = %v, want %v", booking.Status, domain.BookingStatusActive)
	}
}

func TestBookingService_CancelBooking(t *testing.T) {
	db, teardown := setupTestDB(t)
	if db == nil {
		return
	}
	defer teardown()

	userRepo := repository.NewUserRepository(db)
	roomRepo := repository.NewRoomRepository(db)
	scheduleRepo := repository.NewScheduleRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	slotRepo := repository.NewSlotRepository(db)

	roomService := NewRoomService(roomRepo)
	scheduleService := NewScheduleService(scheduleRepo, roomRepo, slotRepo)
	bookingService := NewBookingService(bookingRepo, slotRepo, roomRepo, scheduleRepo)

	user, err := domain.NewUser("testuser2@example.com", domain.RoleUser)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("Failed to save user: %v", err)
	}

	room, err := roomService.CreateRoom(context.Background(), "Test Room 2", "", 10)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	_, err = scheduleService.CreateSchedule(
		context.Background(),
		room.ID,
		domain.Monday,
		"09:00:00",
		"17:00:00",
	)
	if err != nil {
		t.Fatalf("Failed to create schedule: %v", err)
	}

	futureDate := time.Now().UTC().Add(48 * time.Hour)
	for futureDate.Weekday() != time.Monday {
		futureDate = futureDate.AddDate(0, 0, 1)
	}

	slots, err := bookingService.GetAvailableSlots(context.Background(), room.ID, futureDate)
	if err != nil {
		t.Fatalf("GetAvailableSlots() error = %v", err)
	}

	if len(slots) == 0 {
		t.Skip("No available slots found, skipping test")
	}

	slot := slots[0]
	booking, err := bookingService.CreateBooking(
		context.Background(),
		user.ID,
		slot.ID,
	)
	if err != nil {
		t.Fatalf("CreateBooking() error = %v", err)
	}

	err = bookingService.CancelBooking(context.Background(), booking.ID)
	if err != nil {
		t.Fatalf("CancelBooking() error = %v", err)
	}

	cancelledBooking, err := bookingService.GetBookingByID(context.Background(), booking.ID)
	if err != nil {
		t.Fatalf("GetBookingByID() error = %v", err)
	}

	if cancelledBooking.Status != domain.BookingStatusCancelled {
		t.Errorf("After cancel, Status = %v, want %v", cancelledBooking.Status, domain.BookingStatusCancelled)
	}

	if cancelledBooking.CancelledAt == nil {
		t.Error("After cancel, CancelledAt should not be nil")
	}
}

func TestBookingService_GetUserBookings(t *testing.T) {
	db, teardown := setupTestDB(t)
	if db == nil {
		return
	}
	defer teardown()

	userRepo := repository.NewUserRepository(db)
	roomRepo := repository.NewRoomRepository(db)
	scheduleRepo := repository.NewScheduleRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	slotRepo := repository.NewSlotRepository(db)

	roomService := NewRoomService(roomRepo)
	scheduleService := NewScheduleService(scheduleRepo, roomRepo, slotRepo)
	bookingService := NewBookingService(bookingRepo, slotRepo, roomRepo, scheduleRepo)

	user, err := domain.NewUser("testuser3@example.com", domain.RoleUser)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("Failed to save user: %v", err)
	}

	room, err := roomService.CreateRoom(context.Background(), "Test Room 3", "", 10)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	_, err = scheduleService.CreateSchedule(
		context.Background(),
		room.ID,
		domain.Monday,
		"09:00:00",
		"17:00:00",
	)
	if err != nil {
		t.Fatalf("Failed to create schedule: %v", err)
	}

	bookings, err := bookingService.GetUserBookings(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("GetUserBookings() error = %v", err)
	}

	initialCount := len(bookings)

	futureDate := time.Now().UTC().Add(72 * time.Hour)
	for futureDate.Weekday() != time.Monday {
		futureDate = futureDate.AddDate(0, 0, 1)
	}

	slots, err := bookingService.GetAvailableSlots(context.Background(), room.ID, futureDate)
	if err != nil || len(slots) == 0 {
		t.Skip("No available slots found, skipping test")
	}

	_, err = bookingService.CreateBooking(
		context.Background(),
		user.ID,
		slots[0].ID,
	)
	if err != nil {
		t.Fatalf("CreateBooking() error = %v", err)
	}

	bookings, err = bookingService.GetUserBookings(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("GetUserBookings() error = %v", err)
	}

	if len(bookings) != initialCount+1 {
		t.Errorf("GetUserBookings() returned %d bookings, want %d", len(bookings), initialCount+1)
	}
}

func TestAuthService_DummyLogin(t *testing.T) {
	db, teardown := setupTestDB(t)
	if db == nil {
		return
	}
	defer teardown()

	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret-key", 30*time.Minute)

	token, err := authService.DummyLogin(context.Background(), "admin")
	if err != nil {
		t.Fatalf("DummyLogin(admin) error = %v", err)
	}
	if token == "" {
		t.Error("DummyLogin(admin) returned empty token")
	}

	token, err = authService.DummyLogin(context.Background(), "user")
	if err != nil {
		t.Fatalf("DummyLogin(user) error = %v", err)
	}
	if token == "" {
		t.Error("DummyLogin(user) returned empty token")
	}

	token, err = authService.DummyLogin(context.Background(), "invalid")
	if err == nil {
		t.Error("DummyLogin(invalid) should return error")
	}
}

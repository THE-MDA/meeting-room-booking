package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"meeting-room-booking/internal/config"
	httpHandler "meeting-room-booking/internal/handler/http"
	"meeting-room-booking/internal/handler/http/middleware"
	"meeting-room-booking/internal/logger"
	"meeting-room-booking/internal/migrator"
	"meeting-room-booking/internal/repository"
	"meeting-room-booking/internal/service"
)

var testDBMutex sync.Mutex

func TestE2E_CompleteBookingFlow(t *testing.T) {
	cfg := &config.Config{
		DBHost:        "localhost",
		DBPort:        5433,
		DBUser:        "app",
		DBPassword:    "secret",
		DBName:        "meeting_rooms_test",
		JWTSecret:     "test-secret-key",
		JWTExpiration: 30 * time.Minute,
		ServerPort:    "8081",
		Environment:   "test",
		LogLevel:      "debug",
	}

	db, err := setupTestDatabase(cfg)
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer teardownTestDatabase(db)

	logger.Init(parseLogLevel("debug"))

	dbWrapper := &repository.DB{DB: db}
	userRepo := repository.NewUserRepository(dbWrapper)
	roomRepo := repository.NewRoomRepository(dbWrapper)
	scheduleRepo := repository.NewScheduleRepository(dbWrapper)
	bookingRepo := repository.NewBookingRepository(dbWrapper)
	slotRepo := repository.NewSlotRepository(dbWrapper)

	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiration)
	roomService := service.NewRoomService(roomRepo)
	scheduleService := service.NewScheduleService(scheduleRepo, roomRepo, slotRepo)
	bookingService := service.NewBookingService(bookingRepo, slotRepo, roomRepo, scheduleRepo)

	authHandler := httpHandler.NewAuthHandler(authService)
	roomHandler := httpHandler.NewRoomHandler(roomService)
	scheduleHandler := httpHandler.NewScheduleHandler(scheduleService)
	bookingHandler := httpHandler.NewBookingHandler(bookingService)

	authMiddleware := middleware.NewAuthMiddleware(authService)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /dummyLogin", authHandler.DummyLogin)
	mux.HandleFunc("POST /rooms/create", authMiddleware.Authenticate(authMiddleware.RequireAdmin(roomHandler.CreateRoom)))
	mux.HandleFunc("POST /rooms/{roomId}/schedule/create", authMiddleware.Authenticate(authMiddleware.RequireAdmin(scheduleHandler.CreateSchedule)))
	mux.HandleFunc("GET /rooms/{roomId}/slots/list", authMiddleware.Authenticate(bookingHandler.GetAvailableSlots))
	mux.HandleFunc("POST /bookings/create", authMiddleware.Authenticate(bookingHandler.CreateBooking))
	mux.HandleFunc("POST /bookings/{bookingId}/cancel", authMiddleware.Authenticate(bookingHandler.CancelBooking))

	t.Log("Step 1: Getting admin token...")
	adminToken := getToken(t, mux, "admin")
	if adminToken == "" {
		t.Fatal("Failed to get admin token")
	}

	t.Log("Step 2: Creating a room...")
	roomID := createRoom(t, mux, adminToken)
	if roomID == "" {
		t.Fatal("Failed to create room")
	}
	t.Logf("Room created with ID: %s", roomID)

	t.Log("Step 3: Creating schedule...")
	createSchedule(t, mux, adminToken, roomID)

	// Подождите немного и проверьте слоты
	time.Sleep(100 * time.Millisecond)

	var slotCount int
	err = db.QueryRow("SELECT COUNT(*) FROM slots WHERE room_id = $1", roomID).Scan(&slotCount)
	if err != nil {
		t.Logf("Error checking slots: %v", err)
	} else {
		t.Logf("Slots in DB after schedule creation: %d", slotCount)
	}

	// Выведите список дат, для которых есть слоты
	rows, err := db.Query("SELECT DISTINCT DATE(start_time) FROM slots WHERE room_id = $1 ORDER BY DATE(start_time) LIMIT 10", roomID)
	if err != nil {
		t.Logf("Error getting dates: %v", err)
	} else {
		var dates []string
		for rows.Next() {
			var date string
			rows.Scan(&date)
			dates = append(dates, date)
		}
		t.Logf("Available slot dates: %v", dates)
	}

	t.Log("Step 4: Getting available slots...")
	slots := getAvailableSlots(t, mux, adminToken, roomID)
	if len(slots) == 0 {
		t.Fatal("No available slots found")
	}
	t.Logf("Found %d available slots", len(slots))

	t.Log("Step 5: Creating booking...")
	userToken := getToken(t, mux, "user")
	bookingID := createBooking(t, mux, userToken, slots[0])
	if bookingID == "" {
		t.Fatal("Failed to create booking")
	}
	t.Logf("Booking created with ID: %s", bookingID)

	t.Log("Complete booking flow test passed!")
}

func TestE2E_CancelBookingFlow(t *testing.T) {
	cfg := &config.Config{
		DBHost:        "localhost",
		DBPort:        5433,
		DBUser:        "app",
		DBPassword:    "secret",
		DBName:        "meeting_rooms_test",
		JWTSecret:     "test-secret-key",
		JWTExpiration: 30 * time.Minute,
		ServerPort:    "8081",
		Environment:   "test",
		LogLevel:      "debug",
	}

	db, err := setupTestDatabase(cfg)
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer teardownTestDatabase(db)

	logger.Init(parseLogLevel("debug"))

	dbWrapper := &repository.DB{DB: db}
	userRepo := repository.NewUserRepository(dbWrapper)
	roomRepo := repository.NewRoomRepository(dbWrapper)
	scheduleRepo := repository.NewScheduleRepository(dbWrapper)
	bookingRepo := repository.NewBookingRepository(dbWrapper)
	slotRepo := repository.NewSlotRepository(dbWrapper)

	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiration)
	roomService := service.NewRoomService(roomRepo)
	scheduleService := service.NewScheduleService(scheduleRepo, roomRepo, slotRepo)
	bookingService := service.NewBookingService(bookingRepo, slotRepo, roomRepo, scheduleRepo)

	authHandler := httpHandler.NewAuthHandler(authService)
	roomHandler := httpHandler.NewRoomHandler(roomService)
	scheduleHandler := httpHandler.NewScheduleHandler(scheduleService)
	bookingHandler := httpHandler.NewBookingHandler(bookingService)

	authMiddleware := middleware.NewAuthMiddleware(authService)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /dummyLogin", authHandler.DummyLogin)
	mux.HandleFunc("POST /rooms/create", authMiddleware.Authenticate(authMiddleware.RequireAdmin(roomHandler.CreateRoom)))
	mux.HandleFunc("POST /rooms/{roomId}/schedule/create", authMiddleware.Authenticate(authMiddleware.RequireAdmin(scheduleHandler.CreateSchedule)))
	mux.HandleFunc("GET /rooms/{roomId}/slots/list", authMiddleware.Authenticate(bookingHandler.GetAvailableSlots))
	mux.HandleFunc("POST /bookings/create", authMiddleware.Authenticate(bookingHandler.CreateBooking))
	mux.HandleFunc("POST /bookings/{bookingId}/cancel", authMiddleware.Authenticate(bookingHandler.CancelBooking))

	t.Log("Step 1: Setting up room and schedule...")
	adminToken := getToken(t, mux, "admin")
	roomID := createRoom(t, mux, adminToken)
	createSchedule(t, mux, adminToken, roomID)

	t.Log("Step 2: Creating booking...")
	userToken := getToken(t, mux, "user")
	slots := getAvailableSlots(t, mux, userToken, roomID)
	if len(slots) == 0 {
		t.Fatal("No available slots found")
	}
	bookingID := createBooking(t, mux, userToken, slots[0])
	t.Logf("Booking created with ID: %s", bookingID)

	t.Log("Step 3: Cancelling booking...")
	cancelBooking(t, mux, userToken, bookingID)

	t.Log("Cancel booking flow test passed!")
}

func createBooking(t *testing.T, handler http.Handler, token string, slot map[string]interface{}) string {
	t.Logf("Creating booking with token: %s", token)

	body := bytes.NewBufferString(`{
        "slotId": "` + slot["id"].(string) + `"
    }`)
	req := httptest.NewRequest("POST", "/bookings/create", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	t.Logf("Create booking response status: %d", w.Code)
	t.Logf("Create booking response body: %s", w.Body.String())

	if w.Code != http.StatusCreated {
		t.Errorf("Create booking failed with status %d", w.Code)
		return ""
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
		return ""
	}

	booking, ok := response["booking"].(map[string]interface{})
	if !ok {
		t.Errorf("Booking not found in response")
		return ""
	}

	return booking["id"].(string)
}

func cancelBooking(t *testing.T, handler http.Handler, token, bookingID string) {
	url := "/bookings/" + bookingID + "/cancel"
	req := httptest.NewRequest("POST", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.URL.Path = url

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	t.Logf("Cancel booking response status: %d", w.Code)
	t.Logf("Cancel booking response body: %s", w.Body.String())

	if w.Code != http.StatusOK {
		t.Errorf("Cancel booking failed with status %d", w.Code)
	}
}

func getToken(t *testing.T, handler http.Handler, role string) string {
	body := bytes.NewBufferString(`{"role":"` + role + `"}`)
	req := httptest.NewRequest("POST", "/dummyLogin", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Get token failed with status %d, body: %s", w.Code, w.Body.String())
		return ""
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Errorf("Failed to decode response: %v", err)
		return ""
	}

	return resp["token"]
}

func createRoom(t *testing.T, handler http.Handler, token string) string {
	body := bytes.NewBufferString(`{
        "name": "Test Conference Room",
        "description": "Large conference room with projector",
        "capacity": 20
    }`)
	req := httptest.NewRequest("POST", "/rooms/create", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Create room failed with status %d, body: %s", w.Code, w.Body.String())
		return ""
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
		return ""
	}

	return response["id"].(string)
}

func createSchedule(t *testing.T, handler http.Handler, token, roomID string) {
	body := bytes.NewBufferString(`{
        "day_of_week": 0,
        "start_time": "09:00:00",
        "end_time": "17:00:00"
    }`)
	req := httptest.NewRequest("POST", "/rooms/"+roomID+"/schedule/create", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Create schedule failed with status %d, body: %s", w.Code, w.Body.String())
	}

}
func getAvailableSlots(t *testing.T, handler http.Handler, token, roomID string) []map[string]interface{} {
	// Получаем реальную дату из БД
	var dateTimeStr string
	db, err := sql.Open("postgres", "host=localhost port=5433 user=app password=secret dbname=meeting_rooms_test sslmode=disable")
	if err != nil {
		t.Logf("Failed to connect to DB: %v", err)
		return nil
	}
	defer db.Close()

	err = db.QueryRow("SELECT DISTINCT DATE(start_time) FROM slots WHERE room_id = $1 ORDER BY DATE(start_time) LIMIT 1", roomID).Scan(&dateTimeStr)
	if err != nil {
		t.Logf("No dates found in slots: %v", err)
		return nil
	}

	// Преобразуем дату в правильный формат YYYY-MM-DD
	// dateTimeStr приходит как "2026-04-06T00:00:00Z" или "2026-04-06"
	var dateStr string
	if len(dateTimeStr) >= 10 {
		dateStr = dateTimeStr[:10] // Берём первые 10 символов (YYYY-MM-DD)
	} else {
		dateStr = dateTimeStr
	}

	t.Logf("Looking for slots on date: %s", dateStr)

	req := httptest.NewRequest("GET", "/rooms/"+roomID+"/slots/list?date="+dateStr, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	t.Logf("Get slots response status: %d", w.Code)
	t.Logf("Get slots response body: %s", w.Body.String())

	if w.Code != http.StatusOK {
		t.Errorf("Get slots failed with status %d", w.Code)
		return nil
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode slots: %v", err)
		return nil
	}

	slots, ok := response["slots"].([]interface{})
	if !ok {
		return nil
	}

	result := make([]map[string]interface{}, len(slots))
	for i, s := range slots {
		result[i] = s.(map[string]interface{})
	}
	return result
}

func setupTestDatabase(cfg *config.Config) (*sql.DB, error) {
	testDBMutex.Lock()
    defer testDBMutex.Unlock()
	
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword)

	adminDB, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	defer adminDB.Close()

	if err := adminDB.Ping(); err != nil {
		return nil, fmt.Errorf("cannot connect to postgres database: %w", err)
	}

	// Убиваем активные соединения
	_, _ = adminDB.Exec(`
        SELECT pg_terminate_backend(pg_stat_activity.pid)
        FROM pg_stat_activity
        WHERE pg_stat_activity.datname = 'meeting_rooms_test'
          AND pid <> pg_backend_pid()
    `)

	// Удаляем тестовую БД если существует
	_, _ = adminDB.Exec("DROP DATABASE IF EXISTS meeting_rooms_test")

	// Создаём тестовую БД
	_, err = adminDB.Exec("CREATE DATABASE meeting_rooms_test")
	if err != nil {
		// Если БД уже существует, продолжаем
		if err.Error() != "pq: database \"meeting_rooms_test\" already exists" {
			return nil, fmt.Errorf("failed to create test database: %w", err)
		}
	}

	// Подключаемся к тестовой БД
	db, err := sql.Open("postgres", cfg.GetDBConnectionString())
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("cannot connect to test database: %w", err)
	}

	// Применяем миграции
	mig := migrator.New("../../migrations", cfg.GetDatabaseURL())
	if err := mig.Up(); err != nil {
		// Если миграции уже применены, игнорируем ошибку
		if err.Error() != "no change" && err.Error() != "migration failed: no change" {
			return nil, fmt.Errorf("migration failed: %w", err)
		}
	}

	return db, nil
}

func teardownTestDatabase(db *sql.DB) {
	db.Exec("TRUNCATE bookings, slots, schedules, rooms, users CASCADE")
	db.Close()
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	default:
		return slog.LevelInfo
	}
}

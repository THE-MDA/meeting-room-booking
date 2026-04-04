package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"meeting-room-booking/internal/config"
	httpHandler "meeting-room-booking/internal/handler/http"
	"meeting-room-booking/internal/handler/http/middleware"
	"meeting-room-booking/internal/logger"
	"meeting-room-booking/internal/migrator"
	"meeting-room-booking/internal/repository"
	"meeting-room-booking/internal/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found, using environment variables")
	}

	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	logLevel := parseLogLevel(cfg.LogLevel)
	logger.Init(logLevel)

	slog.Info("Starting meeting room booking service",
		"environment", cfg.Environment,
		"port", cfg.ServerPort,
	)

	db, err := sql.Open("postgres", cfg.GetDBConnectionString())
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		slog.Error("Failed to ping database", "error", err)
		os.Exit(1)
	}
	slog.Info("Successfully connected to database",
		"host", cfg.DBHost,
		"port", cfg.DBPort,
		"database", cfg.DBName,
	)

	slog.Info("Running database migrations...")
	mig := migrator.New("migrations", cfg.GetDatabaseURL())
	if err := mig.Up(); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	dbWrapper := &repository.DB{DB: db}
	userRepo := repository.NewUserRepository(dbWrapper)
	roomRepo := repository.NewRoomRepository(dbWrapper)
	scheduleRepo := repository.NewScheduleRepository(dbWrapper)
	bookingRepo := repository.NewBookingRepository(dbWrapper)
	slotRepo := repository.NewSlotRepository(dbWrapper)

	slog.Info("Repositories initialized")

	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiration)
	roomService := service.NewRoomService(roomRepo)
	scheduleService := service.NewScheduleService(scheduleRepo, roomRepo, slotRepo)
	bookingService := service.NewBookingService(bookingRepo, slotRepo, roomRepo, scheduleRepo)

	slog.Info("Services initialized")

	authHandler := httpHandler.NewAuthHandler(authService)
	roomHandler := httpHandler.NewRoomHandler(roomService)
	scheduleHandler := httpHandler.NewScheduleHandler(scheduleService)
	bookingHandler := httpHandler.NewBookingHandler(bookingService)
	adminHandler := httpHandler.NewAdminHandler(bookingService)

	authMiddleware := middleware.NewAuthMiddleware(authService)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /dummyLogin", authHandler.DummyLogin)

	mux.HandleFunc("GET /rooms/list", authMiddleware.Authenticate(roomHandler.GetAllRooms))
	mux.HandleFunc("POST /rooms/create", authMiddleware.Authenticate(authMiddleware.RequireAdmin(roomHandler.CreateRoom)))

	mux.HandleFunc("POST /rooms/{roomId}/schedule/create", authMiddleware.Authenticate(authMiddleware.RequireAdmin(scheduleHandler.CreateSchedule)))

	mux.HandleFunc("GET /rooms/{roomId}/slots/list", authMiddleware.Authenticate(bookingHandler.GetAvailableSlots))

	mux.HandleFunc("POST /bookings/create", authMiddleware.Authenticate(bookingHandler.CreateBooking))
	mux.HandleFunc("GET /bookings/list", authMiddleware.Authenticate(authMiddleware.RequireAdmin(adminHandler.GetAllBookings)))
	mux.HandleFunc("GET /bookings/my", authMiddleware.Authenticate(bookingHandler.GetMyBookings))
	mux.HandleFunc("POST /bookings/{bookingId}/cancel", authMiddleware.Authenticate(bookingHandler.CancelBooking))

	mux.HandleFunc("GET /_info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","timestamp":"` + time.Now().UTC().Format(time.RFC3339) + `"}`))
	})

	handler := loggingMiddleware(mux)

	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("Server starting", "port", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server stopped")
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		slog.Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.statusCode,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

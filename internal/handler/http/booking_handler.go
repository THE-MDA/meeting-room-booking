package http

import (
    "encoding/json"
    "net/http"
    "time"

    "github.com/google/uuid"
    "meeting-room-booking/internal/handler/http/middleware"
    "meeting-room-booking/internal/service"
)

type BookingHandler struct {
    bookingService *service.BookingService
}

func NewBookingHandler(bookingService *service.BookingService) *BookingHandler {
    return &BookingHandler{
        bookingService: bookingService,
    }
}

type getSlotsRequest struct {
    RoomID string `json:"room_id"`
    Date   string `json:"date"`
}

type createBookingRequest struct {
    SlotID    string `json:"slot_id"`
    RoomID    string `json:"room_id"`
    StartTime string `json:"start_time"`
    EndTime   string `json:"end_time"`
}

func (h *BookingHandler) GetAvailableSlots(w http.ResponseWriter, r *http.Request) {
    roomIDStr := r.URL.Query().Get("room_id")
    dateStr := r.URL.Query().Get("date")

    if roomIDStr == "" || dateStr == "" {
        http.Error(w, "room_id and date are required", http.StatusBadRequest)
        return
    }

    roomID, err := uuid.Parse(roomIDStr)
    if err != nil {
        http.Error(w, "invalid room_id", http.StatusBadRequest)
        return
    }

    date, err := time.Parse("2006-01-02", dateStr)
    if err != nil {
        http.Error(w, "invalid date format, expected YYYY-MM-DD", http.StatusBadRequest)
        return
    }

    slots, err := h.bookingService.GetAvailableSlots(r.Context(), roomID, date)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(slots)
}

func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
    userIDStr, ok := middleware.GetUserID(r.Context())
    if !ok {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    userID, err := uuid.Parse(userIDStr)
    if err != nil {
        http.Error(w, "invalid user id", http.StatusUnauthorized)
        return
    }

    var req createBookingRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }

    roomID, err := uuid.Parse(req.RoomID)
    if err != nil {
        http.Error(w, "invalid room_id", http.StatusBadRequest)
        return
    }

    startTime, err := time.Parse(time.RFC3339, req.StartTime)
    if err != nil {
        http.Error(w, "invalid start_time format, expected RFC3339", http.StatusBadRequest)
        return
    }

    endTime, err := time.Parse(time.RFC3339, req.EndTime)
    if err != nil {
        http.Error(w, "invalid end_time format, expected RFC3339", http.StatusBadRequest)
        return
    }

    booking, err := h.bookingService.CreateBooking(r.Context(), userID, req.SlotID, roomID, startTime, endTime)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(booking)
}

func (h *BookingHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
    bookingIDStr := r.URL.Path[len("/bookings/"):]
    
    bookingID, err := uuid.Parse(bookingIDStr)
    if err != nil {
        http.Error(w, "invalid booking id", http.StatusBadRequest)
        return
    }

    userIDStr, ok := middleware.GetUserID(r.Context())
    if !ok {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    userID, err := uuid.Parse(userIDStr)
    if err != nil {
        http.Error(w, "invalid user id", http.StatusUnauthorized)
        return
    }

    booking, err := h.bookingService.GetBookingByID(r.Context(), bookingID)
    if err != nil {
        http.Error(w, "booking not found", http.StatusNotFound)
        return
    }

    if booking.UserID != userID {
        http.Error(w, "access denied: booking belongs to another user", http.StatusForbidden)
        return
    }

    if err := h.bookingService.CancelBooking(r.Context(), bookingID); err != nil {
        http.Error(w, "failed to cancel booking", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "cancelled"})
}

func (h *BookingHandler) GetMyBookings(w http.ResponseWriter, r *http.Request) {
    userIDStr, ok := middleware.GetUserID(r.Context())
    if !ok {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    userID, err := uuid.Parse(userIDStr)
    if err != nil {
        http.Error(w, "invalid user id", http.StatusUnauthorized)
        return
    }

    bookings, err := h.bookingService.GetUserBookings(r.Context(), userID)
    if err != nil {
        http.Error(w, "failed to get bookings", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(bookings)
}
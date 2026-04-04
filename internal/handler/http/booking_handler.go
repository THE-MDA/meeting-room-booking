package http

import (
	"encoding/json"
	"net/http"
	"time"

	"meeting-room-booking/internal/handler/http/middleware"
	"meeting-room-booking/internal/service"

	"github.com/google/uuid"
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
    SlotID              string `json:"slotId"`
}

func (h *BookingHandler) GetAvailableSlots(w http.ResponseWriter, r *http.Request) {
    roomIDStr := r.PathValue("roomId")
    if roomIDStr == "" {
        http.Error(w, "room_id is required", http.StatusBadRequest)
        return
    }
    roomID, err := uuid.Parse(roomIDStr)
    if err != nil {
        http.Error(w, "invalid room_id", http.StatusBadRequest)
        return
    }
    dateStr := r.URL.Query().Get("date")
    if dateStr == "" {
        http.Error(w, "date is required", http.StatusBadRequest)
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
    json.NewEncoder(w).Encode(map[string]interface{}{"slots": slots})
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
    if req.SlotID == "" {
        http.Error(w, "slotId is required", http.StatusBadRequest)
        return
    }
    booking, err := h.bookingService.CreateBooking(r.Context(), userID, req.SlotID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{"booking": booking})
}

func (h *BookingHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	bookingIDStr := r.PathValue("bookingId")
	if bookingIDStr == "" {
		http.Error(w, "booking_id is required", http.StatusBadRequest)
		return
	}

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
		http.Error(w, "cannot cancel another user's booking", http.StatusForbidden)
		return
	}

	if err := h.bookingService.CancelBooking(r.Context(), bookingID); err != nil {
		http.Error(w, "failed to cancel booking", http.StatusInternalServerError)
		return
	}

	updatedBooking, _ := h.bookingService.GetBookingByID(r.Context(), bookingID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"booking": updatedBooking})
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

package http

import (
    "encoding/json"
    "net/http"
    "strconv"

    "meeting-room-booking/internal/service"
)

type AdminHandler struct {
    bookingService *service.BookingService
}

func NewAdminHandler(bookingService *service.BookingService) *AdminHandler {
    return &AdminHandler{
        bookingService: bookingService,
    }
}

func (h *AdminHandler) GetAllBookings(w http.ResponseWriter, r *http.Request) {
    limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
    offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

    if limit <= 0 {
        limit = 20
    }
    if offset < 0 {
        offset = 0
    }

    bookings, err := h.bookingService.GetAllBookings(r.Context(), limit, offset)
    if err != nil {
        http.Error(w, "failed to get bookings", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(bookings)
}
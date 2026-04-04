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
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	bookings, err := h.bookingService.GetAllBookings(r.Context(), pageSize, offset)
	if err != nil {
		http.Error(w, "failed to get bookings", http.StatusInternalServerError)
		return
	}

	total, err := h.bookingService.CountBookings(r.Context())
	if err != nil {
		total = len(bookings)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"bookings": bookings,
		"pagination": map[string]int{
			"page":     page,
			"pageSize": pageSize,
			"total":    total,
		},
	})
}

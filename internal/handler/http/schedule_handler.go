package http

import (
    "encoding/json"
    "net/http"

    "meeting-room-booking/internal/domain"
    "meeting-room-booking/internal/service"
    "github.com/google/uuid"
)

type ScheduleHandler struct {
    scheduleService *service.ScheduleService
}

func NewScheduleHandler(scheduleService *service.ScheduleService) *ScheduleHandler {
    return &ScheduleHandler{
        scheduleService: scheduleService,
    }
}

type createScheduleRequest struct {
    RoomID    string `json:"room_id"`
    DayOfWeek int    `json:"day_of_week"`
    StartTime string `json:"start_time"`
    EndTime   string `json:"end_time"`
}

func (h *ScheduleHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
    var req createScheduleRequest
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }

    roomID, err := uuid.Parse(req.RoomID)
    if err != nil {
        http.Error(w, "invalid room_id", http.StatusBadRequest)
        return
    }

    schedule, err := h.scheduleService.CreateSchedule(
        r.Context(),
        roomID,
        domain.DayOfWeek(req.DayOfWeek),
        req.StartTime,
        req.EndTime,
    )
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(schedule)
}
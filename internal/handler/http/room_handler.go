package http

import (
    "encoding/json"
    "net/http"

    "meeting-room-booking/internal/service"
)

type RoomHandler struct {
    roomService *service.RoomService
}

func NewRoomHandler(roomService *service.RoomService) *RoomHandler {
    return &RoomHandler{
        roomService: roomService,
    }
}

type createRoomRequest struct {
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
    Capacity    int    `json:"capacity"`
}

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
    var req createRoomRequest
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }

    room, err := h.roomService.CreateRoom(r.Context(), req.Name, req.Description, req.Capacity)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(room)
}

func (h *RoomHandler) GetAllRooms(w http.ResponseWriter, r *http.Request) {
    rooms, err := h.roomService.GetAllRooms(r.Context())
    if err != nil {
        http.Error(w, "failed to get rooms", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(rooms)
}
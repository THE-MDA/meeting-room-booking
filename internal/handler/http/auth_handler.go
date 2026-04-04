package http

import (
	"encoding/json"
	"net/http"

	"meeting-room-booking/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

type dummyLoginRequest struct {
	Role string `json:"role"`
}

type dummyLoginResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	var req dummyLoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Role != "admin" && req.Role != "user" {
		http.Error(w, "invalid role, must be 'admin' or 'user'", http.StatusBadRequest)
		return
	}

	token, err := h.authService.DummyLogin(r.Context(), req.Role)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dummyLoginResponse{Token: token})
}
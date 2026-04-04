package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"meeting-room-booking/internal/domain"
	"meeting-room-booking/internal/service"

	"github.com/google/uuid"
)

func setupTestAuthService() *service.AuthService {
	mockUserRepo := &mockUserRepository{
		users: make(map[uuid.UUID]*domain.User),
	}

	adminUser, _ := domain.NewUser("admin@example.com", domain.RoleAdmin)
	adminUser.ID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	mockUserRepo.users[adminUser.ID] = adminUser

	userUser, _ := domain.NewUser("user@example.com", domain.RoleUser)
	userUser.ID = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	mockUserRepo.users[userUser.ID] = userUser

	return service.NewAuthService(mockUserRepo, "test-secret", 30*time.Minute)
}

type mockUserRepository struct {
	users map[uuid.UUID]*domain.User
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, nil
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepository) GetByRole(ctx context.Context, role domain.UserRole) ([]*domain.User, error) {
	var users []*domain.User
	for _, user := range m.users {
		if user.Role == role {
			users = append(users, user)
		}
	}
	return users, nil
}

func TestAuthenticate(t *testing.T) {
	authService := setupTestAuthService()
	m := NewAuthMiddleware(authService)

	token, err := authService.GenerateToken(uuid.MustParse("00000000-0000-0000-0000-000000000001"), "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	handler := m.Authenticate(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			t.Error("GetUserID should return true")
		}
		if userID == "" {
			t.Error("UserID should not be empty")
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
}

func TestAuthenticate_NoHeader(t *testing.T) {
	authService := setupTestAuthService()
	m := NewAuthMiddleware(authService)

	handler := m.Authenticate(func(w http.ResponseWriter, r *http.Request) {})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected Unauthorized, got %d", w.Code)
	}
}

func TestAuthenticate_InvalidHeader(t *testing.T) {
	authService := setupTestAuthService()
	m := NewAuthMiddleware(authService)

	handler := m.Authenticate(func(w http.ResponseWriter, r *http.Request) {})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Invalid test-token")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected Unauthorized, got %d", w.Code)
	}
}

func TestRequireAdmin(t *testing.T) {
	authService := setupTestAuthService()
	m := NewAuthMiddleware(authService)

	handler := m.RequireAdmin(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), UserRoleKey, "admin")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected OK, got %d", w.Code)
	}
}

func TestRequireAdmin_Forbidden(t *testing.T) {
	authService := setupTestAuthService()
	m := NewAuthMiddleware(authService)

	handler := m.RequireAdmin(func(w http.ResponseWriter, r *http.Request) {})

	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), UserRoleKey, "user")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected Forbidden, got %d", w.Code)
	}
}

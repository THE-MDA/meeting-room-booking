package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		role    UserRole
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid user creation",
			email:   "user@example.com",
			role:    RoleUser,
			wantErr: false,
		},
		{
			name:    "valid admin creation",
			email:   "admin@example.com",
			role:    RoleAdmin,
			wantErr: false,
		},
		{
			name:    "empty email",
			email:   "",
			role:    RoleUser,
			wantErr: true,
			errMsg:  "email is required",
		},
		{
			name:    "invalid role",
			email:   "test@example.com",
			role:    "superuser",
			wantErr: true,
			errMsg:  "invalid role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser(tt.email, tt.role)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewUser() expected error but got nil")
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("NewUser() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewUser() unexpected error: %v", err)
			}

			if user == nil {
				t.Fatal("NewUser() returned nil user")
			}

			if user.Email != tt.email {
				t.Errorf("User.Email = %v, want %v", user.Email, tt.email)
			}

			if user.Role != tt.role {
				t.Errorf("User.Role = %v, want %v", user.Role, tt.role)
			}

			if user.ID == uuid.Nil {
				t.Error("User.ID should not be nil")
			}
		})
	}
}

func TestUserMethods(t *testing.T) {
	user, _ := NewUser("test@example.com", RoleAdmin)

	if !user.IsAdmin() {
		t.Error("IsAdmin() should return true for admin user")
	}

	if user.IsUser() {
		t.Error("IsUser() should return false for admin user")
	}

	user2, _ := NewUser("user@example.com", RoleUser)

	if user2.IsAdmin() {
		t.Error("IsAdmin() should return false for regular user")
	}

	if !user2.IsUser() {
		t.Error("IsUser() should return true for regular user")
	}
}

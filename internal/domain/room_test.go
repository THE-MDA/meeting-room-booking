package domain

import (
	"testing"
)

func TestNewRoom(t *testing.T) {
	tests := []struct {
		name        string
		roomName    string
		description string
		capacity    int
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid room",
			roomName:    "Conference Room",
			description: "Conference room",
			capacity:    10,
			wantErr:     false,
		},
		{
			name:        "empty name",
			roomName:    "",
			description: "",
			capacity:    10,
			wantErr:     true,
			errMsg:      "room name is required",
		},
		{
			name:        "zero capacity",
			roomName:    "Test Room",
			description: "Test room",
			capacity:    0,
			wantErr:     true,
			errMsg:      "capacity must be positive",
		},
		{
			name:        "negative capacity",
			roomName:    "Test Room",
			description: "Test room",
			capacity:    -5,
			wantErr:     true,
			errMsg:      "capacity must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			room, err := NewRoom(tt.roomName, tt.description, tt.capacity)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewRoom() expected error but got nil")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("NewRoom() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewRoom() unexpected error: %v", err)
				return
			}

			if room == nil {
				t.Fatal("NewRoom() returned nil room")
			}

			if room.Name != tt.roomName {
				t.Errorf("Room.Name = %v, want %v", room.Name, tt.roomName)
			}

			if room.Capacity != tt.capacity {
				t.Errorf("Room.Capacity = %v, want %v", room.Capacity, tt.capacity)
			}
		})
	}
}

func TestRoomValidate(t *testing.T) {
	room, _ := NewRoom("Test", "Desc", 5)

	if err := room.Validate(); err != nil {
		t.Errorf("Validate() should not return error for valid room: %v", err)
	}

	room.Name = ""
	if err := room.Validate(); err == nil {
		t.Error("Validate() should return error for empty name")
	}

	room.Name = "Test"
	room.Capacity = 0
	if err := room.Validate(); err == nil {
		t.Error("Validate() should return error for zero capacity")
	}
}

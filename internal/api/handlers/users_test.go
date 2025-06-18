package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"Coves/internal/api/handlers"
	"Coves/internal/core/users"
)

type mockUserService struct {
	users      []users.User
	shouldFail bool
}

func (m *mockUserService) CreateUser(req users.CreateUserRequest) (*users.User, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("service: failed to create user")
	}
	
	user := &users.User{
		ID:       1,
		Email:    req.Email,
		Username: req.Username,
	}
	m.users = append(m.users, *user)
	return user, nil
}

func (m *mockUserService) GetUserByID(id int) (*users.User, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("service: failed to get user")
	}
	
	for _, user := range m.users {
		if user.ID == id {
			return &user, nil
		}
	}
	return nil, fmt.Errorf("service: user not found")
}

func (m *mockUserService) GetUserByEmail(email string) (*users.User, error) {
	for i := range m.users {
		if m.users[i].Email == email {
			return &m.users[i], nil
		}
	}
	return nil, fmt.Errorf("service: user not found")
}

func (m *mockUserService) GetUserByUsername(username string) (*users.User, error) {
	for i := range m.users {
		if m.users[i].Username == username {
			return &m.users[i], nil
		}
	}
	return nil, fmt.Errorf("service: user not found")
}

func (m *mockUserService) UpdateUser(id int, req users.UpdateUserRequest) (*users.User, error) {
	for i := range m.users {
		if m.users[i].ID == id {
			if req.Email != "" {
				m.users[i].Email = req.Email
			}
			if req.Username != "" {
				m.users[i].Username = req.Username
			}
			return &m.users[i], nil
		}
	}
	return nil, fmt.Errorf("service: user not found")
}

func (m *mockUserService) DeleteUser(id int) error {
	for i, user := range m.users {
		if user.ID == id {
			m.users = append(m.users[:i], m.users[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("service: user not found")
}

func TestCreateUserHandler(t *testing.T) {
	service := &mockUserService{}
	handler := handlers.NewUserHandler(service)
	
	tests := []struct {
		name       string
		body       users.CreateUserRequest
		wantStatus int
	}{
		{
			name: "valid request",
			body: users.CreateUserRequest{
				Email:    "test@example.com",
				Username: "testuser",
			},
			wantStatus: http.StatusCreated,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/users", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			
			rr := httptest.NewRecorder()
			handler.CreateUser(rr, req)
			
			if status := rr.Code; status != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.wantStatus)
			}
			
			if tt.wantStatus == http.StatusCreated {
				var user users.User
				if err := json.NewDecoder(rr.Body).Decode(&user); err != nil {
					t.Errorf("failed to decode response: %v", err)
				}
				if user.Email != tt.body.Email {
					t.Errorf("expected email %s but got %s", tt.body.Email, user.Email)
				}
			}
		})
	}
}
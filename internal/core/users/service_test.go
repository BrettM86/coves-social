package users_test

import (
	"fmt"
	"testing"
	"time"

	"Coves/internal/core/users"
)

type mockUserRepository struct {
	users      map[int]*users.User
	nextID     int
	shouldFail bool
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users:  make(map[int]*users.User),
		nextID: 1,
	}
}

func (m *mockUserRepository) Create(user *users.User) (*users.User, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("mock: database error")
	}
	
	user.ID = m.nextID
	m.nextID++
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	
	m.users[user.ID] = user
	return user, nil
}

func (m *mockUserRepository) GetByID(id int) (*users.User, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("mock: database error")
	}
	
	user, exists := m.users[id]
	if !exists {
		return nil, fmt.Errorf("repository: user not found")
	}
	return user, nil
}

func (m *mockUserRepository) GetByEmail(email string) (*users.User, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("mock: database error")
	}
	
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, fmt.Errorf("repository: user not found")
}

func (m *mockUserRepository) GetByUsername(username string) (*users.User, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("mock: database error")
	}
	
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, fmt.Errorf("repository: user not found")
}

func (m *mockUserRepository) Update(user *users.User) (*users.User, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("mock: database error")
	}
	
	if _, exists := m.users[user.ID]; !exists {
		return nil, fmt.Errorf("repository: user not found")
	}
	
	user.UpdatedAt = time.Now()
	m.users[user.ID] = user
	return user, nil
}

func (m *mockUserRepository) Delete(id int) error {
	if m.shouldFail {
		return fmt.Errorf("mock: database error")
	}
	
	if _, exists := m.users[id]; !exists {
		return fmt.Errorf("repository: user not found")
	}
	
	delete(m.users, id)
	return nil
}

func TestCreateUser(t *testing.T) {
	repo := newMockUserRepository()
	service := users.NewUserService(repo)
	
	tests := []struct {
		name    string
		req     users.CreateUserRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid user",
			req: users.CreateUserRequest{
				Email:    "test@example.com",
				Username: "testuser",
			},
			wantErr: false,
		},
		{
			name: "empty email",
			req: users.CreateUserRequest{
				Email:    "",
				Username: "testuser",
			},
			wantErr: true,
			errMsg:  "email is required",
		},
		{
			name: "empty username",
			req: users.CreateUserRequest{
				Email:    "test@example.com",
				Username: "",
			},
			wantErr: true,
			errMsg:  "username is required",
		},
		{
			name: "invalid email format",
			req: users.CreateUserRequest{
				Email:    "invalidemail",
				Username: "testuser",
			},
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name: "short username",
			req: users.CreateUserRequest{
				Email:    "test@example.com",
				Username: "ab",
			},
			wantErr: true,
			errMsg:  "username must be at least 3 characters",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.CreateUser(tt.req)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errMsg != "" && err.Error() != "service: "+tt.errMsg {
					t.Errorf("expected error message '%s' but got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if user == nil {
					t.Errorf("expected user but got nil")
				}
			}
		})
	}
}

func TestCreateUserDuplicates(t *testing.T) {
	repo := newMockUserRepository()
	service := users.NewUserService(repo)
	
	req := users.CreateUserRequest{
		Email:    "test@example.com",
		Username: "testuser",
	}
	
	_, err := service.CreateUser(req)
	if err != nil {
		t.Fatalf("unexpected error creating first user: %v", err)
	}
	
	_, err = service.CreateUser(req)
	if err == nil {
		t.Errorf("expected error for duplicate email but got none")
	} else if err.Error() != "service: email already exists" {
		t.Errorf("unexpected error message: %v", err)
	}
	
	req2 := users.CreateUserRequest{
		Email:    "different@example.com",
		Username: "testuser",
	}
	
	_, err = service.CreateUser(req2)
	if err == nil {
		t.Errorf("expected error for duplicate username but got none")
	} else if err.Error() != "service: username already exists" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestGetUserByID(t *testing.T) {
	repo := newMockUserRepository()
	service := users.NewUserService(repo)
	
	createdUser, err := service.CreateUser(users.CreateUserRequest{
		Email:    "test@example.com",
		Username: "testuser",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	
	tests := []struct {
		name    string
		id      int
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid ID",
			id:      createdUser.ID,
			wantErr: false,
		},
		{
			name:    "invalid ID",
			id:      0,
			wantErr: true,
			errMsg:  "invalid user ID",
		},
		{
			name:    "non-existent ID",
			id:      999,
			wantErr: true,
			errMsg:  "user not found",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.GetUserByID(tt.id)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errMsg != "" && err.Error() != "service: "+tt.errMsg {
					t.Errorf("expected error message '%s' but got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if user == nil {
					t.Errorf("expected user but got nil")
				}
			}
		})
	}
}
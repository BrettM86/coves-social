package users

import (
	"fmt"
	"strings"
)

type UserService struct {
	userRepo UserRepository
}

func NewUserService(userRepo UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) CreateUser(req CreateUserRequest) (*User, error) {
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}
	
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Username = strings.TrimSpace(req.Username)
	
	existingUser, _ := s.userRepo.GetByEmail(req.Email)
	if existingUser != nil {
		return nil, fmt.Errorf("service: email already exists")
	}
	
	existingUser, _ = s.userRepo.GetByUsername(req.Username)
	if existingUser != nil {
		return nil, fmt.Errorf("service: username already exists")
	}
	
	user := &User{
		Email:    req.Email,
		Username: req.Username,
	}
	
	return s.userRepo.Create(user)
}

func (s *UserService) GetUserByID(id int) (*User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("service: invalid user ID")
	}
	
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("service: user not found")
		}
		return nil, fmt.Errorf("service: %w", err)
	}
	
	return user, nil
}

func (s *UserService) GetUserByEmail(email string) (*User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return nil, fmt.Errorf("service: email is required")
	}
	
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("service: user not found")
		}
		return nil, fmt.Errorf("service: %w", err)
	}
	
	return user, nil
}

func (s *UserService) GetUserByUsername(username string) (*User, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, fmt.Errorf("service: username is required")
	}
	
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("service: user not found")
		}
		return nil, fmt.Errorf("service: %w", err)
	}
	
	return user, nil
}

func (s *UserService) UpdateUser(id int, req UpdateUserRequest) (*User, error) {
	user, err := s.GetUserByID(id)
	if err != nil {
		return nil, err
	}
	
	if req.Email != "" {
		req.Email = strings.TrimSpace(strings.ToLower(req.Email))
		if req.Email != user.Email {
			existingUser, _ := s.userRepo.GetByEmail(req.Email)
			if existingUser != nil && existingUser.ID != id {
				return nil, fmt.Errorf("service: email already exists")
			}
		}
		user.Email = req.Email
	}
	
	if req.Username != "" {
		req.Username = strings.TrimSpace(req.Username)
		if req.Username != user.Username {
			existingUser, _ := s.userRepo.GetByUsername(req.Username)
			if existingUser != nil && existingUser.ID != id {
				return nil, fmt.Errorf("service: username already exists")
			}
		}
		user.Username = req.Username
	}
	
	return s.userRepo.Update(user)
}

func (s *UserService) DeleteUser(id int) error {
	if id <= 0 {
		return fmt.Errorf("service: invalid user ID")
	}
	
	err := s.userRepo.Delete(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("service: user not found")
		}
		return fmt.Errorf("service: %w", err)
	}
	
	return nil
}

func (s *UserService) validateCreateRequest(req CreateUserRequest) error {
	if strings.TrimSpace(req.Email) == "" {
		return fmt.Errorf("service: email is required")
	}
	
	if strings.TrimSpace(req.Username) == "" {
		return fmt.Errorf("service: username is required")
	}
	
	if !strings.Contains(req.Email, "@") {
		return fmt.Errorf("service: invalid email format")
	}
	
	if len(req.Username) < 3 {
		return fmt.Errorf("service: username must be at least 3 characters")
	}
	
	return nil
}

type UpdateUserRequest struct {
	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
}
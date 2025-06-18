package users

type UserServiceInterface interface {
	CreateUser(req CreateUserRequest) (*User, error)
	GetUserByID(id int) (*User, error)
	GetUserByEmail(email string) (*User, error)
	GetUserByUsername(username string) (*User, error)
	UpdateUser(id int, req UpdateUserRequest) (*User, error)
	DeleteUser(id int) error
}
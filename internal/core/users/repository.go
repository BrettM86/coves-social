package users

type UserRepository interface {
	Create(user *User) (*User, error)
	GetByID(id int) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByUsername(username string) (*User, error)
	Update(user *User) (*User, error)
	Delete(id int) error
}
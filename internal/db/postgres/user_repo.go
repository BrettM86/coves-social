package postgres

import (
	"database/sql"
	"fmt"

	"Coves/internal/core/users"
)

type PostgresUserRepo struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) users.UserRepository {
	return &PostgresUserRepo{db: db}
}

func (r *PostgresUserRepo) Create(user *users.User) (*users.User, error) {
	query := `
		INSERT INTO users (email, username) 
		VALUES ($1, $2) 
		RETURNING id, email, username, created_at, updated_at`

	err := r.db.QueryRow(query, user.Email, user.Username).
		Scan(&user.ID, &user.Email, &user.Username, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("repository: failed to create user: %w", err)
	}

	return user, nil
}

func (r *PostgresUserRepo) GetByID(id int) (*users.User, error) {
	user := &users.User{}
	query := `SELECT id, email, username, created_at, updated_at FROM users WHERE id = $1`

	err := r.db.QueryRow(query, id).
		Scan(&user.ID, &user.Email, &user.Username, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("repository: user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("repository: failed to get user: %w", err)
	}

	return user, nil
}

func (r *PostgresUserRepo) GetByEmail(email string) (*users.User, error) {
	user := &users.User{}
	query := `SELECT id, email, username, created_at, updated_at FROM users WHERE email = $1`

	err := r.db.QueryRow(query, email).
		Scan(&user.ID, &user.Email, &user.Username, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("repository: user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("repository: failed to get user by email: %w", err)
	}

	return user, nil
}

func (r *PostgresUserRepo) GetByUsername(username string) (*users.User, error) {
	user := &users.User{}
	query := `SELECT id, email, username, created_at, updated_at FROM users WHERE username = $1`

	err := r.db.QueryRow(query, username).
		Scan(&user.ID, &user.Email, &user.Username, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("repository: user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("repository: failed to get user by username: %w", err)
	}

	return user, nil
}

func (r *PostgresUserRepo) Update(user *users.User) (*users.User, error) {
	query := `
		UPDATE users 
		SET email = $2, username = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, email, username, created_at, updated_at`

	err := r.db.QueryRow(query, user.ID, user.Email, user.Username).
		Scan(&user.ID, &user.Email, &user.Username, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("repository: user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("repository: failed to update user: %w", err)
	}

	return user, nil
}

func (r *PostgresUserRepo) Delete(id int) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("repository: failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("repository: user not found")
	}

	return nil
}

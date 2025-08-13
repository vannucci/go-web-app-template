package services

import (
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	UserTier  string    `json:"user_tier"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserService struct {
	db *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{db: db}
}

// Authenticate user with email/password
func (u *UserService) Authenticate(email, password string) (*User, error) {
	query := `
		SELECT id, email, name, user_tier, password_hash, created_at, updated_at 
		FROM users 
		WHERE email = $1
	`

	var user User
	var passwordHash string

	row := u.db.QueryRow(query, email)
	err := row.Scan(&user.ID, &user.Email, &user.Name, &user.UserTier, &passwordHash, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return &user, nil
}

// Get user by ID
func (u *UserService) GetByID(id int) (*User, error) {
	query := `
		SELECT id, email, name, user_tier, created_at, updated_at 
		FROM users 
		WHERE id = $1
	`

	var user User
	row := u.db.QueryRow(query, id)
	err := row.Scan(&user.ID, &user.Email, &user.Name, &user.UserTier, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &user, nil
}

// Get user by email
func (u *UserService) GetByEmail(email string) (*User, error) {
	query := `
		SELECT id, email, name, user_tier, created_at, updated_at 
		FROM users 
		WHERE email = $1
	`

	var user User
	row := u.db.QueryRow(query, email)
	err := row.Scan(&user.ID, &user.Email, &user.Name, &user.UserTier, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &user, nil
}

// Create new user
func (u *UserService) Create(email, password, name, userTier string) (*User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	query := `
		INSERT INTO users (email, password_hash, name, user_tier, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, email, name, user_tier, created_at, updated_at
	`

	now := time.Now()
	var user User

	row := u.db.QueryRow(query, email, hashedPassword, name, userTier, now, now)
	err = row.Scan(&user.ID, &user.Email, &user.Name, &user.UserTier, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// Update user
func (u *UserService) Update(id int, name, userTier string) (*User, error) {
	query := `
		UPDATE users 
		SET name = $2, user_tier = $3, updated_at = $4
		WHERE id = $1
		RETURNING id, email, name, user_tier, created_at, updated_at
	`

	var user User
	row := u.db.QueryRow(query, id, name, userTier, time.Now())
	err := row.Scan(&user.ID, &user.Email, &user.Name, &user.UserTier, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &user, nil
}

// List all users (admin function)
func (u *UserService) List() ([]*User, error) {
	query := `
		SELECT id, email, name, user_tier, created_at, updated_at 
		FROM users 
		ORDER BY created_at DESC
	`

	rows, err := u.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.UserTier, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

// Change password
func (u *UserService) ChangePassword(id int, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	query := `UPDATE users SET password_hash = $2, updated_at = $3 WHERE id = $1`

	result, err := u.db.Exec(query, id, hashedPassword, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// Delete user
func (u *UserService) Delete(id int) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := u.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID                string     `db:"string"`
	Email             string     `db:"email"`
	Name              string     `db:"name"`
	PasswordHash      string     `db:"password_hash"`
	CompanyID         string     `db:"company_id"`
	Role              string     `db:"role"`
	IsActive          bool       `db:"is_active"`
	PasswordChangedAt time.Time  `db:"password_changed_at"`
	CreatedAt         time.Time  `db:"created_at"`
	LastLogin         *time.Time `db:"last_login"`

	// Joined fields
	Company   *Company   `db:"-"`
	Workspace *Workspace `db:"-"`
}

func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	u.PasswordChangedAt = time.Now()
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

func (u *User) CanAccessCompany(companyID string) bool {
	switch u.Role {
	case "super_admin":
		return true
	case "workspace_admin":
		// Would need to check if company is in user's workspace
		return true // Simplified for now
	case "company_admin", "user":
		return u.CompanyID == companyID
	}
	return false
}

func (u *User) CanManageUsers() bool {
	return u.Role == "super_admin" || u.Role == "workspace_admin" || u.Role == "company_admin"
}

func (u *User) CanManageDestinations() bool {
	return u.Role != "user"
}

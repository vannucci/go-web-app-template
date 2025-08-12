package destinations

import (
	"time"
)

type Destination interface {
	Validate() error
	GetType() string
	GetDisplayName() string
	ToJSON() ([]byte, error)
	FromJSON([]byte) error
}

type BaseDestination struct {
	ID        int       `json:"id" db:"id"`
	CompanyID int       `json:"company_id" db:"company_id"`
	Name      string    `json:"name" db:"name"`
	Type      string    `json:"type" db:"type"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedBy int       `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

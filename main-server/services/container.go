package services

import (
	"database/sql"
	"main-server/config"
)

type Container struct {
	DB     *sql.DB
	Config *config.Config
	// AuditService *AuditService
	// TimeService  *TimeService
	AuthService *AuthService
}

func NewContainer(db *sql.DB, cfg *config.Config, authService *AuthService) *Container {
	return &Container{
		DB:     db,
		Config: cfg,
		// AuditService: &AuditService{db: db}, // You'll need to create this
		// TimeService:  &TimeService{},        // You'll need to create this
		AuthService: authService,
	}
}

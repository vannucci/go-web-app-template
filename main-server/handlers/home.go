// handlers/home.go
package handlers

import (
	"database/sql"
	"net/http"

	"main-server/config"

	"github.com/labstack/echo/v4"
)

type HomeHandler struct {
	config *config.Config
	db     *sql.DB
}

func NewHomeHandler(config *config.Config, db *sql.DB) *HomeHandler {
	return &HomeHandler{
		config: config,
		db:     db,
	}
}

func (h *HomeHandler) Home(c echo.Context) error {
	// Log audit event for page access
	// h.services.AuditService.LogEvent("page_access", "user123", "/", "home_page_view")

	// Data to pass to the template
	data := map[string]interface{}{
		"Title":       "Home",
		"Environment": h.config.Environment,
		"Version":     h.config.Version,
		"Links": []map[string]string{
			{"URL": "/health", "Text": "Health Check"},
			{"URL": "/metrics", "Text": "Prometheus Metrics"},
			{"URL": "/audit", "Text": "Generate Audit Event"},
			{"URL": "/destinations", "Text": "Destinations"},
		},
	}

	return c.Render(http.StatusOK, "home.html", data)
}

func (h *HomeHandler) Health(c echo.Context) error {
	// Test database connection
	if err := h.db.Ping(); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status": "unhealthy",
			"error":  "database connection failed",
			// "timestamp": h.TimeService.Now().Unix(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":      "healthy",
		"environment": h.config.Environment,
		"version":     h.config.Version,
		// "timestamp":   h.services.TimeService.Now().Unix(),
	})
}

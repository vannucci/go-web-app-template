// handlers/home.go
package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"main-server/services"
)

type HomeHandler struct {
	services *services.Container
}

func NewHomeHandler(services *services.Container) *HomeHandler {
	return &HomeHandler{
		services: services,
	}
}

func (h *HomeHandler) Home(c echo.Context) error {
	// Log audit event for page access
	h.services.AuditService.LogEvent("page_access", "user123", "/", "home_page_view")
	
	// Data to pass to the template
	data := map[string]interface{}{
		"Title":       "Home",
		"Environment": h.services.Config.Environment,
		"Version":     h.services.Config.Version,
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
	if err := h.services.DB.Ping(); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status":    "unhealthy",
			"error":     "database connection failed",
			"timestamp": h.services.TimeService.Now().Unix(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":      "healthy",
		"environment": h.services.Config.Environment,
		"version":     h.services.Config.Version,
		"timestamp":   h.services.TimeService.Now().Unix(),
	})
}
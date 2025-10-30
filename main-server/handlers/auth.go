package handlers

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"

	"main-server/config"
	"main-server/services"
)

type AuthHandler struct {
	db     *sql.DB
	config *config.Config
}

func NewAuthHandler(db *sql.DB, config *config.Config) *AuthHandler {
	return &AuthHandler{
		db:     db,
		config: config,
	}
}

// Just ADD these methods to your existing AuthHandler:

// Add this method - shows splash page with login form
func (h *AuthHandler) ShowSplash(c echo.Context) error {
	// Check if already logged in
	if user := h.getCurrentUser(c); user != nil {
		return c.Redirect(http.StatusFound, "/dashboard")
	}

	data := map[string]interface{}{
		"Title": "Login",
	}
	return c.Render(http.StatusOK, "splash.html", data)
}

func (h *AuthHandler) LoginForm(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	// Simple validation
	if email == "" || password == "" {
		data := map[string]interface{}{
			"Title": "Login",
			"Error": "Email and password are required",
			"Email": email,
		}
		return c.Render(http.StatusOK, "splash.html", data)
	}

	// Check hardcoded credentials
	if email == "admin@throtle.io" && password == "Password123!" {
		// Create user object
		user := &services.User{
			ID:    "admin-123",
			Email: email,
			Name:  "Administrator",
		}

		// Save session using your existing helper
		if err := h.saveUserSession(c, user); err != nil {
			data := map[string]interface{}{
				"Title": "Login",
				"Error": "Failed to create session",
				"Email": email,
			}
			return c.Render(http.StatusOK, "splash.html", data)
		}

		return c.Redirect(http.StatusFound, "/dashboard")
	}

	// Invalid credentials
	data := map[string]interface{}{
		"Title": "Login",
		"Error": "Invalid email or password",
		"Email": email,
	}
	return c.Render(http.StatusOK, "splash.html", data)
}

// Add these helper methods (moved from your services)
func (h *AuthHandler) saveUserSession(c echo.Context, user *services.User) error {
	sess, _ := session.Get("session", c)
	sess.Values["user_id"] = user.ID
	sess.Values["user_email"] = user.Email
	sess.Values["user_name"] = user.Name
	sess.Values["logged_in"] = true
	return sess.Save(c.Request(), c.Response())
}

func (h *AuthHandler) getCurrentUser(c echo.Context) *services.User {
	sess, err := session.Get("session", c)
	if err != nil {
		return nil
	}

	loggedIn, ok := sess.Values["logged_in"].(bool)
	if !ok || !loggedIn {
		return nil
	}

	return &services.User{
		ID:    sess.Values["user_id"].(string),
		Email: sess.Values["user_email"].(string),
		Name:  sess.Values["user_name"].(string),
	}
}

func (h *AuthHandler) Logout(c echo.Context) error {
	// Clear local session
	h.clearUserSession(c)

	// Redirect to home page
	return c.Redirect(http.StatusFound, "/")
}

// Add this helper method
func (h *AuthHandler) clearUserSession(c echo.Context) error {
	sess, _ := session.Get("session", c)
	sess.Values = make(map[interface{}]interface{})
	return sess.Save(c.Request(), c.Response())
}

// UPDATE existing Dashboard method - change redirect target
func (h *AuthHandler) Dashboard(c echo.Context) error {
	user := h.getCurrentUser(c) // Use our helper instead of service
	if user == nil {
		return c.Redirect(http.StatusFound, "/") // CHANGED from /auth/login
	}

	data := map[string]interface{}{
		"Title": "Dashboard",
		"User":  user,
	}

	return c.Render(http.StatusOK, "dashboard.html", data)
}

func (h *AuthHandler) ShowLogin(c echo.Context) error {
	return c.Render(http.StatusOK, "login.html", map[string]interface{}{
		"Title": "Login",
	})
}

func (h *AuthHandler) Login(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	// Simple hardcoded check for now
	if email == "admin@example.com" && password == "password123" {
		user := &services.User{
			ID:    "1",
			Email: email,
			Name:  "Admin User",
		}

		h.saveUserSession(c, user)
		return c.Redirect(http.StatusFound, "/app/dashboard")
	}

	return c.Render(http.StatusOK, "login.html", map[string]interface{}{
		"Title": "Login",
		"Error": "Invalid credentials",
	})
}

func (h *AuthHandler) ShowRegister(c echo.Context) error {
	return c.Render(http.StatusOK, "register.html", map[string]interface{}{
		"Title": "Register",
	})
}

func (h *AuthHandler) Register(c echo.Context) error {
	// For now, just redirect to login
	return c.Redirect(http.StatusFound, "/auth/login")
}

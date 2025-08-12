package handlers

import (
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"

	"main-server/services"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) ShowLogin(c echo.Context) error {
	data := map[string]interface{}{
		"Title": "Login",
	}
	return c.Render(http.StatusOK, "auth/login.html", data)
}

func (h *AuthHandler) Login(c echo.Context) error {
	loginURL, state, err := h.authService.GetLoginURL()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to generate login URL")
	}

	// Store state in session for validation
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}
	sess.Values["oauth_state"] = state
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, loginURL)
}

func (h *AuthHandler) Callback(c echo.Context) error {
	code := c.QueryParam("code")
	state := c.QueryParam("state")

	if code == "" {
		return c.String(http.StatusBadRequest, "Missing authorization code")
	}

	// Get stored state from session
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}

	storedState, ok := sess.Values["oauth_state"].(string)
	if !ok {
		return c.String(http.StatusBadRequest, "Missing stored state")
	}

	// Handle the callback
	claimsData, err := h.authService.HandleCallback(code, state, storedState)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Authentication failed: "+err.Error())
	}

	// Store user session
	if err := h.authService.StoreUserSession(c, claimsData); err != nil {
		return c.String(http.StatusInternalServerError, "Failed to store session")
	}

	// Redirect to dashboard or home
	return c.Redirect(http.StatusFound, "/dashboard")
}

func (h *AuthHandler) Logout(c echo.Context) error {
	// Get ID token for Cognito logout
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}

	idToken, _ := sess.Values["id_token"].(string)

	// Clear local session
	if err := h.authService.ClearUserSession(c); err != nil {
		return err
	}

	// Redirect to Cognito logout if we have an ID token
	if idToken != "" {
		logoutURL := h.authService.GetLogoutURL("http://localhost:8080/")
		return c.Redirect(http.StatusFound, logoutURL)
	}

	// Otherwise just redirect to home
	return c.Redirect(http.StatusFound, "http://localhost:8080/")
}

func (h *AuthHandler) Dashboard(c echo.Context) error {
	user, err := h.authService.GetCurrentUser(c)
	if err != nil {
		return c.Redirect(http.StatusFound, "/auth/login")
	}

	data := map[string]interface{}{
		"Title": "Dashboard",
		"User":  user,
	}

	return c.Render(http.StatusOK, "dashboard.html", data)
}

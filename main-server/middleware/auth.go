package middleware

import (
	"context"
	"net/http"

	"main-server/models"

	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func AuthMiddleware(store *sessions.CookieStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := store.Get(r, "session")
			if err != nil || session.Values["user_id"] == nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			// Check if user is still active
			// userID := session.Values["user_id"].(int)
			// This would be loaded from DB in real implementation

			next.ServeHTTP(w, r)
		})
	}
}

func LoadContextMiddleware(db *sqlx.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, _ := r.Context().Value("session").(*sessions.Session)
			if session == nil {
				next.ServeHTTP(w, r)
				return
			}

			if userID, ok := session.Values["user_id"].(int); ok {
				var user models.User
				err := db.Get(&user, `
                    SELECT u.*, c.workspace_id 
                    FROM users u 
                    JOIN companies c ON u.company_id = c.id 
                    WHERE u.id = $1 AND u.is_active = true
                `, userID)

				if err == nil {
					ctx := context.WithValue(r.Context(), "user", &user)
					r = r.WithContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sess, err := session.Get("session", c)
			if err != nil {
				return c.Redirect(http.StatusFound, "/auth/login")
			}

			authenticated, ok := sess.Values["authenticated"].(bool)
			if !ok || !authenticated {
				return c.Redirect(http.StatusFound, "/auth/login")
			}

			return next(c)
		}
	}
}

func RequireUserTier(requiredTier string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sess, err := session.Get("session", c)
			if err != nil {
				return c.String(http.StatusForbidden, "Access denied")
			}

			userTier, ok := sess.Values["user_tier"].(string)
			if !ok {
				return c.String(http.StatusForbidden, "Access denied")
			}

			// Define tier hierarchy (adjust as needed)
			tierLevels := map[string]int{
				"basic":      1,
				"premium":    2,
				"business":   3,
				"enterprise": 4,
			}

			userLevel, userExists := tierLevels[userTier]
			requiredLevel, reqExists := tierLevels[requiredTier]

			if !userExists || !reqExists || userLevel < requiredLevel {
				return c.String(http.StatusForbidden, "Insufficient privileges")
			}

			return next(c)
		}
	}
}

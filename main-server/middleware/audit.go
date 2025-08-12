package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"main-server/models"

	"github.com/jmoiron/sqlx"
)

type AuditLog struct {
	UserID      string          `db:"user_id"`
	Action      string          `db:"action"`
	EntityType  string          `db:"entity_type"`
	EntityID    string          `db:"entity_id"`
	CompanyID   string          `db:"company_id"`
	WorkspaceID string          `db:"workspace_id"`
	Changes     json.RawMessage `db:"changes"`
	IPAddress   string          `db:"ip_address"`
	UserAgent   string          `db:"user_agent"`
	CreatedAt   time.Time       `db:"created_at"`
}

func AuditMiddleware(db *sqlx.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only audit state-changing requests
			if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
				next.ServeHTTP(w, r)
				return
			}

			// Capture response
			rec := httptest.NewRecorder()
			next.ServeHTTP(rec, r)

			// Log if successful
			if rec.Code < 400 {
				go logAudit(db, r)
			}

			// Copy response to original writer
			for k, v := range rec.Header() {
				w.Header()[k] = v
			}
			w.WriteHeader(rec.Code)
			w.Write(rec.Body.Bytes())
		})
	}
}

func logAudit(db *sqlx.DB, r *http.Request) {
	user, _ := r.Context().Value("user").(*models.User)
	if user == nil {
		return
	}

	action := r.Method + " " + r.URL.Path

	log := AuditLog{
		UserID:    user.ID,
		Action:    action,
		CompanyID: user.CompanyID,
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
		CreatedAt: time.Now(),
	}

	_, _ = db.NamedExec(`
        INSERT INTO audit_logs (user_id, action, company_id, ip_address, user_agent, created_at)
        VALUES (:user_id, :action, :company_id, :ip_address, :user_agent, :created_at)
    `, log)
}

func getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

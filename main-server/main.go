package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"main-server/config"
	"main-server/handlers"
	customMiddleware "main-server/middleware"
	"main-server/services"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type App struct {
	db      *sql.DB
	metrics *Metrics
	echo    *echo.Echo
	config  *config.Config
}

type Metrics struct {
	httpRequests *prometheus.CounterVec
	httpDuration *prometheus.HistogramVec
	auditLogs    *prometheus.CounterVec
}

type AWSConfig struct {
	Region              string
	CognitoUserPoolID   string
	CognitoClientID     string
	CognitoClientSecret string
	CognitoDomain       string
}

type Config struct {
	Port        string
	Environment string
	Version     string
	SessionKey  string
	Debug       bool
	LogLevel    string
	Database    *DatabaseConfig
	Services    *ServiceConfig
	BaseURL     string
	AWS         AWSConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type ServiceConfig struct {
	OnboardingURL string
	StaticURL     string
	S3Endpoint    string
	PrometheusURL string
	GrafanaURL    string
}

var store *sessions.CookieStore

func Load() (*config.Config, error) {

	env := os.Getenv("GO_ENV")
	if "" == env {
		env = "development"
	}

	// Load .env file
	if err := godotenv.Load(".env." + env); err != nil {
		log.Println("No .env file found")
	}

	dbConfig := &DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "appuser"),
		Password: getEnv("DB_PASSWORD", "apppaswword"),
		Name:     getEnv("DB_NAME", "appdb"),
		SSLMode:  getEnv("DB_SSL_MODE", "disable"),
	}

	serviceConfig := &ServiceConfig{}

	debug := true
	if getEnv("DEBUG", "true") == "false" {
		debug = false
	}

	return &config.Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", env),
		Version:     getEnv("VERSION", "1.0.0"),
		SessionKey:  getEnv("SESSION_KEY", "default-dev-key"),
		Debug:       debug,
		LogLevel:    getEnv("LOG_LEVEL", "debug"),
		Database:    dbConfig,
		Services:    serviceConfig,
		AWS: AWSConfig{
			Region:              getEnv("AWS_REGION", "us-east-1"),
			CognitoUserPoolID:   getEnv("COGNITO_USER_POOL_ID", ""),
			CognitoClientID:     getEnv("COGNITO_CLIENT_ID", ""),
			CognitoClientSecret: getEnv("COGNITO_CLIENT_SECRET", ""),
			CognitoDomain:       getEnv("COGNITO_USER_POOL_DOMAIN", ""),
		},
		BaseURL: getEnv("BASE_URL", "http://localhost:8080"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func NewMetrics() *Metrics {
	return &Metrics{
		httpRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		httpDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		auditLogs: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "audit_logs_total",
				Help: "Total number of audit log events",
			},
			[]string{"event_type", "user_id", "resource"},
		),
	}
}

func (m *Metrics) Register() {
	prometheus.MustRegister(m.httpRequests)
	prometheus.MustRegister(m.httpDuration)
	prometheus.MustRegister(m.auditLogs)
}

// Echo middleware for metrics collection
func (app *App) metricsMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			duration := time.Since(start).Seconds()
			status := c.Response().Status
			if err != nil {
				// Handle Echo HTTP errors
				if he, ok := err.(*echo.HTTPError); ok {
					status = he.Code
				} else {
					status = http.StatusInternalServerError
				}
			}

			app.metrics.httpRequests.WithLabelValues(
				c.Request().Method,
				c.Path(),
				fmt.Sprintf("%d", status),
			).Inc()

			app.metrics.httpDuration.WithLabelValues(
				c.Request().Method,
				c.Path(),
			).Observe(duration)

			// Example audit log
			app.LogAudit("page_view", "anonymous", c.Path())

			return err
		}
	}
}

func (app *App) LogAudit(eventType, userID, resource string) {
	app.metrics.auditLogs.WithLabelValues(eventType, userID, resource).Inc()

	// Also log to stdout for debugging
	log.Printf("AUDIT: event=%s user=%s resource=%s timestamp=%s",
		eventType, userID, resource, time.Now().Format(time.RFC3339))
}

// Template renderer for Echo
type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	config, err := Load()
	if err != nil {
		log.Fatal("Failed to load configuration")
	}

	// Initialize metrics
	metrics := NewMetrics()
	metrics.Register()

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Password,
		config.Database.Name,
		config.Database.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Initialize Echo
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(session.Middleware(store))

	// Template Renderer
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}
	e.Renderer = renderer

	authService, err := services.NewAuthService(
		config.AWS.CognitoClientID,
		config.AWS.CognitoClientSecret,
		config.AWS.CognitoDomain,
		config.BaseURL+"/auth/callback",
		config.AWS.Region,
		config.AWS.CognitoUserPoolID,
	)
	userService := services.NewUserService(db)

	app := &App{
		db:      db,
		metrics: metrics,
		echo:    e,
		config:  config,
	}

	// Add metrics middleware
	e.Use(app.metricsMiddleware())

	// In setupRoutes() function
	authHandler := handlers.NewAuthHandler(authService, userService)
	homeHandler := handlers.NewHomeHandler(config, db)

	// Routes
	e.GET("/", authHandler.ShowSplash)

	// Auth routes
	auth := e.Group("/auth")
	auth.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	auth.GET("/login", authHandler.ShowSplash)
	auth.POST("/login", authHandler.LoginForm)
	// auth.GET("/callback", authHandler.Callback)
	auth.POST("/logout", authHandler.Logout)

	main := auth.Group("/main")
	main.GET("/health", homeHandler.Health)
	// main.GET("/audit", app.auditHandler)
	// main.POST("/audit", app.auditHandler)
	main.GET("/dashboard", authHandler.Dashboard, customMiddleware.RequireAuth())

	// Protected routes

	fmt.Printf("Main server starting on port %s\n", config.Port)
	fmt.Printf("Metrics available at http://localhost:%s/metrics\n", config.Port)

	// Start server
	e.Logger.Fatal(e.Start(":" + config.Port))
}

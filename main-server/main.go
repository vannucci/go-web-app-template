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
	"os"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type App struct {
	db     *sql.DB
	echo   *echo.Echo
	config *config.Config
}

var store *sessions.CookieStore

// Template renderer for Echo
type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration")
	}

	// Create upload directory
	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		log.Fatal("Failed to create upload directory:", err)
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Initialize session store
	store = sessions.NewCookieStore([]byte(cfg.SessionKey))

	// Initialize Echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(session.Middleware(store))

	// Template Renderer
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}
	e.Renderer = renderer

	app := &App{
		db:     db,
		echo:   e,
		config: cfg,
	}

	// In setupRoutes() function
	authHandler := handlers.NewAuthHandler(db, cfg)
	homeHandler := handlers.NewHomeHandler(db, cfg)
	uploadHandler := handlers.NewUploadHandler(cfg.UploadDir) // New local upload handler

	// Simplified routes
	e.GET("/", homeHandler.Home)
	e.GET("/health", homeHandler.Health)

	// Auth routes
	auth := e.Group("/auth")
	auth.GET("/login", authHandler.ShowLogin)
	auth.POST("/login", authHandler.Login)
	auth.POST("/logout", authHandler.Logout)
	auth.GET("/register", authHandler.ShowRegister)
	auth.POST("/register", authHandler.Register)

	// Protected routes
	protected := e.Group("/app")
	protected.Use(customMiddleware.RequireAuth())
	protected.GET("/dashboard", authHandler.Dashboard)
	protected.POST("/upload", uploadHandler.Upload)
	protected.GET("/uploads/*", uploadHandler.Serve) // Serve uploaded files

	// Metrics endpoint (keep this)
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	fmt.Printf("Server starting on port %s\n", cfg.Port)
	fmt.Printf("Upload directory: %s\n", cfg.UploadDir)
	fmt.Printf("Metrics: http://localhost:%s/metrics\n", cfg.Port)

	e.Logger.Fatal(e.Start(":" + cfg.Port))

	defer app.Shutdown()
}

func (app *App) Shutdown() {
	app.db.Close()
}

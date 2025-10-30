# Go Web App Template

Quick and dirty setup for local development.

## Prerequisites

- Go 1.19+
- Docker & Docker Compose
- Make

## Quick Start

1. **Start infrastructure**
   ```bash
   docker-compose up -d postgres static-server
   Run the app
   ```

cd main-server
make run-dev
Access the app

App: http://localhost:8080
Health: http://localhost:8080/health
Metrics: http://localhost:8080/metrics
Login
Email: admin@example.com
Password: password123
File Structure
├── main-server/ # Go web application
├── static/ # Static files (CSS, JS, images)
├── docker-compose.yml # Infrastructure (postgres, nginx)
└── README.md
Development
Database: PostgreSQL on port 5432
Uploads: Saved to ./uploads/ locally
Config: Edit .env.development for settings
Useful Commands

# Reset database

make db-reset

# Run tests

make test

# Stop everything

docker-compose down
That's it! 🚀

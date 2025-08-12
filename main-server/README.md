# Ad Tech Platform

A simple, maintainable ad tech platform for managing audience files and campaign destinations.

## Features

- **Multi-tenant architecture** with workspaces and companies
- **Role-based access control** (Super Admin, Workspace Admin, Company Admin, User)
- **Audience file management** with S3 storage
- **Multiple destination platforms** (Meta, Google Ads, The Trade Desk)
- **Campaign execution** with external API integration
- **Audit logging** for compliance
- **Server-side rendering** with HTMX for interactivity

## Quick Start

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 14 or higher
- AWS S3 bucket (or local storage for development)

### Installation

1. Clone the repository:
```bash
git clone https://main-server.git
cd adtech-platform
```

2. Install dependencies:
```bash
go mod download
```

3. Copy the environment file:
```bash
cp .env.development .env
```

4. Update `.env` with your configuration

5. Run database migrations:
```bash
psql -U your_user -d your_database -f database/migrations/001_initial_schema.sql
psql -U your_user -d your_database -f database/migrations/002_audit_logs.sql
```

6. Run the application:
```bash
go run main.go
```

The application will be available at `http://localhost:8080`

## Project Structure

- `/handlers` - HTTP request handlers
- `/models` - Data models and validation
- `/middleware` - Authentication, permissions, audit logging
- `/templates` - HTML templates
- `/static` - CSS and JavaScript files
- `/database` - Database migrations and connection
- `/services` - Business logic (storage, campaign execution)

## Adding New Destinations

To add a new destination platform:

1. Create model in `/models/destinations/`
2. Create handler in `/handlers/destinations/`
3. Create template in `/templates/destinations/`
4. Register routes in `main.go`

Example for adding a new destination:

```go
// models/destinations/new_platform.go
type NewPlatformDestination struct {
    BaseDestination
    // Add platform-specific fields
}

// Implement required methods
func (n *NewPlatformDestination) Validate() error { }
func (n *NewPlatformDestination) GetType() string { }
```

## Development

### Running locally

```bash
# Install air for hot reload (optional)
go install github.com/cosmtrek/air@latest

# Run with hot reload
air

# Or run directly
go run main.go
```

### Testing

```bash
go test ./...
```

## Deployment

### Using Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o adtech-platform

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/adtech-platform .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
CMD ["./adtech-platform"]
```

### Environment Variables

- `DATABASE_URL` - PostgreSQL connection string
- `PORT` - Server port (default: 8080)
- `SESSION_KEY` - 32-byte session encryption key
- `AWS_REGION` - AWS region for S3
- `AWS_ACCESS_KEY_ID` - AWS access key
- `AWS_SECRET_ACCESS_KEY` - AWS secret key
- `S3_BUCKET` - S3 bucket name for audience files
- `EXTERNAL_API_URL` - External service API endpoint
- `EXTERNAL_API_KEY` - External service API key

## Security

- All passwords are bcrypt hashed
- Session-based authentication with secure cookies
- CSRF protection on all state-changing operations
- Input validation on all forms
- SQL injection prevention via parameterized queries
- XSS protection via HTML escaping

## License

MIT


## Instructions for Getting Started

1. **Create a new GitHub repository**
2. **Copy all the files from both artifacts into your repository structure**
3. **Initialize the Go module**:
   ```bash
   go mod init main-server
   go mod tidy
   ```
4. **Set up your PostgreSQL database**
5. **Configure your `.env` file**
6. **Run the migrations**
7. **Start the application**

The platform is designed to be simple and maintainable, with clear separation of concerns and explicit code over magic. Each destination type has its own form, handler, and validation logic, making it easy to add new platforms or modify existing ones.


------

# Multi-Service Docker Compose Project

This project sets up a complete development environment with Go services, PostgreSQL, Prometheus, and Grafana.

## Project Structure

```
your-project/
├── docker-compose.yml
├── prometheus.yml
├── nginx.conf
├── dev-start.sh
├── dev-stop.sh
├── dev-main.sh
├── dev-logs.sh
├── dev-metrics.sh
├── onboarding-server/
│   ├── Dockerfile
│   ├── main.go
│   ├── go.mod
│   └── go.sum
├── main-server/
│   ├── config/
│   ├── database/
│       └── migrations/
│           ├── 001_initial_schema.sql
│           └── 002_audit_logs.sql
│   ├── handlers/
│       └── destinations/
│           └── meta.go
│   ├── middleware/
│           ├── audit.go
│           └── auth.go
│   ├── models/
│       ├── user.go
│       ├── workspace.go
│       └── destinations/
│           ├── base.go
│           └── meta.go
│   ├── services/
│       └── storage.go
│   ├── static/
│       ├── css/
│           └── style.css
│       └── js/
│           └── htmx.min.js
│   ├── templates/
│       └── layout.html
│       └── destinations/
│           └── meta.html
│   ├── Dockerfile
│   ├── main.go
│   ├── README.md
│   ├── NOTES.md
│   ├── go.mod
│   └── go.sum
├── static/
│   ├── styles.css
│   ├── app.js
│   └── index.html
├── grafana/
│   ├── provisioning/
│   │   ├── datasources/
│   │   │   └── prometheus.yml
│   │   └── dashboards/
│   │       └── dashboard.yml
│   └── dashboards/
│       └── main-dashboard.json
├── init-db/
│   └── (place any .sql initialization files here)
```

## Setup Instructions

1. Create the directory structure above
2. Copy all files below into their respective locations
3. Make scripts executable: `chmod +x dev-*.sh`
4. Start services: `./dev-start.sh`
5. Start main server: `./dev-main.sh`

## Service URLs

- Main Server: http://localhost:8080
- Onboarding Server: http://localhost:8081
- Static Files: http://localhost:8082
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin123)
- PostgreSQL: localhost:5432

---

## Quick Start Commands

```bash
# Make scripts executable
chmod +x dev-*.sh

# Start all supporting services
./dev-start.sh

# In another terminal, start your main server
./dev-main.sh

# Generate some test metrics
curl http://localhost:8080
curl -X POST http://localhost:8080/audit

# View metrics
curl http://localhost:8080/metrics
```

## Development Tips

- Edit files in `main-server/` and restart with `./dev-main.sh` for development
- Supporting services stay running independently
- Check logs with `./dev-logs.sh`
- Grafana dashboard will be automatically provisioned
- Use Prometheus at http://localhost:9090 for metric queries and debugging
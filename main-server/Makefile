.PHONY: run build test clean dev install-tools

# Run locally (assumes infrastructure is running)
run-dev:
	@echo "🏃 Running in DEVELOPMENT mode..."
	GO_ENV=development go run .

# Build binary
build:
	@echo "🔨 Building main-server..."
	go build -o bin/main-server .

# Run tests
test:
	@echo "🧪 Running tests..."
	go test -v ./...

# Test with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report: coverage.html"

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install development tools
install-tools:
	go install github.com/cosmtrek/air@latest
	@echo "✅ Installed development tools"

# Database operations
db-migrate:
	@echo "🗃️  Running migrations..."
	go run . -migrate

db-reset:
	@echo "⚠️  Resetting database..."
	docker-compose exec postgres psql -U appuser -d appdb -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	$(MAKE) db-migrate
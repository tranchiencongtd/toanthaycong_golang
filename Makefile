# note: call scripts from /scripts

# Docker commands
.PHONY: docker-up docker-down docker-build docker-logs docker-clean

# Start all services
docker-up:
	docker-compose up -d

# Stop all services
docker-down:
	docker-compose down

# Build and start services
docker-build:
	docker-compose up -d --build

# View logs
docker-logs:
	docker-compose logs -f

# Clean up everything
docker-clean:
	docker-compose down -v --remove-orphans
	docker system prune -f

# Database commands
.PHONY: db-migrate db-seed db-reset db-connect migrate-up migrate-down

# Connect to PostgreSQL
db-connect:
	docker-compose exec postgres psql -U postgres -d toanthaycong

# Backup database
db-backup:
	docker-compose exec postgres pg_dump -U postgres toanthaycong > backup.sql

# Restore database
db-restore:
	docker-compose exec -T postgres psql -U postgres toanthaycong < backup.sql

# Run database migrations
migrate-up:
	go run cmd/migrate/main.go

# Generate sqlc code
sqlc-generate:
	sqlc generate

# Install sqlc (if not installed)
install-sqlc:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Development commands
.PHONY: dev run test deps

# Install dependencies
deps:
	go mod download
	go mod tidy
	go mod vendor

# Run in development mode
dev:
	go run cmd/_your_app_/main.go

# Run tests
test:
	go test -v ./...

# Build application
build:
	go build -o bin/app cmd/_your_app_/main.go

# Setup development environment
setup:
	@echo "Setting up development environment..."
	make deps
	make install-sqlc
	@echo "Starting Docker containers..."
	make docker-up
	@echo "Waiting for database to be ready..."
	sleep 10
	@echo "Running migrations..."
	make migrate-up
	@echo "Generating SQL code..."
	make sqlc-generate
	@echo "Setup completed!"

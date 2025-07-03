# lưu ý: gọi scripts từ thư mục /scripts

# Lệnh Docker
.PHONY: docker-up docker-down docker-build docker-logs docker-clean

# Khởi động tất cả services
docker-up:
	docker-compose up -d

# Dừng tất cả services
docker-down:
	docker-compose down

# Build và khởi động services
docker-build:
	docker-compose up -d --build

# Xem logs
docker-logs:
	docker-compose logs -f

# Dọn dẹp tất cả
docker-clean:
	docker-compose down -v --remove-orphans
	docker system prune -f

# Lệnh Database
.PHONY: db-migrate db-seed db-reset db-connect migrate-up migrate-down

# Kết nối tới PostgreSQL
db-connect:
	docker-compose exec postgres psql -U postgres -d toanthaycong

# Khởi tạo dữ liệu mẫu
db-seed:
	go run cmd/seed/main.go

# Reset database (migration + seed)
db-reset:
	make migrate-up
	make db-seed

# Sao lưu database
db-backup:
	docker-compose exec postgres pg_dump -U postgres toanthaycong > backup.sql

# Khôi phục database
db-restore:
	docker-compose exec -T postgres psql -U postgres toanthaycong < backup.sql

# Chạy database migrations
migrate-up:
	go run cmd/migrate/main.go

# Tạo mã sqlc
sqlc-generate:
	sqlc generate

# Cài đặt sqlc (nếu chưa cài)
install-sqlc:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Lệnh Development
.PHONY: dev run test deps

# Cài đặt dependencies
deps:
	go mod tidy
	go mod download
	go mod download
	go mod tidy
	go mod vendor

# Tạo thư mục vendor
vendor:
	go mod vendor
	@echo "✅ Vendor directory created with all dependencies"

# Xóa vendor
clean-vendor:
	rm -rf vendor/
	@echo "🗑️ Vendor directory removed"

# Chạy ở chế độ development
dev:
	go run cmd/api/main.go

# Chạy API server
api:
	go run cmd/api/main.go

# Chạy tests
test:
	go test -v ./...

# Build ứng dụng
build:
	go build -o bin/app cmd/_your_app_/main.go

# Thiết lập môi trường development
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

# Setup với dữ liệu mẫu
setup-with-data:
	@echo "Setting up development environment with sample data..."
	make deps
	make install-sqlc
	@echo "Starting Docker containers..."
	make docker-up
	@echo "Waiting for database to be ready..."
	sleep 10
	@echo "Running migrations..."
	make migrate-up
	@echo "Seeding sample data..."
	make db-seed
	@echo "Generating SQL code..."
	make sqlc-generate
	@echo "✅ Setup completed with sample data!"

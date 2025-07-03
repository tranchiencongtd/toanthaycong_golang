# ToanthayconG Golang API

REST API cho nền tảng học tập trực tuyến ToanthayconG được xây dựng bằng Go (Golang).

## 🚀 Quick Start

### 1. Setup Environment

```bash
# Copy file .env
cp .env.example .env

# Cập nhật thông tin database trong .env
```

### 2. Start Database

```bash
# Khởi động Docker containers
make docker-up

# Chờ database sẵn sàng và chạy migrations
make migrate-up

# Seed dữ liệu mẫu
make db-seed
```

### 3. Install Dependencies

```bash
make deps
```

### 4. Run API Server

```bash
make api
# hoặc
make dev
```

API sẽ chạy tại: `http://localhost:8080`

## 📚 API Documentation

### Base URL
```
http://localhost:8080/api/v1
```

### Health Check
```
GET /api/v1/health
```

### 📂 Categories API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET    | `/categories` | Lấy danh sách categories |
| GET    | `/categories/:id` | Lấy category theo ID |
| POST   | `/categories` | Tạo category mới |
| PUT    | `/categories/:id` | Cập nhật category |
| DELETE | `/categories/:id` | Xóa category |

**Query Parameters:**
- `page` (int): Trang hiện tại (default: 1)
- `limit` (int): Số items per page (default: 10, max: 100)
- `parent_id` (string): Filter theo parent category

**Example:**
```bash
GET /api/v1/categories?page=1&limit=10&parent_id=null
```

### 👥 Users API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET    | `/users` | Lấy danh sách users |
| GET    | `/users/:id` | Lấy user theo ID |
| POST   | `/users` | Tạo user mới |
| PUT    | `/users/:id` | Cập nhật user |
| DELETE | `/users/:id` | Xóa user |

**Query Parameters:**
- `page`, `limit`: Pagination
- `role` (string): Filter theo role (student, instructor, admin)

### 📖 Courses API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET    | `/courses` | Lấy danh sách courses |
| GET    | `/courses/:id` | Lấy course theo ID |
| POST   | `/courses` | Tạo course mới |
| PUT    | `/courses/:id` | Cập nhật course |
| DELETE | `/courses/:id` | Xóa course |

**Query Parameters:**
- `page`, `limit`: Pagination
- `category_id` (string): Filter theo category
- `instructor_id` (string): Filter theo instructor
- `level` (string): Filter theo level (beginner, intermediate, advanced)
- `status` (string): Filter theo status (draft, pending, published, archived)

### 🏷️ Tags API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET    | `/tags` | Lấy danh sách tags |
| GET    | `/tags/:id` | Lấy tag theo ID |
| POST   | `/tags` | Tạo tag mới |
| PUT    | `/tags/:id` | Cập nhật tag |
| DELETE | `/tags/:id` | Xóa tag |

**Query Parameters:**
- `page`, `limit`: Pagination
- `search` (string): Tìm kiếm theo tên tag

## 📋 Request/Response Examples

### Create Category
```bash
POST /api/v1/categories
Content-Type: application/json

{
  "name": "Web Development",
  "slug": "web-development",
  "description": "Learn web development",
  "icon_url": "/icons/web.svg",
  "parent_id": "uuid-of-parent",
  "sort_order": 1
}
```

### Create User
```bash
POST /api/v1/users
Content-Type: application/json

{
  "email": "user@example.com",
  "username": "newuser",
  "password": "password123",
  "first_name": "John",
  "last_name": "Doe",
  "role": "student"
}
```

### Create Course
```bash
POST /api/v1/courses
Content-Type: application/json

{
  "title": "Learn React.js",
  "slug": "learn-reactjs",
  "description": "Complete React.js course",
  "short_description": "Learn React from scratch",
  "instructor_id": "uuid-of-instructor",
  "category_id": "uuid-of-category",
  "price": 999000,
  "discount_price": 699000,
  "language": "vi",
  "level": "beginner",
  "requirements": ["HTML", "CSS", "JavaScript"],
  "what_you_learn": ["React Components", "State Management"],
  "target_audience": ["Beginners", "Developers"]
}
```

## 📊 Standard API Response Format

### Success Response
```json
{
  "success": true,
  "message": "Data retrieved successfully",
  "data": {
    // Response data
  }
}
```

### Error Response
```json
{
  "success": false,
  "message": "Error message",
  "error": "Detailed error information"
}
```

### Paginated Response
```json
{
  "success": true,
  "message": "Data retrieved successfully",
  "data": {
    "items": [...],
    "pagination": {
      "total": 100,
      "page": 1,
      "limit": 10,
      "total_page": 10
    }
  }
}
```

## 🛠️ Available Make Commands

```bash
# Docker commands
make docker-up          # Start containers
make docker-down        # Stop containers
make docker-build       # Build and start containers
make docker-logs        # View logs
make docker-clean       # Clean up containers

# Database commands
make migrate-up         # Run migrations
make db-seed           # Seed sample data
make db-reset          # Reset database (migrate + seed)
make db-connect        # Connect to PostgreSQL

# Development commands
make deps              # Install dependencies
make dev               # Run in development mode
make api               # Run API server
make test              # Run tests
make build             # Build application

# Setup commands
make setup             # Complete setup (no data)
make setup-with-data   # Complete setup with sample data
```

## 🗄️ Database Schema

### Tables
- `users` - Người dùng hệ thống
- `instructor_profiles` - Hồ sơ giảng viên
- `categories` - Danh mục khóa học
- `courses` - Khóa học
- `tags` - Thẻ tag
- `course_tags` - Liên kết course và tag
- `enrollments` - Đăng ký khóa học
- `lessons` - Bài học
- `reviews` - Đánh giá

### Sample Data
Chạy `make db-seed` để có dữ liệu mẫu:
- Admin: `admin@toanthaycong.com` / `password`
- Instructor: `instructor@toanthaycong.com` / `password`
- Student: `student1@example.com` / `password`

## 🔧 Environment Variables

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password123
DB_NAME=toanthaycong
DB_SSL_MODE=disable

# Server
SERVER_PORT=8080
ENV=development
```

## 🧪 Testing

```bash
# Run all tests
make test

# Test specific package
go test ./internal/api/handlers/...

# Test with coverage
go test -cover ./...
```

## 📝 TODO

- [ ] Authentication & Authorization
- [ ] File upload cho images/videos
- [ ] Email service
- [ ] Payment integration
- [ ] Real-time notifications
- [ ] Caching với Redis
- [ ] Rate limiting
- [ ] API documentation với Swagger

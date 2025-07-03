# Toàn Thầy Cồng - E-Learning Platform

Nền tảng học trực tuyến giống Udemy được xây dựng bằng Go và PostgreSQL.

## 🏗️ Kiến trúc

- **Backend**: Go với Gin framework
- **Database**: PostgreSQL 15
- **SQL Generator**: SQLC
- **Migration**: golang-migrate
- **Containerization**: Docker & Docker Compose
- **Admin Interface**: pgAdmin

## 🚀 Quick Start

### 1. Thiết lập môi trường development

```bash
# Clone repository
git clone <your-repo-url>
cd toanthaycong

# Chạy setup tự động
make setup
```

Lệnh `make setup` sẽ thực hiện:
- Cài đặt Go dependencies
- Cài đặt SQLC
- Khởi động Docker containers (PostgreSQL + pgAdmin)
- Chạy database migrations
- Generate SQL code

### 2. Thủ công (nếu cần)

```bash
# Cài đặt dependencies
make deps

# Khởi động database
make docker-up

# Chạy migrations
make migrate-up

# Generate SQL code
make sqlc-generate
```

## 📊 Database Schema

### Core Tables

- **users**: Người dùng (students, instructors, admins)
- **instructor_profiles**: Thông tin chi tiết giảng viên
- **categories**: Danh mục khóa học (có hierarchy)
- **courses**: Khóa học
- **course_sections**: Chương trong khóa học
- **course_lectures**: Bài giảng (video, article, quiz, file)

### Learning Tables

- **enrollments**: Đăng ký khóa học
- **lecture_progress**: Tiến độ học tập
- **course_reviews**: Đánh giá khóa học
- **certificates**: Chứng chỉ hoàn thành

### E-commerce Tables

- **carts**: Giỏ hàng
- **orders**: Đơn hàng
- **order_items**: Chi tiết đơn hàng
- **coupons**: Mã giảm giá
- **wishlists**: Danh sách yêu thích

### Additional Tables

- **course_questions**: Hỏi đáp
- **course_answers**: Câu trả lời
- **notifications**: Thông báo
- **tags**: Tags cho khóa học
- **course_announcements**: Thông báo khóa học

## 🔧 Cấu trúc thư mục

```
├── cmd/                    # Applications
│   ├── migrate/           # Database migration tool
│   └── _your_app_/        # Main application
├── internal/              # Private application code
│   ├── db/               # Database layer
│   │   ├── migrations/   # SQL migration files
│   │   └── queries/      # SQLC query files
│   ├── app/              # Application logic
│   └── pkg/              # Internal packages
├── pkg/                   # Public packages
├── configs/               # Configuration files
├── init/                  # System initialization
│   └── postgres/         # Database init scripts
└── docker-compose.yml     # Docker setup
```

## 🐳 Docker Services

### PostgreSQL Database
- **Port**: 5432
- **Username**: postgres
- **Password**: password123
- **Database**: toanthaycong

### pgAdmin
- **URL**: http://localhost:5050
- **Email**: admin@admin.com
- **Password**: admin123

## 📝 Environment Variables

File `.env` đã được tạo với:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password123
DB_NAME=toanthaycong
DB_SSL_MODE=disable

# Application
APP_ENV=development
APP_PORT=8080
```

## 🛠️ Development Commands

```bash
# Database
make docker-up              # Start containers
make docker-down            # Stop containers
make db-connect             # Connect to PostgreSQL
make migrate-up             # Run migrations
make sqlc-generate          # Generate SQL code

# Dependencies
make deps                   # Install & update dependencies
make vendor                 # Create vendor directory
make clean-vendor           # Remove vendor directory

# Development
make dev                    # Run application
make test                   # Run tests
make build                  # Build binary

# Utilities
make docker-logs            # View logs
make docker-clean           # Clean containers
```

## 📦 Vendor Directory

Project sử dụng **vendor** directory để:
- **Lưu trữ dependencies cục bộ** (offline builds)
- **Đảm bảo version consistency** across environments
- **CI/CD reliability** (không phụ thuộc external networks)

### Vendor Commands:
```bash
# Tạo vendor với tất cả dependencies
make vendor

# Xóa vendor (sử dụng go modules)
make clean-vendor

# Build using vendor
go build -mod=vendor
```

### Khi nào commit vendor?
- ✅ **Applications**: Có thể commit vendor/ 
- ❌ **Libraries**: Không commit vendor/
- 🎯 **Production**: Khuyến nghị dùng vendor

## 📦 SQLC Usage

### 1. Thêm query mới

Tạo file `.sql` trong `internal/db/queries/`:

```sql
-- name: GetUserCourses :many
SELECT c.* FROM courses c
JOIN enrollments e ON c.id = e.course_id
WHERE e.user_id = $1
ORDER BY e.enrolled_at DESC;
```

### 2. Generate code

```bash
make sqlc-generate
```

### 3. Sử dụng trong Go

```go
courses, err := q.GetUserCourses(ctx, userID)
```

## 🗄️ Sample Data

Database được khởi tạo với:
- Admin user: `admin@toanthaycong.com`
- Instructor: `instructor@toanthaycong.com`
- Sample students
- Categories (Lập trình, Thiết kế, Kinh doanh...)
- Sample course: "React.js từ cơ bản đến nâng cao"

## 🔍 API Design (Sắp tới)

API sẽ theo REST principles:

```
GET    /api/v1/courses              # List courses
POST   /api/v1/courses              # Create course
GET    /api/v1/courses/:id          # Get course
PUT    /api/v1/courses/:id          # Update course
DELETE /api/v1/courses/:id          # Delete course

GET    /api/v1/users/me/enrollments # My enrollments
POST   /api/v1/courses/:id/enroll   # Enroll course
```

## 🚧 Next Steps

1. **API Layer**: Tạo REST API với Gin
2. **Authentication**: JWT-based auth
3. **File Upload**: Video/image upload
4. **Payment**: Stripe/VNPay integration
5. **Frontend**: React.js application
6. **Deployment**: Kubernetes/Docker Swarm

## 📚 Resources

- [SQLC Documentation](https://docs.sqlc.dev/)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [Gin Framework](https://gin-gonic.com/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)

---

**Happy Coding! 🎓**

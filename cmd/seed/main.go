package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	// Tải biến môi trường từ file .env
	if err := godotenv.Load(); err != nil {
		log.Println("Không tìm thấy file .env, sử dụng biến môi trường hệ thống")
	}

	// Thiết lập kết nối database
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "password123")
	dbName := getEnv("DB_NAME", "toanthaycong")
	dbSSLMode := getEnv("DB_SSL_MODE", "disable")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName, dbSSLMode)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal("Không thể kết nối đến database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Không thể ping database:", err)
	}

	log.Println("✅ Kết nối database thành công!")

	// Đọc và thực thi file SQL seed data
	sqlFile := "init/postgres/01_init.sql"
	log.Printf("📖 Đọc file SQL: %s", sqlFile)

	content, err := os.ReadFile(sqlFile)
	if err != nil {
		log.Fatal("❌ Không thể đọc file SQL:", err)
	}

	log.Println("🚀 Đang thực thi dữ liệu mẫu...")

	// Bắt đầu transaction để đảm bảo tính toàn vẹn dữ liệu
	tx, err := db.Begin()
	if err != nil {
		log.Fatal("❌ Không thể bắt đầu transaction:", err)
	}
	defer tx.Rollback()

	// Thực thi SQL
	if _, err := tx.Exec(string(content)); err != nil {
		log.Fatal("❌ Không thể thực thi SQL:", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Fatal("❌ Không thể commit transaction:", err)
	}

	log.Println("✅ Khởi tạo dữ liệu mẫu thành công!")
	log.Println("")
	log.Println("📊 Dữ liệu mẫu đã được thêm:")
	log.Println("   - Extensions: uuid-ossp, citext")
	log.Println("   - Categories: Lập trình, Thiết kế, Kinh doanh, Marketing, Nhiếp ảnh")
	log.Println("   - Users: Admin, Instructor, Students")
	log.Println("   - Instructor profiles")
	log.Println("   - Tags: JavaScript, React, Node.js, Python, etc.")
	log.Println("   - Sample course: React.js từ cơ bản đến nâng cao")
	log.Println("")
	log.Println("🎯 Bạn có thể đăng nhập với:")
	log.Println("   Admin: admin@toanthaycong.com / password")
	log.Println("   Instructor: instructor@toanthaycong.com / password")
	log.Println("   Student: student1@example.com / password")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

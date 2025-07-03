package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strings"

    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
    _ "github.com/jackc/pgx/v5/stdlib"
    "github.com/joho/godotenv"
)

func main() {
    // Tải các biến môi trường từ file .env
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

    // Kết nối trực tiếp đến database đích (không drop/create)
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

    log.Println("Kết nối database thành công!")

    // Chạy migrations (chỉ update)
    driver, err := postgres.WithInstance(db, &postgres.Config{})
    if err != nil {
        log.Fatal("Không thể tạo migration driver:", err)
    }

    // Lấy đường dẫn tuyệt đối cho Windows
    pwd, err := os.Getwd()
    if err != nil {
        log.Fatal("Không thể lấy thư mục hiện tại:", err)
    }
    
    // Tạo đường dẫn migration tuyệt đối và chuẩn hóa cho Windows
    migrationDir := filepath.Join(pwd, "internal", "db", "migrations")
    // Chuyển đổi backslash thành forward slash cho file URL
    migrationPath := "file://" + strings.ReplaceAll(migrationDir, "\\", "/")

    m, err := migrate.NewWithDatabaseInstance(
        migrationPath,
        "postgres", driver)
    if err != nil {
        log.Fatal("Không thể tạo migration instance:", err)
    }

    // Kiểm tra version hiện tại
    version, dirty, err := m.Version()
    if err != nil && err != migrate.ErrNilVersion {
        log.Fatal("Không thể lấy migration version:", err)
    }
    
    if err == migrate.ErrNilVersion {
        log.Println("Database chưa có migration nào, sẽ chạy tất cả migrations...")
    } else {
        log.Printf("Migration version hiện tại: %d, dirty: %t", version, dirty)
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        log.Fatal("Không thể chạy migrations:", err)
    }

    if err == migrate.ErrNoChange {
        log.Println("Không có migration mới nào để chạy!")
    } else {
        log.Println("Migrations hoàn thành thành công!")
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
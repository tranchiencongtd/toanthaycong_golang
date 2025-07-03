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

	// Bước 1: Kết nối đến postgres database để tạo database mới
	postgresDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=%s",
		dbUser, dbPassword, dbHost, dbPort, dbSSLMode)

	postgresDB, err := sql.Open("pgx", postgresDSN)
	if err != nil {
		log.Fatal("Không thể kết nối đến postgres database:", err)
	}
	defer postgresDB.Close()

	if err := postgresDB.Ping(); err != nil {
		log.Fatal("Không thể ping postgres database:", err)
	}

	log.Println("Kết nối đến postgres database thành công!")

	// Bước 2: Drop và tạo lại database để đảm bảo clean state
	dropDBQuery := fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)
	_, err = postgresDB.Exec(dropDBQuery)
	if err != nil {
		log.Fatal("Không thể drop database:", err)
	}
	log.Printf("Đã drop database %s (nếu tồn tại)", dbName)

	createDBQuery := fmt.Sprintf("CREATE DATABASE %s", dbName)
	_, err = postgresDB.Exec(createDBQuery)
	if err != nil {
		log.Fatal("Không thể tạo database:", err)
	}
	log.Printf("Đã tạo database %s thành công!", dbName)

	// Bước 3: Kết nối đến database mới tạo
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

	// Chạy migrations (di chuyển cơ sở dữ liệu)
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

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		// Nếu database ở trạng thái dirty, reset về version 0 và thử lại
		if strings.Contains(err.Error(), "Dirty database") {
			log.Println("Database ở trạng thái dirty, đang reset...")
			if forceErr := m.Force(-1); forceErr != nil {
				log.Fatal("Không thể force reset migration:", forceErr)
			}
			// Thử chạy lại migration
			if err := m.Up(); err != nil && err != migrate.ErrNoChange {
				log.Fatal("Không thể chạy migrations sau khi reset:", err)
			}
		} else {
			log.Fatal("Không thể chạy migrations:", err)
		}
	}

	log.Println("Migrations hoàn thành thành công!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

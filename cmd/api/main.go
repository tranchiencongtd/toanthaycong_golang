package main

import (
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"internal/api/routes"
	"internal/config"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Setup logger
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	// Connect to database
	db, err := sql.Open("pgx", cfg.DatabaseURL())
	if err != nil {
		logrus.Fatal("Failed to connect to database: ", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		logrus.Fatal("Failed to ping database: ", err)
	}

	logrus.Info("Successfully connected to database")

	// Setup routes
	router := routes.SetupRoutes(db)

	// Start server
	logrus.Infof("Starting server on port %s", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		logrus.Fatal("Failed to start server: ", err)
	}
}

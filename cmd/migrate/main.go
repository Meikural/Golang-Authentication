package main

import (
	"log"

	"auth-service/internal/db"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("🚀 Starting migration runner...")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using system environment variables")
	}

	// Connect to database
	db.ConnectDB()
	defer db.CloseDB()

	// Run migrations
	migrationsPath := "scripts/migrations"
	db.RunMigrations(migrationsPath)

	log.Println("✅ Database migrations finished successfully.")
}

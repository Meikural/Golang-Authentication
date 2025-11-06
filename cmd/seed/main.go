package main

import (
	"log"

	"auth-service/internal/db"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("🚀 Starting Super Admin seeder...")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using system environment variables")
	}

	// Connect to DB
	db.ConnectDB()
	defer db.CloseDB()

	// Run seed
	db.SeedInitialData()

	log.Println("✅ Seeding completed successfully.")
}

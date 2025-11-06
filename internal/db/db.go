package db

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB is the global database pool instance
var DB *pgxpool.Pool

// ConnectDB initializes a PostgreSQL connection pool
func ConnectDB() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("❌ DATABASE_URL not set in environment")
	}

	// Parse connection config
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Fatalf("❌ Unable to parse DATABASE_URL: %v", err)
	}

	// Optional pool configuration tuning
	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.HealthCheckPeriod = 30 * time.Second
	cfg.MaxConnLifetime = 1 * time.Hour

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create new pool
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL: %v", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("❌ Database ping failed: %v", err)
	}

	DB = pool
	log.Println("✅ Connected to PostgreSQL successfully")
}

// CloseDB cleanly shuts down the database pool
func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("🛑 Database connection closed")
	}
}

// CheckHealth performs a simple ping to verify DB health (optional)
func CheckHealth() bool {
	if DB == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := DB.Ping(ctx); err != nil {
		log.Printf("⚠️ Database health check failed: %v", err)
		return false
	}
	return true
}

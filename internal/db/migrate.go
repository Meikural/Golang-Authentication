package db

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// RunMigrations scans the migrations folder and applies any new SQL files
func RunMigrations(migrationsPath string) {
	if DB == nil {
		log.Fatal("❌ Database not initialized. Call ConnectDB() first.")
	}

	ctx := context.Background()

	// Ensure schema_migrations table exists
	_, err := DB.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			applied_at TIMESTAMP DEFAULT NOW()
		);
	`)
	if err != nil {
		log.Fatalf("❌ Failed to ensure schema_migrations table: %v", err)
	}

	// Read migration directory
	files, err := ioutil.ReadDir(migrationsPath)
	if err != nil {
		log.Fatalf("❌ Failed to read migrations directory: %v", err)
	}

	// Sort files (001_*, 002_*, etc.)
	sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })

	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".sql") {
			continue
		}

		fileName := f.Name()
		log.Printf("🔍 Checking migration: %s", fileName)

		// Check if already applied
		var count int
		err := DB.QueryRow(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE name=$1;", fileName).Scan(&count)
		if err != nil {
			log.Fatalf("❌ Failed to check migration status: %v", err)
		}
		if count > 0 {
			log.Printf("⏭️  Skipping already applied migration: %s", fileName)
			continue
		}

		// Read SQL file
		sqlPath := filepath.Join(migrationsPath, fileName)
		content, err := os.ReadFile(sqlPath)
		if err != nil {
			log.Fatalf("❌ Failed to read migration file %s: %v", fileName, err)
		}

		// Execute SQL
		log.Printf("🚀 Applying migration: %s", fileName)
		start := time.Now()

		if _, err := DB.Exec(ctx, string(content)); err != nil {
			log.Fatalf("❌ Migration failed (%s): %v", fileName, err)
		}

		log.Printf("✅ Applied %s in %v", fileName, time.Since(start))
	}

	log.Println("🎉 All migrations completed successfully!")
}

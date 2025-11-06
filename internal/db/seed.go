package db

import (
	"context"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// SeedInitialData ensures the super_admin role and Super Admin user exist
func SeedInitialData() {
	if DB == nil {
		log.Fatal("❌ Database not initialized. Call ConnectDB() first.")
	}

	ctx := context.Background()

	superAdminEmail := os.Getenv("SUPERADMIN_EMAIL")
	superAdminPassword := os.Getenv("SUPERADMIN_PASSWORD")

	if superAdminEmail == "" || superAdminPassword == "" {
		log.Fatal("❌ SUPERADMIN_EMAIL or SUPERADMIN_PASSWORD not set in .env")
	}

	// 1️⃣ Ensure super_admin role exists
	var roleID int
	err := DB.QueryRow(ctx, "SELECT id FROM roles WHERE name=$1;", "super_admin").Scan(&roleID)
	if err != nil {
		log.Println("⚠️  super_admin role not found, creating...")
		err = DB.QueryRow(ctx,
			"INSERT INTO roles (name, description) VALUES ($1, $2) RETURNING id;",
			"super_admin", "Has all system permissions",
		).Scan(&roleID)
		if err != nil {
			log.Fatalf("❌ Failed to create super_admin role: %v", err)
		}
	} else {
		log.Printf("✅ Found super_admin role (id=%d)", roleID)
	}

	// 2️⃣ Check if Super Admin user already exists
	var userID int
	err = DB.QueryRow(ctx, "SELECT id FROM users WHERE email=$1;", superAdminEmail).Scan(&userID)
	if err != nil {
		log.Println("⚠️  Super Admin user not found, creating...")

		// Hash password
		hashed, err := bcrypt.GenerateFromPassword([]byte(superAdminPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("❌ Failed to hash Super Admin password: %v", err)
		}

		// Insert user
		err = DB.QueryRow(ctx,
			"INSERT INTO users (email, password_hash, is_active, created_at, updated_at) VALUES ($1, $2, TRUE, NOW(), NOW()) RETURNING id;",
			superAdminEmail, string(hashed),
		).Scan(&userID)
		if err != nil {
			log.Fatalf("❌ Failed to create Super Admin user: %v", err)
		}

		log.Printf("✅ Created Super Admin user (id=%d)", userID)
	} else {
		log.Printf("✅ Found Super Admin user (id=%d)", userID)
	}

	// 3️⃣ Link user to super_admin role (if not already)
	var count int
	err = DB.QueryRow(ctx,
		"SELECT COUNT(*) FROM user_roles WHERE user_id=$1 AND role_id=$2;", userID, roleID,
	).Scan(&count)
	if err != nil {
		log.Fatalf("❌ Failed to check user_roles mapping: %v", err)
	}

	if count == 0 {
		_, err = DB.Exec(ctx, "INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2);", userID, roleID)
		if err != nil {
			log.Fatalf("❌ Failed to link Super Admin user to role: %v", err)
		}
		log.Println("✅ Linked Super Admin user to super_admin role.")
	} else {
		log.Println("⏭️  Super Admin user already linked to super_admin role.")
	}

	log.Println("🎉 Super Admin seeding complete at", time.Now().Format(time.RFC1123))
}

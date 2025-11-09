package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"auth-service/internal/db"
	"auth-service/internal/handlers"
	"auth-service/internal/middleware"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using system environment variables")
	}

	db.ConnectDB()
	defer db.CloseDB()

	app := fiber.New()

	app.Get("api/v1/health", func(c *fiber.Ctx) error {
		dbStatus := "disconnected"
		if db.CheckHealth() {
			dbStatus = "connected"
		}

		return c.JSON(fiber.Map{
			"status": "ok",
			"db":     dbStatus,
		})
	})

	app.Get("api/v1/version", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": "auth-service",
			"version": "v0.1.0",
		})
	})
	app.Post("/api/v1/login", handlers.Login)
	app.Get("/api/v1/me", middleware.AuthRequired(), handlers.Me)
	app.Post("/api/v1/register", handlers.Register)

	app.Get("/api/v1/superadmin/policies", middleware.AuthRequired(), handlers.GetAllPolicies)
	app.Get("/api/v1/superadmin/policies/:name", middleware.AuthRequired(), handlers.GetPolicyByName)
	app.Post("/api/v1/superadmin/policies", middleware.AuthRequired(), handlers.UpsertPolicies)

	app.Get("/api/v1/admin/roles", middleware.AuthRequired(), handlers.GetRoles)
	app.Post("/api/v1/admin/roles", middleware.AuthRequired(), handlers.CreateRole)
	app.Post("/api/v1/admin/assign-role", middleware.AuthRequired(), handlers.AssignRole)
	app.Delete("/api/v1/admin/revoke-role", middleware.AuthRequired(), handlers.RevokeRole)

	app.Get("/api/v1/admin/users", middleware.AuthRequired(), handlers.ListUsers)
	app.Get("/api/v1/admin/users/:id", middleware.AuthRequired(), handlers.GetUserByID)
	app.Patch("/api/v1/admin/users/:id/status", middleware.AuthRequired(), handlers.UpdateUserStatus)
	app.Delete("/api/v1/admin/users/:id", middleware.AuthRequired(), handlers.DeleteUser)

	// ----------------------------------------------------
	// 6️⃣ Start Server
	// ----------------------------------------------------
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Auth service running on port %s", port)
	if err := app.Listen(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

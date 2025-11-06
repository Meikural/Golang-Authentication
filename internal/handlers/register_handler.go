package handlers

import (
	"context"
	"log"
	"strings"

	"auth-service/internal/db"
	"auth-service/internal/utils"
	jwtpkg "auth-service/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

// RegisterRequest – expected request body
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// POST /register
func Register(c *fiber.Ctx) error {
	ctx := context.Background()
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	// 1️⃣  Load registration policy
	var registrationMode string
	err := db.DB.QueryRow(ctx, "SELECT value FROM auth_policies WHERE name='registration_mode';").Scan(&registrationMode)
	if err != nil {
		log.Printf("⚠️  Could not fetch registration_mode: %v", err)
		registrationMode = `"super_admin_only"`
	}
	registrationMode = trimQuotes(registrationMode)

	// 2️⃣  Check access rules based on mode
	switch registrationMode {
	case "super_admin_only":
		// Must be logged in as super_admin
		user := c.Locals("user")
		if user == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Registration disabled: Super Admin only",
			})
		}
		claims := user.(*jwtpkg.CustomClaims)
		if !hasRole(claims.Roles, "super_admin") {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Only Super Admin can register users",
			})
		}

	case "restricted":
		// Only certain roles can register others
		user := c.Locals("user")
		if user == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Restricted registration: login required",
			})
		}
		claims := user.(*jwtpkg.CustomClaims)

		// Fetch allowed roles from DB
		var allowedRolesJSON string
		db.DB.QueryRow(ctx, "SELECT value FROM auth_policies WHERE name='allowed_roles_for_registration';").Scan(&allowedRolesJSON)
		allowedRoles := parseStringArray(allowedRolesJSON)

		if !hasAnyRole(claims.Roles, allowedRoles) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Your role cannot register users",
			})
		}

	case "open":
		// Public registration — no restrictions
	default:
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Invalid registration mode configuration",
		})
	}

	// 3️⃣  Hash password
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to hash password",
		})
	}

	// 4️⃣  Insert new user
	var userID int
	err = db.DB.QueryRow(ctx,
		"INSERT INTO users (email, password_hash, is_active, created_at, updated_at) VALUES ($1, $2, TRUE, NOW(), NOW()) RETURNING id;",
		req.Email, hash).Scan(&userID)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Email already exists or invalid",
		})
	}

	log.Printf("✅ Registered new user: %s (id=%d)", req.Email, userID)

	return c.JSON(fiber.Map{
		"message": "User registered successfully",
		"user": fiber.Map{
			"id":    userID,
			"email": req.Email,
		},
	})
}

// Utility helpers
func trimQuotes(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

func hasRole(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

func hasAnyRole(userRoles []string, allowed []string) bool {
	for _, ur := range userRoles {
		for _, ar := range allowed {
			if ur == ar {
				return true
			}
		}
	}
	return false
}

func parseStringArray(jsonStr string) []string {
	// expects something like ["admin","service"]
	var arr []string
	jsonStr = trimQuotes(jsonStr)
	jsonStr = jsonStr[1 : len(jsonStr)-1] // strip [ ]
	for _, v := range strings.Split(jsonStr, ",") {
		arr = append(arr, strings.Trim(strings.TrimSpace(v), "\""))
	}
	return arr
}

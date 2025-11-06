package handlers

import (
	"context"
	"log"
	"time"

	"auth-service/internal/db"
	"auth-service/internal/utils"
	jwtpkg "auth-service/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

// LoginRequest defines incoming payload for /login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserInfoResponse defines structure for /me output
type UserInfoResponse struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Roles     []string  `json:"roles"`
	CreatedAt time.Time `json:"created_at"`
	IsActive  bool      `json:"is_active"`
}

// POST /login
func Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	// Fetch user
	ctx := context.Background()
	var id int
	var email string
	var passwordHash string
	var isActive bool

	err := db.DB.QueryRow(ctx, "SELECT id, email, password_hash, is_active FROM users WHERE email=$1;", req.Email).
		Scan(&id, &email, &passwordHash, &isActive)

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	// Check if user is active
	if !isActive {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "User account is inactive",
		})
	}

	// Verify password
	if !utils.CheckPassword(req.Password, passwordHash) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	// Fetch user roles
	rows, err := db.DB.Query(ctx, `
		SELECT r.name FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1;
	`, id)
	if err != nil {
		log.Printf("⚠️  Failed to load roles: %v", err)
	}

	var roles []string
	for rows.Next() {
		var roleName string
		rows.Scan(&roleName)
		roles = append(roles, roleName)
	}

	// Generate JWT
	token, err := jwtpkg.GenerateAccessToken(id, email, roles)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate access token",
		})
	}

	// Return response
	return c.JSON(fiber.Map{
		"access_token": token,
		"user": fiber.Map{
			"id":     id,
			"email":  email,
			"roles":  roles,
			"status": "active",
		},
	})
}

// GET /me
func Me(c *fiber.Ctx) error {
	userClaims := c.Locals("user") // set by JWT middleware later
	if userClaims == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	claims, ok := userClaims.(*jwtpkg.CustomClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token claims",
		})
	}

	return c.JSON(fiber.Map{
		"user": fiber.Map{
			"id":    claims.UserID,
			"email": claims.Email,
			"roles": claims.Roles,
		},
		"issued_at":  claims.IssuedAt.Time.Format(time.RFC3339),
		"expires_at": claims.ExpiresAt.Time.Format(time.RFC3339),
	})
}

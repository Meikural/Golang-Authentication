package handlers

import (
	"context"
	"log"
	"strconv"
	"time"

	"auth-service/internal/db"
	jwtpkg "auth-service/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

// ✅ GET /admin/users

func ListUsers(c *fiber.Ctx) error {
	user := c.Locals("user")
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	claims := user.(*jwtpkg.CustomClaims)
	if !(hasRole(claims.Roles, "super_admin") || hasRole(claims.Roles, "admin")) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	ctx := context.Background()
	rows, err := db.DB.Query(ctx, `
		SELECT 
			u.id, 
			u.email, 
			u.is_active, 
			u.created_at, 
			COALESCE(array_agg(r.name) FILTER (WHERE r.name IS NOT NULL), '{}') AS roles
		FROM users u
		LEFT JOIN user_roles ur ON u.id = ur.user_id
		LEFT JOIN roles r ON ur.role_id = r.id
		GROUP BY u.id
		ORDER BY u.id;
	`)
	if err != nil {
		log.Printf("❌ Error fetching users: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}
	defer rows.Close()

	type userResp struct {
		ID        int       `json:"id"`
		Email     string    `json:"email"`
		IsActive  bool      `json:"is_active"`
		CreatedAt time.Time `json:"created_at"`
		Roles     []string  `json:"roles"`
	}

	var users []userResp

	for rows.Next() {
		var u userResp
		err := rows.Scan(&u.ID, &u.Email, &u.IsActive, &u.CreatedAt, &u.Roles)
		if err != nil {
			log.Printf("⚠️  Scan error: %v", err)
			continue
		}
		users = append(users, u)
	}

	return c.JSON(fiber.Map{"users": users})
}


// ✅ GET /admin/users/:id
func GetUserByID(c *fiber.Ctx) error {
	user := c.Locals("user")
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	claims := user.(*jwtpkg.CustomClaims)
	if !(hasRole(claims.Roles, "super_admin") || hasRole(claims.Roles, "admin")) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	idParam := c.Params("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	ctx := context.Background()
	var email string
	var isActive bool
	var createdAt time.Time

	// ✅ Correct: time.Time for timestamp
	err = db.DB.QueryRow(ctx, `
		SELECT email, is_active, created_at 
		FROM users 
		WHERE id=$1;
	`, userID).Scan(&email, &isActive, &createdAt)

	if err != nil {
		log.Printf("⚠️  QueryRow failed for user %d: %v", userID, err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	// ✅ Get roles (ensure no null array)
	rows, err := db.DB.Query(ctx, `
		SELECT r.name 
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1;
	`, userID)
	if err != nil {
		log.Printf("⚠️  Failed to fetch roles: %v", err)
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var roleName string
		rows.Scan(&roleName)
		roles = append(roles, roleName)
	}

	// ✅ Response
	return c.JSON(fiber.Map{
		"id":         userID,
		"email":      email,
		"is_active":  isActive,
		"created_at": createdAt,
		"roles":      roles,
	})
}


// ✅ PATCH /admin/users/:id/status
func UpdateUserStatus(c *fiber.Ctx) error {
	user := c.Locals("user")
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	claims := user.(*jwtpkg.CustomClaims)
	if !(hasRole(claims.Roles, "super_admin") || hasRole(claims.Roles, "admin")) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	idParam := c.Params("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var body struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON payload"})
	}

	ctx := context.Background()
	_, err = db.DB.Exec(ctx, "UPDATE users SET is_active=$1, updated_at=NOW() WHERE id=$2;", body.IsActive, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update user status"})
	}

	log.Printf("🔄 Updated user %d status to %v", userID, body.IsActive)
	return c.JSON(fiber.Map{"message": "User status updated successfully"})
}

// ✅ DELETE /admin/users/:id
func DeleteUser(c *fiber.Ctx) error {
	user := c.Locals("user")
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	claims := user.(*jwtpkg.CustomClaims)
	if !hasRole(claims.Roles, "super_admin") {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Only super_admin can delete users"})
	}

	idParam := c.Params("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	ctx := context.Background()
	_, err = db.DB.Exec(ctx, "DELETE FROM users WHERE id=$1;", userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete user"})
	}

	log.Printf("🗑️  Deleted user %d", userID)
	return c.JSON(fiber.Map{"message": "User deleted successfully"})
}

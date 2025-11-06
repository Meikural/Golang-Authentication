package handlers

import (
	"context"
	"log"

	"auth-service/internal/db"
	jwtpkg "auth-service/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

// ✅ GET /admin/roles
func GetRoles(c *fiber.Ctx) error {
	user := c.Locals("user")
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	claims := user.(*jwtpkg.CustomClaims)
	if !hasRole(claims.Roles, "super_admin") {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Only super_admin can view roles"})
	}

	ctx := context.Background()
	rows, err := db.DB.Query(ctx, "SELECT id, name, description FROM roles ORDER BY id;")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch roles"})
	}
	defer rows.Close()

	var roles []fiber.Map
	for rows.Next() {
		var id int
		var name, desc string
		rows.Scan(&id, &name, &desc)
		roles = append(roles, fiber.Map{
			"id":          id,
			"name":        name,
			"description": desc,
		})
	}

	return c.JSON(fiber.Map{"roles": roles})
}

// ✅ POST /admin/roles
func CreateRole(c *fiber.Ctx) error {
	user := c.Locals("user")
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
	claims := user.(*jwtpkg.CustomClaims)
	if !hasRole(claims.Roles, "super_admin") {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Only super_admin can create roles"})
	}

	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON payload"})
	}

	ctx := context.Background()
	_, err := db.DB.Exec(ctx, `
		INSERT INTO roles (name, description)
		VALUES ($1, $2)
		ON CONFLICT (name) DO NOTHING;
	`, body.Name, body.Description)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create role"})
	}

	log.Printf("✅ Role created: %s", body.Name)
	return c.JSON(fiber.Map{"message": "Role created successfully", "role": body})
}

// ✅ POST /admin/assign-role
func AssignRole(c *fiber.Ctx) error {
	user := c.Locals("user")
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
	claims := user.(*jwtpkg.CustomClaims)
	if !hasRole(claims.Roles, "super_admin") {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Only super_admin can assign roles"})
	}

	var body struct {
		UserID int    `json:"user_id"`
		Role   string `json:"role"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON payload"})
	}

	ctx := context.Background()
	var roleID int
	err := db.DB.QueryRow(ctx, "SELECT id FROM roles WHERE name=$1;", body.Role).Scan(&roleID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Role not found"})
	}

	_, err = db.DB.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING;
	`, body.UserID, roleID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to assign role"})
	}

	log.Printf("✅ Assigned role %s to user %d", body.Role, body.UserID)
	return c.JSON(fiber.Map{"message": "Role assigned successfully"})
}

// ✅ DELETE /admin/revoke-role
func RevokeRole(c *fiber.Ctx) error {
	user := c.Locals("user")
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
	claims := user.(*jwtpkg.CustomClaims)
	if !hasRole(claims.Roles, "super_admin") {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Only super_admin can revoke roles"})
	}

	var body struct {
		UserID int    `json:"user_id"`
		Role   string `json:"role"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON payload"})
	}

	ctx := context.Background()
	var roleID int
	err := db.DB.QueryRow(ctx, "SELECT id FROM roles WHERE name=$1;", body.Role).Scan(&roleID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Role not found"})
	}

	_, err = db.DB.Exec(ctx, `
		DELETE FROM user_roles WHERE user_id=$1 AND role_id=$2;
	`, body.UserID, roleID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to revoke role"})
	}

	log.Printf("🚫 Revoked role %s from user %d", body.Role, body.UserID)
	return c.JSON(fiber.Map{"message": "Role revoked successfully"})
}

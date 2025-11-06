package handlers

import (
	"context"
	"log"

	"auth-service/internal/db"
	jwtpkg "auth-service/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

// ✅ GET /superadmin/policies
func GetAllPolicies(c *fiber.Ctx) error {
	user := c.Locals("user")
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	claims := user.(*jwtpkg.CustomClaims)
	if !hasRole(claims.Roles, "super_admin") {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Only super_admin can view policies"})
	}

	ctx := context.Background()
	rows, err := db.DB.Query(ctx, "SELECT name, value FROM auth_policies ORDER BY id;")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch policies"})
	}
	defer rows.Close()

	policies := make(map[string]string)
	for rows.Next() {
		var name, value string
		rows.Scan(&name, &value)
		policies[name] = value
	}

	return c.JSON(fiber.Map{"policies": policies})
}

// ✅ GET /superadmin/policies/:name
func GetPolicyByName(c *fiber.Ctx) error {
	user := c.Locals("user")
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	claims := user.(*jwtpkg.CustomClaims)
	if !hasRole(claims.Roles, "super_admin") {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Only super_admin can view policy"})
	}

	name := c.Params("name")
	ctx := context.Background()

	var value string
	err := db.DB.QueryRow(ctx, "SELECT value FROM auth_policies WHERE name=$1;", name).Scan(&value)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Policy not found"})
	}

	return c.JSON(fiber.Map{"name": name, "value": value})
}

// ✅ POST /superadmin/policies
func UpsertPolicies(c *fiber.Ctx) error {
	user := c.Locals("user")
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	claims := user.(*jwtpkg.CustomClaims)
	if !hasRole(claims.Roles, "super_admin") {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Only super_admin can modify policies"})
	}

	body := make(map[string]string)
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON body"})
	}

	ctx := context.Background()
	for name, value := range body {
		log.Printf("⚙️  Updating policy: %s = %s", name, value)
		_, err := db.DB.Exec(ctx, `
			INSERT INTO auth_policies (name, value, updated_at)
			VALUES ($1, $2, NOW())
			ON CONFLICT (name)
			DO UPDATE SET value = EXCLUDED.value, updated_at = NOW();
		`, name, value)
		if err != nil {
			log.Printf("❌ Failed to update policy %s: %v", name, err)
		}
	}

	return c.JSON(fiber.Map{
		"message":  "Policies updated successfully",
		"policies": body,
	})
}

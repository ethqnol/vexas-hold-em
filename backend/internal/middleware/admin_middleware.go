package middleware

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/vexas-hold-em/backend/internal/db"
	"github.com/vexas-hold-em/backend/internal/models"
)

func AdminOnly(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: Missing or invalid token"})
	}

	idToken := strings.TrimPrefix(authHeader, "Bearer ")

	if db.AuthClient == nil || db.Client == nil {
		log.Println("AdminOnly: Firebase services not initialized")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	token, err := db.AuthClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		log.Printf("AdminOnly: Failed to verify token: %v\n", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: Invalid token"})
	}

	doc, err := db.Client.Collection("users").Doc(token.UID).Get(ctx)
	if err != nil {
		log.Printf("AdminOnly: Failed to get user doc for UID %s: %v\n", token.UID, err)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden: User not found"})
	}

	var user models.User
	if err := doc.DataTo(&user); err != nil {
		log.Printf("AdminOnly: Failed to parse user doc: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	if !user.IsAdmin {
		log.Printf("AdminOnly: User %s (%s) attempted admin action but is not admin\n", token.UID, user.Email)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden: Admin privileges required"})
	}

	return c.Next()
}

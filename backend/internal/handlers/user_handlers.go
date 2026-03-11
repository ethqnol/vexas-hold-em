package handlers

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/vexas-hold-em/backend/internal/db"
	"github.com/vexas-hold-em/backend/internal/models"
)

// retrieves info and balance
func GetUserProfile(c *fiber.Ctx) error {
	id := c.Params("id")

	if db.Client == nil {
		log.Println("GetUserProfile: Database not initialized")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database not initialized"})
	}

	doc, err := db.Client.Collection("users").Doc(id).Get(context.Background())
	if err != nil {
		log.Printf("GetUserProfile: User not found or error getting doc: %v\n", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	var user models.User
	if err := doc.DataTo(&user); err != nil {
		log.Printf("GetUserProfile: Failed to parse user data: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user data"})
	}

	return c.JSON(fiber.Map{
		"message": "User profile for " + id,
		"status":  "success",
		"data":    user,
	})
}

// creates new User doc if user doesnt already exist
func SyncUser(c *fiber.Ctx) error {
	id := c.Params("id")

	if db.Client == nil {
		log.Println("SyncUser: Database not initialized")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database not initialized"})
	}

	// Parse body for email/display name (from Firebase Auth)
	var req struct {
		Email       string `json:"email"`
		DisplayName string `json:"displayName"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	docRef := db.Client.Collection("users").Doc(id)
	doc, err := docRef.Get(context.Background())

	if err != nil || !doc.Exists() {
		newUser := models.User{
			Balance:     1000.0,
			Email:       req.Email,
			DisplayName: req.DisplayName,
		}

		_, err := docRef.Set(context.Background(), newUser)
		if err != nil {
			log.Printf("SyncUser: Error creating user doc: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user"})
		}

		return c.JSON(fiber.Map{
			"message": "User created successfully",
			"status":  "success",
			"data":    newUser,
		})
	}

	return c.JSON(fiber.Map{
		"message": "User already exists",
		"status":  "success",
	})
}

// retrieves active positions
func GetUserPortfolio(c *fiber.Ctx) error {
	id := c.Params("id")

	if db.Client == nil {
		log.Println("GetUserPortfolio: Database not initialized")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database not initialized"})
	}

	iter := db.Client.Collection("users").Doc(id).Collection("positions").Documents(context.Background())
	docs, err := iter.GetAll()
	if err != nil {
		log.Printf("GetUserPortfolio: Failed to fetch portfolio: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch portfolio"})
	}

	portfolio := make(map[string]models.Position)
	for _, doc := range docs {
		var pos models.Position
		if err := doc.DataTo(&pos); err == nil {
			portfolio[doc.Ref.ID] = pos
		} else {
			log.Printf("GetUserPortfolio: Failed to parse position document %s: %v\n", doc.Ref.ID, err)
		}
	}

	return c.JSON(fiber.Map{
		"message": "User portfolio for " + id,
		"status":  "success",
		"data":    portfolio,
	})
}

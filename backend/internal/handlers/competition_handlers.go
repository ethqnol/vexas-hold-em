package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/vexas-hold-em/backend/internal/db"
	"github.com/vexas-hold-em/backend/internal/models"
)

// list all active comps
func GetCompetitions(c *fiber.Ctx) error {
	if db.Client == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database not initialized"})
	}

	docs, err := db.Client.Collection("competitions").Documents(context.Background()).GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch competitions"})
	}

	var competitions []fiber.Map
	for _, doc := range docs {
		var comp models.Competition
		if err := doc.DataTo(&comp); err == nil {
			competitions = append(competitions, fiber.Map{
				"id":   doc.Ref.ID,
				"data": comp,
			})
		}
	}

	return c.JSON(fiber.Map{
		"competitions": competitions,
		"status":       "success",
	})
}

// get a single comp by ID
func GetCompetitionByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if db.Client == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database not initialized"})
	}

	doc, err := db.Client.Collection("competitions").Doc(id).Get(context.Background())
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Competition not found"})
	}

	var comp models.Competition
	if err := doc.DataTo(&comp); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse competition data"})
	}

	return c.JSON(fiber.Map{
		"competition": fiber.Map{
			"id":   doc.Ref.ID,
			"data": comp,
		},
		"status": "success",
	})
}

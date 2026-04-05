package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/vexas-hold-em/backend/internal/db"
	"github.com/vexas-hold-em/backend/internal/models"
)

func GetCompetitions(c *fiber.Ctx) error {
	if db.Client == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database not initialized"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	docs, err := db.Client.Collection("competitions").Documents(ctx).GetAll()
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

func GetCompetitionByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if db.Client == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database not initialized"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	doc, err := db.Client.Collection("competitions").Doc(id).Get(ctx)
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

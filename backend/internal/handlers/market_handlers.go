package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/vexas-hold-em/backend/internal/db"
	"github.com/vexas-hold-em/backend/internal/models"
)

// get all team markets for a comp
func GetMarketsByCompetition(c *fiber.Ctx) error {
	compID := c.Params("id")
	if db.Client == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database not initialized"})
	}

	docs, err := db.Client.Collection("competitions").Doc(compID).Collection("markets").Documents(context.Background()).GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch markets"})
	}

	var markets []fiber.Map
	for _, doc := range docs {
		var market models.Market
		if err := doc.DataTo(&market); err == nil {
			markets = append(markets, fiber.Map{
				"id":   doc.Ref.ID,
				"data": market,
			})
		}
	}

	return c.JSON(fiber.Map{
		"markets": markets,
		"status":  "success",
	})
}

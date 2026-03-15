package handlers

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
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
			// cpmm implied prob: yes% = nopool / (yespool + nopool)
			yesOdds := 0.5
			if total := market.YesPool + market.NoPool; total > 0 {
				yesOdds = market.NoPool / total
			}
			markets = append(markets, fiber.Map{
				"id":      doc.Ref.ID,
				"data":    market,
				"yesOdds": yesOdds,
			})
		}
	}

	return c.JSON(fiber.Map{
		"markets": markets,
		"status":  "success",
	})
}

// gets odds history for specific mkt (for chart)
func GetMarketHistory(c *fiber.Ctx) error {
	compID := c.Params("id")
	marketID := c.Params("marketId")
	if db.Client == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "db not initialized"})
	}

	docs, err := db.Client.Collection("competitions").Doc(compID).
		Collection("ledger").
		OrderBy("timestamp", firestore.Asc).
		Documents(context.Background()).GetAll()
	if err != nil {
		log.Printf("Error fetching history: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch history"})
	}

	type Point struct {
		Timestamp int64   `json:"t"`
		YesOdds   float64 `json:"y"`
	}

	points := []Point{}
	for _, doc := range docs {
		var tx models.Transaction
		if err := doc.DataTo(&tx); err == nil {
			if tx.TeamID == marketID {
				points = append(points, Point{
					Timestamp: tx.Timestamp,
					YesOdds:   tx.YesOdds,
				})
			}
		}
	}

	return c.JSON(fiber.Map{
		"history": points,
		"status":  "success",
	})
}

// gets odds history for all mkts in a comp (for global chart)
func GetCompetitionHistory(c *fiber.Ctx) error {
	compID := c.Params("id")
	if db.Client == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "db not initialized"})
	}

	docs, err := db.Client.Collection("competitions").Doc(compID).
		Collection("ledger").
		OrderBy("timestamp", firestore.Asc).
		Documents(context.Background()).GetAll()
	if err != nil {
		log.Printf("Error fetching global history: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch global history"})
	}

	type Point struct {
		Timestamp int64   `json:"t"`
		TeamID    string  `json:"teamId"`
		YesOdds   float64 `json:"y"`
	}

	points := []Point{}
	for _, doc := range docs {
		var tx models.Transaction
		if err := doc.DataTo(&tx); err == nil {
			points = append(points, Point{
				Timestamp: tx.Timestamp,
				TeamID:    tx.TeamID,
				YesOdds:   tx.YesOdds,
			})
		}
	}

	return c.JSON(fiber.Map{
		"history": points,
		"status":  "success",
	})
}

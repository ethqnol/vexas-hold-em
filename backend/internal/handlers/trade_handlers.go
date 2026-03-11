package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// body for buying shares
type TradeRequest struct {
	CompetitionID string  `json:"competitionId"`
	MarketID      string  `json:"marketId"` // team ID
	UserID        string  `json:"userId"`
	TradeType     string  `json:"tradeType"` // "YES" or "NO"
	Amount        float64 `json:"amount"`    // amt of VEX to spend
}

// body for selling shares
type SellRequest struct {
	CompetitionID string  `json:"competitionId"`
	MarketID      string  `json:"marketId"` // team ID
	UserID        string  `json:"userId"`
	Shares        float64 `json:"shares"` // # of shares to sell
}

// buy YES or NO shares for a team
func PlaceTrade(c *fiber.Ctx) error {
	var req TradeRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid JSON payload",
			"error":   err.Error(),
		})
	}

	// TODO: verify comp + market exist
	// TODO: AMM math — calc share price & output
	// TODO: update user balance + firestore
	return c.JSON(fiber.Map{
		"message": "Trade successfully received (Math Pending)",
		"status":  "success",
		"data":    req,
	})
}

// liquidate an existing position
func SellShares(c *fiber.Ctx) error {
	var req SellRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid JSON payload",
			"error":   err.Error(),
		})
	}

	// TODO: verify position exists w/ enough shares
	// TODO: AMM math — calc return value
	// TODO: update balance + position in firestore
	return c.JSON(fiber.Map{
		"message": "Sell order successfully received (Math Pending)",
		"status":  "success",
		"data":    req,
	})
}

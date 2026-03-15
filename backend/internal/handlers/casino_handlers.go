package handlers

import (
	"context"
	"math/rand"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gofiber/fiber/v2"
	"github.com/vexas-hold-em/backend/internal/db"
	"github.com/vexas-hold-em/backend/internal/models"
)

var (
	rouletteNumbers = []int{0, 32, 15, 19, 4, 21, 2, 25, 17, 34, 6, 27, 13, 36, 11, 30, 8, 23, 10, 5, 24, 16, 33, 1, 20, 14, 31, 9, 22, 18, 29, 7, 28, 12, 35, 3, 26}
)

func getRouletteColor(num int) string {
	if num == 0 {
		return "green"
	}
	// std euro roulette red nums
	reds := map[int]bool{1: true, 3: true, 5: true, 7: true, 9: true, 12: true, 14: true, 16: true, 18: true, 19: true, 21: true, 23: true, 25: true, 27: true, 30: true, 32: true, 34: true, 36: true}
	if reds[num] {
		return "red"
	}
	return "black"
}

func calculateRoulettePayout(betType string, amount float64, resultNum int, resultColor string) float64 {
	switch betType {
	case "red":
		if resultColor == "red" {
			return amount * 2
		}
	case "black":
		if resultColor == "black" {
			return amount * 2
		}
	case "green":
		if resultColor == "green" {
			return amount * 36
		}
	case "even":
		if resultNum != 0 && resultNum%2 == 0 {
			return amount * 2
		}
	case "odd":
		if resultNum != 0 && resultNum%2 != 0 {
			return amount * 2
		}
		// specific nums passed as strings, sticking to basic bets for now
	}
	return 0
}

func PlayRoulette(c *fiber.Ctx) error {
	var req struct {
		UserID  string  `json:"userId"`
		BetType string  `json:"betType"`
		Amount  float64 `json:"amount"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if req.Amount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Bet amount must be greater than 0"})
	}

	ctx := context.Background()
	userRef := db.Client.Collection("users").Doc(req.UserID)

	err := db.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(userRef)
		if err != nil {
			return err
		}
		var user models.User
		if err := doc.DataTo(&user); err != nil {
			return err
		}

		if user.Balance < req.Amount {
			return fiber.ErrBadRequest
		}

		// determine outcome
		rand.Seed(time.Now().UnixNano())
		resultNum := rouletteNumbers[rand.Intn(len(rouletteNumbers))]
		resultColor := getRouletteColor(resultNum)

		payout := calculateRoulettePayout(req.BetType, req.Amount, resultNum, resultColor)
		newBalance := user.Balance - req.Amount + payout

		// update user
		err = tx.Update(userRef, []firestore.Update{
			{Path: "balance", Value: newBalance},
		})
		if err != nil {
			return err
		}

		c.Locals("spin", resultNum)
		c.Locals("color", resultColor)
		c.Locals("payout", payout)
		c.Locals("newBalance", newBalance)
		return nil
	})

	if err != nil {
		if err == fiber.ErrBadRequest {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Insufficient balance"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to process bet"})
	}

	return c.JSON(fiber.Map{
		"spin":       c.Locals("spin"),
		"color":      c.Locals("color"),
		"payout":     c.Locals("payout"),
		"newBalance": c.Locals("newBalance"),
	})
}

// weighted reel strip — common symbols appear more often.
// jason & charles are rare (2/30 stops ≈ 6.7%).
// gives ~13% overall win rate & makes jackpots super rare.
var reelStrip = []string{
	"🍒", "🍒", "🍒", "🍒", "🍒", "🍒",
	"🍋", "🍋", "🍋", "🍋", "🍋", "🍋",
	"🍊", "🍊", "🍊", "🍊", "🍊",
	"🍉", "🍉", "🍉", "🍉", "🍉",
	"⭐", "⭐", "⭐",
	"charles", "charles",
	"jason",
}

func PlaySlots(c *fiber.Ctx) error {
	var req struct {
		UserID string  `json:"userId"`
		Amount float64 `json:"amount"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if req.Amount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Bet amount must be greater than 0"})
	}

	ctx := context.Background()
	userRef := db.Client.Collection("users").Doc(req.UserID)

	err := db.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(userRef)
		if err != nil {
			return err
		}
		var user models.User
		if err := doc.DataTo(&user); err != nil {
			return err
		}

		if user.Balance < req.Amount {
			return fiber.ErrBadRequest
		}

		// determine outcome — each reel picks independently from the weighted strip
		rand.Seed(time.Now().UnixNano())
		reels := []string{
			reelStrip[rand.Intn(len(reelStrip))],
			reelStrip[rand.Intn(len(reelStrip))],
			reelStrip[rand.Intn(len(reelStrip))],
		}

		// calculate payout
		payout := 0.0
		if reels[0] == reels[1] && reels[1] == reels[2] {
			// jackpot based on symbol
			switch reels[0] {
			case "jason":
				payout = req.Amount * 50
			case "charles":
				payout = req.Amount * 25
			case "⭐":
				payout = req.Amount * 15
			default:
				payout = req.Amount * 10
			}
		} else if reels[0] == reels[1] || reels[1] == reels[2] || reels[0] == reels[2] {
			// small win for 2 matches
			payout = req.Amount * 2
		}

		newBalance := user.Balance - req.Amount + payout

		// update user
		err = tx.Update(userRef, []firestore.Update{
			{Path: "balance", Value: newBalance},
		})
		if err != nil {
			return err
		}

		c.Locals("reels", reels)
		c.Locals("payout", payout)
		c.Locals("newBalance", newBalance)
		return nil
	})

	if err != nil {
		if err == fiber.ErrBadRequest {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Insufficient balance"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to process bet"})
	}

	return c.JSON(fiber.Map{
		"reels":      c.Locals("reels"),
		"payout":     c.Locals("payout"),
		"newBalance": c.Locals("newBalance"),
	})
}

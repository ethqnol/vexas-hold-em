package handlers

import (
	"context"
	"math"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gofiber/fiber/v2"
	"github.com/vexas-hold-em/backend/internal/db"
	"github.com/vexas-hold-em/backend/internal/models"
)

const tradeTimeout = 10 * time.Second

type TradeRequest struct {
	CompetitionID string  `json:"competitionId"`
	MarketID      string  `json:"marketId"`
	UserID        string  `json:"userId"`
	TradeType     string  `json:"tradeType"`
	Amount        float64 `json:"amount"`
}

type SellRequest struct {
	CompetitionID string  `json:"competitionId"`
	MarketID      string  `json:"marketId"`
	UserID        string  `json:"userId"`
	TradeType     string  `json:"tradeType"`
	Shares        float64 `json:"shares"`
}

func PlaceTrade(c *fiber.Ctx) error {
	if db.Client == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "db not initialized"})
	}

	var req TradeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON"})
	}

	if req.TradeType != "YES" && req.TradeType != "NO" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "tradeType must be YES or NO"})
	}
	if req.Amount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "amount must be > 0"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), tradeTimeout)
	defer cancel()

	var shares, newYesPool, newNoPool float64

	err := db.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		compDoc, err := tx.Get(db.Client.Collection("competitions").Doc(req.CompetitionID))
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, "competition not found")
		}
		var comp models.Competition
		if err := compDoc.DataTo(&comp); err != nil || comp.Status != "active" {
			return fiber.NewError(fiber.StatusBadRequest, "competition is not active")
		}

		marketRef := db.Client.Collection("competitions").Doc(req.CompetitionID).Collection("markets").Doc(req.MarketID)
		marketDoc, err := tx.Get(marketRef)
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, "market not found")
		}
		var market models.Market
		if err := marketDoc.DataTo(&market); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to parse market")
		}

		userRef := db.Client.Collection("users").Doc(req.UserID)
		userDoc, err := tx.Get(userRef)
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		}
		var user models.User
		if err := userDoc.DataTo(&user); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to parse user")
		}
		if user.Balance < req.Amount {
			return fiber.NewError(fiber.StatusBadRequest, "insufficient balance")
		}

		// CPMM math: maintain constant product k = yesPool * noPool
		yesPool := market.YesPool
		noPool := market.NoPool
		k := yesPool * noPool

		var sharesInside, newYesPoolInside, newNoPoolInside float64
		if req.TradeType == "YES" {
			newNoPoolInside = noPool + req.Amount
			newYesPoolInside = k / newNoPoolInside
			sharesInside = yesPool - newYesPoolInside
		} else {
			newYesPoolInside = yesPool + req.Amount
			newNoPoolInside = k / newYesPoolInside
			sharesInside = noPool - newNoPoolInside
		}

		sharesInside = math.Round(sharesInside*1e6) / 1e6

		if sharesInside <= 0 {
			return fiber.NewError(fiber.StatusBadRequest, "trade too small to produce shares")
		}
		
		shares = sharesInside
		newYesPool = newYesPoolInside
		newNoPool = newNoPoolInside

		positionRef := userRef.Collection("positions").Doc(req.MarketID)
		positionField := "yesShares"
		if req.TradeType == "NO" {
			positionField = "noShares"
		}

		if err := tx.Update(marketRef, []firestore.Update{
			{Path: "yesPool", Value: newYesPool},
			{Path: "noPool", Value: newNoPool},
		}); err != nil {
			return err
		}

		if err := tx.Update(userRef, []firestore.Update{
			{Path: "balance", Value: user.Balance - req.Amount},
		}); err != nil {
			return err
		}

		if err := tx.Set(positionRef, map[string]interface{}{
			positionField:   firestore.Increment(shares),
			"competitionId": req.CompetitionID,
			"teamName":      market.TeamName,
		}, firestore.MergeAll); err != nil {
			return err
		}

		if err := tx.Set(
			db.Client.Collection("competitions").Doc(req.CompetitionID).Collection("ledger").NewDoc(),
			models.Transaction{
				Timestamp:    time.Now().UnixMilli(),
				UserID:       req.UserID,
				TeamID:       req.MarketID,
				TradeType:    req.TradeType,
				AmountSpent:  req.Amount,
				SharesBought: shares,
				YesOdds:      newNoPool / (newYesPool + newNoPool),
			},
		); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if e, ok := err.(*fiber.Error); ok {
			return c.Status(e.Code).JSON(fiber.Map{"error": e.Message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "trade failed: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"status":      "success",
		"shares":      shares,
		"amountSpent": req.Amount,
		"newYesPool":  newYesPool,
		"newNoPool":   newNoPool,
	})
}

func SellShares(c *fiber.Ctx) error {
	if db.Client == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "db not initialized"})
	}

	var req SellRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON"})
	}

	if req.TradeType != "YES" && req.TradeType != "NO" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "tradeType must be YES or NO"})
	}
	if req.Shares <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "shares must be > 0"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), tradeTimeout)
	defer cancel()

	var payout, newYesPool, newNoPool float64

	err := db.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		compDoc, err := tx.Get(db.Client.Collection("competitions").Doc(req.CompetitionID))
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, "competition not found")
		}
		var comp models.Competition
		if err := compDoc.DataTo(&comp); err != nil || comp.Status != "active" {
			return fiber.NewError(fiber.StatusBadRequest, "competition is not active")
		}

		marketRef := db.Client.Collection("competitions").Doc(req.CompetitionID).Collection("markets").Doc(req.MarketID)
		marketDoc, err := tx.Get(marketRef)
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, "market not found")
		}
		var market models.Market
		if err := marketDoc.DataTo(&market); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to parse market")
		}

		userRef := db.Client.Collection("users").Doc(req.UserID)
		userDoc, err := tx.Get(userRef)
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		}
		var user models.User
		if err := userDoc.DataTo(&user); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to parse user")
		}

		positionRef := userRef.Collection("positions").Doc(req.MarketID)
		posDoc, err := tx.Get(positionRef)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "no position found")
		}
		var position models.Position
		if err := posDoc.DataTo(&position); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to parse position")
		}

		heldShares := position.YesShares
		positionField := "yesShares"
		if req.TradeType == "NO" {
			heldShares = position.NoShares
			positionField = "noShares"
		}
		if heldShares < req.Shares {
			return fiber.NewError(fiber.StatusBadRequest, "not enough shares to sell")
		}

		// Return shares to pool, extract VEX
		yesPool := market.YesPool
		noPool := market.NoPool
		k := yesPool * noPool

		var payoutInside, newYesPoolInside, newNoPoolInside float64
		if req.TradeType == "YES" {
			newYesPoolInside = yesPool + req.Shares
			newNoPoolInside = k / newYesPoolInside
			payoutInside = noPool - newNoPoolInside
		} else {
			newNoPoolInside = noPool + req.Shares
			newYesPoolInside = k / newNoPoolInside
			payoutInside = yesPool - newYesPoolInside
		}

		payoutInside = math.Round(payoutInside*1e6) / 1e6

		if payoutInside <= 0 {
			return fiber.NewError(fiber.StatusBadRequest, "sell too small to produce payout")
		}
		
		payout = payoutInside
		newYesPool = newYesPoolInside
		newNoPool = newNoPoolInside

		// update docs
		if err := tx.Update(marketRef, []firestore.Update{
			{Path: "yesPool", Value: newYesPool},
			{Path: "noPool", Value: newNoPool},
		}); err != nil {
			return err
		}

		if err := tx.Update(userRef, []firestore.Update{
			{Path: "balance", Value: user.Balance + payout},
		}); err != nil {
			return err
		}

		if err := tx.Set(positionRef, map[string]interface{}{
			positionField:   firestore.Increment(-req.Shares),
			"competitionId": req.CompetitionID,
			"teamName":      market.TeamName,
		}, firestore.MergeAll); err != nil {
			return err
		}

		if err := tx.Set(
			db.Client.Collection("competitions").Doc(req.CompetitionID).Collection("ledger").NewDoc(),
			models.Transaction{
				Timestamp:    time.Now().UnixMilli(),
				UserID:       req.UserID,
				TeamID:       req.MarketID,
				TradeType:    "SELL_" + req.TradeType,
				AmountSpent:  -payout,
				SharesBought: -req.Shares,
				YesOdds:      newNoPool / (newYesPool + newNoPool),
			},
		); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if e, ok := err.(*fiber.Error); ok {
			return c.Status(e.Code).JSON(fiber.Map{"error": e.Message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "sell failed: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"status":           "success",
		"payout":           payout,
		"sharesLiquidated": req.Shares,
		"newYesPool":       newYesPool,
		"newNoPool":        newNoPool,
	})
}

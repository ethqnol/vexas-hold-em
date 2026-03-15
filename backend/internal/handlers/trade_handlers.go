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
	TradeType     string  `json:"tradeType"` // "YES" or "NO"
	Shares        float64 `json:"shares"`    // # of shares to sell
}

// buy YES or NO shares for a team
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

	ctx := context.Background()

	// verify comp is active
	compDoc, err := db.Client.Collection("competitions").Doc(req.CompetitionID).Get(ctx)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "competition not found"})
	}
	var comp models.Competition
	if err := compDoc.DataTo(&comp); err != nil || comp.Status != "active" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "competition is not active"})
	}

	// fetch market + pools
	marketRef := db.Client.Collection("competitions").Doc(req.CompetitionID).Collection("markets").Doc(req.MarketID)
	marketDoc, err := marketRef.Get(ctx)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "market not found"})
	}
	var market models.Market
	if err := marketDoc.DataTo(&market); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to parse market"})
	}

	// fetch user + verify balance
	userRef := db.Client.Collection("users").Doc(req.UserID)
	userDoc, err := userRef.Get(ctx)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}
	var user models.User
	if err := userDoc.DataTo(&user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to parse user"})
	}
	if user.Balance < req.Amount {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "insufficient balance"})
	}

	// cpmm buy math: k = yespool * nopool
	// buying yes: inject amt into nopool, extract from yespool
	yesPool := market.YesPool
	noPool := market.NoPool
	k := yesPool * noPool

	var shares, newYesPool, newNoPool float64
	if req.TradeType == "YES" {
		newNoPool = noPool + req.Amount
		newYesPool = k / newNoPool
		shares = yesPool - newYesPool
	} else {
		newYesPool = yesPool + req.Amount
		newNoPool = k / newYesPool
		shares = noPool - newNoPool
	}

	shares = math.Round(shares*1e6) / 1e6 // round to 6 decimal places

	if shares <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "trade too small to produce shares"})
	}

	// position ref for incrementing shares
	positionRef := userRef.Collection("positions").Doc(req.MarketID)
	positionField := "yesShares"
	if req.TradeType == "NO" {
		positionField = "noShares"
	}

	// batch write — all or nothing
	batch := db.Client.Batch()
	batch.Update(marketRef, []firestore.Update{
		{Path: "yesPool", Value: newYesPool},
		{Path: "noPool", Value: newNoPool},
	})
	batch.Update(userRef, []firestore.Update{
		{Path: "balance", Value: user.Balance - req.Amount},
	})
	batch.Set(positionRef, map[string]interface{}{
		positionField:   firestore.Increment(shares),
		"competitionId": req.CompetitionID,
		"teamName":      market.TeamName,
	}, firestore.MergeAll)
	batch.Set(
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
	)

	if _, err := batch.Commit(ctx); err != nil {
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

// liquidate an existing position
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

	ctx := context.Background()

	// verify comp is active
	compDoc, err := db.Client.Collection("competitions").Doc(req.CompetitionID).Get(ctx)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "competition not found"})
	}
	var comp models.Competition
	if err := compDoc.DataTo(&comp); err != nil || comp.Status != "active" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "competition is not active"})
	}

	// fetch market
	marketRef := db.Client.Collection("competitions").Doc(req.CompetitionID).Collection("markets").Doc(req.MarketID)
	marketDoc, err := marketRef.Get(ctx)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "market not found"})
	}
	var market models.Market
	if err := marketDoc.DataTo(&market); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to parse market"})
	}

	// verify position has enough shares
	userRef := db.Client.Collection("users").Doc(req.UserID)
	positionRef := userRef.Collection("positions").Doc(req.MarketID)
	posDoc, err := positionRef.Get(ctx)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "no position found"})
	}
	var position models.Position
	if err := posDoc.DataTo(&position); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to parse position"})
	}

	heldShares := position.YesShares
	positionField := "yesShares"
	if req.TradeType == "NO" {
		heldShares = position.NoShares
		positionField = "noShares"
	}
	if heldShares < req.Shares {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "not enough shares to sell"})
	}

	// fetch user balance
	userDoc, err := userRef.Get(ctx)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}
	var user models.User
	if err := userDoc.DataTo(&user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to parse user"})
	}

	// cpmm sell math: req shares to pool, extract vex
	// selling yes: inject shares into yespool, extract from nopool
	yesPool := market.YesPool
	noPool := market.NoPool
	k := yesPool * noPool

	var payout, newYesPool, newNoPool float64
	if req.TradeType == "YES" {
		newYesPool = yesPool + req.Shares
		newNoPool = k / newYesPool
		payout = noPool - newNoPool
	} else {
		newNoPool = noPool + req.Shares
		newYesPool = k / newNoPool
		payout = yesPool - newYesPool
	}

	payout = math.Round(payout*1e6) / 1e6

	if payout <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "sell too small to produce payout"})
	}

	// batch write
	batch := db.Client.Batch()
	batch.Update(marketRef, []firestore.Update{
		{Path: "yesPool", Value: newYesPool},
		{Path: "noPool", Value: newNoPool},
	})
	batch.Update(userRef, []firestore.Update{
		{Path: "balance", Value: user.Balance + payout},
	})
	batch.Set(positionRef, map[string]interface{}{
		positionField:   firestore.Increment(-req.Shares),
		"competitionId": req.CompetitionID,
		"teamName":      market.TeamName,
	}, firestore.MergeAll)
	batch.Set(
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
	)

	if _, err := batch.Commit(ctx); err != nil {
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

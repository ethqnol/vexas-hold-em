package handlers

import (
	"context"
	"crypto/rand"
	"math/big"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gofiber/fiber/v2"
	"github.com/vexas-hold-em/backend/internal/db"
	"github.com/vexas-hold-em/backend/internal/models"
)

const maxBetAmount = 10000000.0
const casinoTimeout = 10 * time.Second

var rouletteNumbers = []int{0, 32, 15, 19, 4, 21, 2, 25, 17, 34, 6, 27, 13, 36, 11, 30, 8, 23, 10, 5, 24, 16, 33, 1, 20, 14, 31, 9, 22, 18, 29, 7, 28, 12, 35, 3, 26}

func getRouletteColor(num int) string {
	if num == 0 {
		return "green"
	}
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
	if req.Amount > maxBetAmount {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Bet amount exceeds maximum limit of 10 million VEX"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), casinoTimeout)
	defer cancel()

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

		bg, _ := rand.Int(rand.Reader, big.NewInt(int64(len(rouletteNumbers))))
		resultNum := rouletteNumbers[bg.Int64()]
		resultColor := getRouletteColor(resultNum)

		payout := calculateRoulettePayout(req.BetType, req.Amount, resultNum, resultColor)
		newBalance := user.Balance - req.Amount + payout

		updates := []firestore.Update{
			{Path: "balance", Value: newBalance},
		}
		if payout == 0 {
			updates = append(updates, firestore.Update{Path: "totalLost", Value: firestore.Increment(req.Amount)})
		}
		if err := tx.Update(userRef, updates); err != nil {
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

// Weighted reel - common symbols more frequent, jackpots rare
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
	if req.Amount > maxBetAmount {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Bet amount exceeds maximum limit of 10 million VEX"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), casinoTimeout)
	defer cancel()

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

		bg1, _ := rand.Int(rand.Reader, big.NewInt(int64(len(reelStrip))))
		bg2, _ := rand.Int(rand.Reader, big.NewInt(int64(len(reelStrip))))
		bg3, _ := rand.Int(rand.Reader, big.NewInt(int64(len(reelStrip))))
		reels := []string{
			reelStrip[bg1.Int64()],
			reelStrip[bg2.Int64()],
			reelStrip[bg3.Int64()],
		}

		payout := 0.0
		if reels[0] == reels[1] && reels[1] == reels[2] {
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
			payout = req.Amount * 2
		}

		newBalance := user.Balance - req.Amount + payout
		updates := []firestore.Update{
			{Path: "balance", Value: newBalance},
		}
		if payout == 0 {
			updates = append(updates, firestore.Update{Path: "totalLost", Value: firestore.Increment(req.Amount)})
		}
		if err := tx.Update(userRef, updates); err != nil {
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

func randomCard() string {
	ranks := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}
	suits := []string{"♠", "♥", "♦", "♣"}
	rIdx, _ := rand.Int(rand.Reader, big.NewInt(13))
	sIdx, _ := rand.Int(rand.Reader, big.NewInt(4))
	return ranks[rIdx.Int64()] + suits[sIdx.Int64()]
}

func cardValue(card string) int {
	rank := strings.TrimRightFunc(card, func(r rune) bool {
		return r == '♠' || r == '♥' || r == '♦' || r == '♣'
	})
	switch rank {
	case "A":
		return 11
	case "J", "Q", "K":
		return 10
	default:
		v, _ := strconv.Atoi(rank)
		return v
	}
}

func handTotal(cards []string) int {
	total, aces := 0, 0
	for _, c := range cards {
		v := cardValue(c)
		total += v
		if v == 11 {
			aces++
		}
	}
	for total > 21 && aces > 0 {
		total -= 10
		aces--
	}
	return total
}

func PlayBlackjackDeal(c *fiber.Ctx) error {
	var req struct {
		UserID    string  `json:"userId"`
		BetAmount float64 `json:"betAmount"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.BetAmount <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "bet must be > 0"})
	}
	if req.BetAmount > maxBetAmount {
		return c.Status(400).JSON(fiber.Map{"error": "Bet amount exceeds maximum limit of 10 million VEX"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), casinoTimeout)
	defer cancel()

	userRef := db.Client.Collection("users").Doc(req.UserID)
	sessionRef := userRef.Collection("blackjack_session").Doc("active")

	// Forfeit any active session
	sessDoc, err := sessionRef.Get(ctx)
	oldBet := 0.0
	if err == nil && sessDoc.Exists() {
		var oldSess models.BlackjackSession
		if err := sessDoc.DataTo(&oldSess); err == nil {
			oldBet = oldSess.BetAmount
		}
	}

	err = db.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(userRef)
		if err != nil {
			return err
		}
		var user models.User
		if err := doc.DataTo(&user); err != nil {
			return err
		}

		if user.Balance < req.BetAmount {
			return fiber.NewError(400, "insufficient balance")
		}

		updates := []firestore.Update{
			{Path: "balance", Value: user.Balance - req.BetAmount},
		}
		if oldBet > 0 {
			updates = append(updates, firestore.Update{Path: "totalLost", Value: firestore.Increment(oldBet)})
		}

		return tx.Update(userRef, updates)
	})

	if err != nil {
		if e, ok := err.(*fiber.Error); ok {
			return c.Status(e.Code).JSON(fiber.Map{"error": e.Message})
		}
		return c.Status(500).JSON(fiber.Map{"error": "failed to process transaction"})
	}

	sessionRef.Delete(ctx)

	playerCards := []string{randomCard(), randomCard()}
	dealerCards := []string{randomCard(), randomCard()}
	playerTotal := handTotal(playerCards)
	dealerTotal := handTotal(dealerCards)

	// Check for natural blackjack
	if playerTotal == 21 || dealerTotal == 21 {
		var result string
		var payout, lostAmount float64

		if playerTotal == 21 && dealerTotal == 21 {
			result = "push"
			payout = req.BetAmount
			lostAmount = 0
		} else if playerTotal == 21 {
			result = "blackjack"
			payout = req.BetAmount * 2.5
			lostAmount = 0
		} else {
			result = "lose"
			payout = 0
			lostAmount = req.BetAmount
		}

		db.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
			doc, _ := tx.Get(userRef)
			var user models.User
			doc.DataTo(&user)
			updates := []firestore.Update{
				{Path: "balance", Value: user.Balance + payout},
			}
			if lostAmount > 0 {
				updates = append(updates, firestore.Update{Path: "totalLost", Value: firestore.Increment(lostAmount)})
			}
			return tx.Update(userRef, updates)
		})

		return c.JSON(fiber.Map{
			"result":      result,
			"playerCards": playerCards,
			"dealerCards": dealerCards,
			"playerTotal": playerTotal,
			"dealerTotal": dealerTotal,
			"payout":      payout,
			"status":      "done",
		})
	}

	sessionRef.Set(ctx, models.BlackjackSession{
		BetAmount:   req.BetAmount,
		PlayerCards: playerCards,
		DealerCards: dealerCards,
		Status:      "playing",
	})

	return c.JSON(fiber.Map{
		"playerCards":       playerCards,
		"dealerVisibleCard": dealerCards[0],
		"playerTotal":       playerTotal,
		"dealerVisibleTotal": cardValue(dealerCards[0]),
		"status":            "playing",
	})
}

func PlayBlackjackAction(c *fiber.Ctx) error {
	var req struct {
		UserID string `json:"userId"`
		Action string `json:"action"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), casinoTimeout)
	defer cancel()

	userRef := db.Client.Collection("users").Doc(req.UserID)
	sessionRef := userRef.Collection("blackjack_session").Doc("active")

	sessDoc, err := sessionRef.Get(ctx)
	if err != nil || !sessDoc.Exists() {
		return c.Status(400).JSON(fiber.Map{"error": "no active session"})
	}
	var session models.BlackjackSession
	if err := sessDoc.DataTo(&session); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to parse session"})
	}

	bet := session.BetAmount

	if req.Action == "hit" {
		session.PlayerCards = append(session.PlayerCards, randomCard())
		playerTotal := handTotal(session.PlayerCards)

		if playerTotal > 21 {
			db.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
				doc, _ := tx.Get(userRef)
				var user models.User
				doc.DataTo(&user)
				return tx.Update(userRef, []firestore.Update{
					{Path: "totalLost", Value: firestore.Increment(bet)},
				})
			})
			sessionRef.Delete(ctx)
			return c.JSON(fiber.Map{
				"result":      "bust",
				"playerCards": session.PlayerCards,
				"dealerCards": session.DealerCards,
				"playerTotal": playerTotal,
				"dealerTotal": handTotal(session.DealerCards),
				"payout":      0,
				"status":      "done",
			})
		}

		sessionRef.Set(ctx, session)
		return c.JSON(fiber.Map{
			"playerCards":       session.PlayerCards,
			"dealerVisibleCard": session.DealerCards[0],
			"playerTotal":       playerTotal,
			"dealerVisibleTotal": cardValue(session.DealerCards[0]),
			"status":            "playing",
		})
	}

	if req.Action == "double" {
		err := db.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
			doc, err := tx.Get(userRef)
			if err != nil {
				return err
			}
			var user models.User
			if err := doc.DataTo(&user); err != nil {
				return err
			}
			if user.Balance < bet {
				return fiber.NewError(400, "insufficient balance to double")
			}
			return tx.Update(userRef, []firestore.Update{
				{Path: "balance", Value: user.Balance - bet},
			})
		})
		if err != nil {
			if e, ok := err.(*fiber.Error); ok {
				return c.Status(e.Code).JSON(fiber.Map{"error": e.Message})
			}
			return c.Status(500).JSON(fiber.Map{"error": "failed to double"})
		}
		session.PlayerCards = append(session.PlayerCards, randomCard())
		bet = bet * 2
	}

	// Dealer draws to 17
	dealerTotal := handTotal(session.DealerCards)
	for dealerTotal < 17 {
		session.DealerCards = append(session.DealerCards, randomCard())
		dealerTotal = handTotal(session.DealerCards)
	}
	playerTotal := handTotal(session.PlayerCards)

	var result string
	var payout, lostAmount float64

	if playerTotal > 21 {
		result = "bust"
		lostAmount = bet
	} else if dealerTotal > 21 || playerTotal > dealerTotal {
		result = "win"
		payout = bet * 2
	} else if playerTotal == dealerTotal {
		result = "push"
		payout = bet
	} else {
		result = "lose"
		lostAmount = bet
	}

	db.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, _ := tx.Get(userRef)
		var user models.User
		doc.DataTo(&user)
		updates := []firestore.Update{
			{Path: "balance", Value: user.Balance + payout},
		}
		if lostAmount > 0 {
			updates = append(updates, firestore.Update{Path: "totalLost", Value: firestore.Increment(lostAmount)})
		}
		return tx.Update(userRef, updates)
	})

	sessionRef.Delete(ctx)
	return c.JSON(fiber.Map{
		"result":      result,
		"playerCards": session.PlayerCards,
		"dealerCards": session.DealerCards,
		"playerTotal": playerTotal,
		"dealerTotal": dealerTotal,
		"payout":      payout,
		"status":      "done",
	})
}

const lotteryTicketCost = 10.0
const lotteryOdds = 1000000

func GetLotteryStatus(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	doc, err := db.Client.Collection("meta").Doc("lottery").Get(ctx)
	if err != nil || !doc.Exists() {
		return c.JSON(models.LotteryConfig{Jackpot: 100.0})
	}
	var lottery models.LotteryConfig
	doc.DataTo(&lottery)
	return c.JSON(lottery)
}

func PlayLottery(c *fiber.Ctx) error {
	var req struct {
		UserID string `json:"userId"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), casinoTimeout)
	defer cancel()

	userRef := db.Client.Collection("users").Doc(req.UserID)
	lotteryRef := db.Client.Collection("meta").Doc("lottery")

	var won bool
	var jackpotSnapshot float64

	err := db.Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		userDoc, err := tx.Get(userRef)
		if err != nil {
			return err
		}
		var user models.User
		if err := userDoc.DataTo(&user); err != nil {
			return err
		}
		if user.Balance < lotteryTicketCost {
			return fiber.NewError(400, "insufficient balance — tickets cost 10 S.H.I.T.")
		}

		lotDoc, _ := tx.Get(lotteryRef)
		var lot models.LotteryConfig
		if lotDoc.Exists() {
			lotDoc.DataTo(&lot)
		} else {
			lot = models.LotteryConfig{Jackpot: 100.0}
		}
		jackpotSnapshot = lot.Jackpot

		draw, _ := rand.Int(rand.Reader, big.NewInt(lotteryOdds))
		won = draw.Int64() == 0

		if won {
			if err := tx.Update(userRef, []firestore.Update{
				{Path: "balance", Value: user.Balance - lotteryTicketCost + lot.Jackpot},
				{Path: "totalLost", Value: firestore.Increment(lotteryTicketCost)},
			}); err != nil {
				return err
			}
			return tx.Set(lotteryRef, models.LotteryConfig{
				Jackpot:        100.0,
				LastWinner:     req.UserID,
				LastWinnerName: user.DisplayName,
				LastWonAt:      time.Now().UnixMilli(),
			})
		}

		if err := tx.Update(userRef, []firestore.Update{
			{Path: "balance", Value: user.Balance - lotteryTicketCost},
			{Path: "totalLost", Value: firestore.Increment(lotteryTicketCost)},
		}); err != nil {
			return err
		}
		return tx.Update(lotteryRef, []firestore.Update{
			{Path: "jackpot", Value: firestore.Increment(lotteryTicketCost * 0.8)},
		})
	})

	if err != nil {
		if e, ok := err.(*fiber.Error); ok {
			return c.Status(e.Code).JSON(fiber.Map{"error": e.Message})
		}
		return c.Status(500).JSON(fiber.Map{"error": "lottery failed"})
	}

	if won {
		return c.JSON(fiber.Map{
			"result":  "WIN",
			"jackpot": jackpotSnapshot,
			"message": "🌊 THE WATER GAME IS REAL. YOU WON THE LOTTERY.",
		})
	}
	return c.JSON(fiber.Map{
		"result":  "lose",
		"jackpot": jackpotSnapshot,
		"message": "the water game never comes. the cycle continues.",
	})
}

type storeItemDef struct {
	Name  string
	Cost  float64
	Type  string
	Value string
}

var storeItems = map[string]storeItemDef{
	"theme_default":    {"Original Theme", 0, "cosmetic", ""},
	"theme_industrial": {"Windows XP Mode", 6767, "cosmetic", "theme_industrial"},
	"theme_gold":       {"Baller Mode (I Have Money)", 500000, "cosmetic", "theme_gold"},
	"theme_neon":       {"Seizure Mode", 750000, "cosmetic", "theme_neon"},
	"theme_vomit":      {"Vomit Green", 10000, "cosmetic", "theme_vomit"},
	"title_none":          {"No Title", 0, "title", ""},
	"title_degenerate":    {"Verified Degenerate", 2500, "title", "Verified Degenerate"},
	"title_water_truther": {"Water Game Truther", 50000, "title", "Water Game Truther"},
	"title_down_bad":      {"Down Bad", 420, "title", "Down Bad"},
	"title_financially_cooked": {"Financially Cooked", 150000, "title", "Financially Cooked"},
	"title_jason_victim":  {"jason's victim", 999999, "title", "jason's victim"},
	"title_pavilion":      {"Pavilion Regular", 75000, "title", "Pavilion Regular"},
}

func GetStoreItems(c *fiber.Ctx) error {
	items := []fiber.Map{}
	for id, item := range storeItems {
		items = append(items, fiber.Map{
			"id":    id,
			"name":  item.Name,
			"cost":  item.Cost,
			"type":  item.Type,
			"value": item.Value,
		})
	}
	return c.JSON(fiber.Map{"items": items, "status": "success"})
}

func BuyStoreItem(c *fiber.Ctx) error {
	var req struct {
		UserID string `json:"userId"`
		ItemID string `json:"itemId"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	item, ok := storeItems[req.ItemID]
	if !ok {
		return c.Status(400).JSON(fiber.Map{"error": "item not found"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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

		// already owned or free default → just equip, no charge
		isDefault := req.ItemID == "theme_default" || req.ItemID == "title_none"
		isAlreadyOwned := false
		for _, owned := range user.OwnedItems {
			if owned == req.ItemID {
				isAlreadyOwned = true
				break
			}
		}

		if isAlreadyOwned || isDefault {
			fieldPath := "cosmetic"
			if item.Type == "title" {
				fieldPath = "title"
			}
			return tx.Update(userRef, []firestore.Update{
				{Path: fieldPath, Value: item.Value},
			})
		}

		// new purchase
		if user.Balance < item.Cost {
			return fiber.NewError(400, "insufficient balance")
		}

		fieldPath := "cosmetic"
		if item.Type == "title" {
			fieldPath = "title"
		}
		return tx.Update(userRef, []firestore.Update{
			{Path: "balance", Value: user.Balance - item.Cost},
			{Path: "ownedItems", Value: firestore.ArrayUnion(req.ItemID)},
			{Path: fieldPath, Value: item.Value},
		})
	})

	if err != nil {
		if e, ok := err.(*fiber.Error); ok {
			return c.Status(e.Code).JSON(fiber.Map{"error": e.Message})
		}
		return c.Status(500).JSON(fiber.Map{"error": "purchase failed"})
	}

	return c.JSON(fiber.Map{"status": "success", "message": "item equipped!"})
}

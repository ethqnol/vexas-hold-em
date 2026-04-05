package handlers

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/vexas-hold-em/backend/internal/db"
	"github.com/vexas-hold-em/backend/internal/models"
)

// leaderboard cache
var (
	cachedLeaderboard     []fiber.Map
	leaderboardCacheTime  time.Time
	leaderboardCacheMutex sync.RWMutex
	leaderboardCacheTTL   = 5 * time.Minute
)

// gets user info & bal
func GetUserProfile(c *fiber.Ctx) error {
	id := c.Params("id")

	if db.Client == nil {
		log.Println("GetUserProfile: Database not initialized")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database not initialized"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	doc, err := db.Client.Collection("users").Doc(id).Get(ctx)
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

// creates new user doc if user doesnt exist
func SyncUser(c *fiber.Ctx) error {
	id := c.Params("id")

	if db.Client == nil {
		log.Println("SyncUser: Database not initialized")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database not initialized"})
	}

	// parse body for email/disp name (from fb auth)
	var req struct {
		Email       string `json:"email"`
		DisplayName string `json:"displayName"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	docRef := db.Client.Collection("users").Doc(id)
	doc, err := docRef.Get(ctx)

	if err != nil || !doc.Exists() {
		newUser := models.User{
			Balance:     1000.0,
			Email:       req.Email,
			DisplayName: req.DisplayName,
		}

		_, err := docRef.Set(ctx, newUser)
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

// gets active positions
func GetUserPortfolio(c *fiber.Ctx) error {
	id := c.Params("id")

	if db.Client == nil {
		log.Println("GetUserPortfolio: Database not initialized")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database not initialized"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	iter := db.Client.Collection("users").Doc(id).Collection("positions").Documents(ctx)
	docs, err := iter.GetAll()
	if err != nil {
		log.Printf("GetUserPortfolio: Failed to fetch portfolio: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch portfolio"})
	}

	portfolio := make(map[string]map[string]models.Position)
	for _, doc := range docs {
		var pos models.Position
		if err := doc.DataTo(&pos); err == nil {
			compID := pos.CompetitionID
			if compID == "" {
				compID = "unknown_competition"
			}
			if portfolio[compID] == nil {
				portfolio[compID] = make(map[string]models.Position)
			}
			portfolio[compID][doc.Ref.ID] = pos
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

// top 50 users by current balance
func GetLeaderboard(c *fiber.Ctx) error {
	if db.Client == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "db not initialized"})
	}

	// check cache first
	leaderboardCacheMutex.RLock()
	if time.Since(leaderboardCacheTime) < leaderboardCacheTTL && cachedLeaderboard != nil {
		cached := cachedLeaderboard
		leaderboardCacheMutex.RUnlock()
		return c.JSON(fiber.Map{
			"leaderboard": cached,
			"status":      "success",
			"cached":      true,
		})
	}
	leaderboardCacheMutex.RUnlock()

	// cache miss - fetch from firestore
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	docs, err := db.Client.Collection("users").OrderBy("balance", 1).Limit(50).Documents(ctx).GetAll()
	if err != nil {
		log.Printf("GetLeaderboard: failed to fetch: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch leaderboard"})
	}

	// reverse for descending (Firestore Desc constant = 1 in the go SDK)
	for i, j := 0, len(docs)-1; i < j; i, j = i+1, j-1 {
		docs[i], docs[j] = docs[j], docs[i]
	}

	leaderboard := []fiber.Map{}
	for i, doc := range docs {
		var user models.User
		if err := doc.DataTo(&user); err != nil {
			continue
		}
		leaderboard = append(leaderboard, fiber.Map{
			"rank":        i + 1,
			"userId":      doc.Ref.ID,
			"displayName": user.DisplayName,
			"balance":     user.Balance,
			"totalLost":   user.TotalLost,
			"title":       user.Title,
		})
	}

	// update cache
	leaderboardCacheMutex.Lock()
	cachedLeaderboard = leaderboard
	leaderboardCacheTime = time.Now()
	leaderboardCacheMutex.Unlock()

	return c.JSON(fiber.Map{
		"leaderboard": leaderboard,
		"status":      "success",
		"cached":      false,
	})
}


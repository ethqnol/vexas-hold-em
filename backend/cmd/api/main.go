package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/vexas-hold-em/backend/internal/db"
	"github.com/vexas-hold-em/backend/internal/handlers"
	"github.com/vexas-hold-em/backend/internal/middleware"
)

func main() {
	app := fiber.New()

	db.InitFirestore()
	defer db.CloseFirestore()

	app.Use(logger.New())

	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded. Please try again later.",
			})
		},
	}))

	frontendOrigin := os.Getenv("FRONTEND_ORIGIN")
	if frontendOrigin == "" {
		frontendOrigin = "http://localhost:5173,https://vexasholdem.web.app"
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins: frontendOrigin,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	api := app.Group("/api/v1")

	api.Get("/users/:id", handlers.GetUserProfile)
	api.Post("/users/:id/sync", handlers.SyncUser)
	api.Get("/users/:id/portfolio", handlers.GetUserPortfolio)

	api.Get("/competitions", handlers.GetCompetitions)
	api.Get("/competitions/:id", handlers.GetCompetitionByID)
	api.Get("/competitions/:id/markets", handlers.GetMarketsByCompetition)
	api.Get("/competitions/:id/history", handlers.GetCompetitionHistory)
	api.Get("/competitions/:id/markets/:marketId/history", handlers.GetMarketHistory)

	api.Post("/trade", handlers.PlaceTrade)
	api.Post("/trade/sell", handlers.SellShares)

	api.Post("/casino/roulette", handlers.PlayRoulette)
	api.Post("/casino/slots", handlers.PlaySlots)
	api.Post("/casino/blackjack/deal", handlers.PlayBlackjackDeal)
	api.Post("/casino/blackjack/action", handlers.PlayBlackjackAction)
	api.Get("/casino/lottery", handlers.GetLotteryStatus)
	api.Post("/casino/lottery", handlers.PlayLottery)
	api.Get("/casino/store", handlers.GetStoreItems)
	api.Post("/casino/store/buy", handlers.BuyStoreItem)

	api.Get("/leaderboard", handlers.GetLeaderboard)

	admin := api.Group("/admin")
	admin.Use(middleware.AdminOnly)
	admin.Post("/competitions", handlers.CreateCompetition)
	admin.Put("/competitions/:id/status", handlers.UpdateCompetitionStatus)
	admin.Post("/competitions/:id/resolve", handlers.ResolveCompetition)
	admin.Post("/competitions/:id/reset", handlers.ResetCompetition)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s...\n", port)
	log.Fatal(app.Listen(":" + port))
}

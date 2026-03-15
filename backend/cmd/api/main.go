package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/vexas-hold-em/backend/internal/db"
	"github.com/vexas-hold-em/backend/internal/handlers"
)

func main() {
	app := fiber.New()

	db.InitFirestore()
	defer db.CloseFirestore()

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	api := app.Group("/api/v1")

	// users
	api.Get("/users/:id", handlers.GetUserProfile)
	api.Post("/users/:id/sync", handlers.SyncUser)
	api.Get("/users/:id/portfolio", handlers.GetUserPortfolio)

	// comps + markets
	api.Get("/competitions", handlers.GetCompetitions)
	api.Get("/competitions/:id", handlers.GetCompetitionByID)
	api.Get("/competitions/:id/markets", handlers.GetMarketsByCompetition)
	api.Get("/competitions/:id/history", handlers.GetCompetitionHistory)
	api.Get("/competitions/:id/markets/:marketId/history", handlers.GetMarketHistory)

	// trading
	api.Post("/trade", handlers.PlaceTrade)
	api.Post("/trade/sell", handlers.SellShares)

	// casino
	api.Post("/casino/roulette", handlers.PlayRoulette)
	api.Post("/casino/slots", handlers.PlaySlots)

	// admin (unsecured for now)
	admin := api.Group("/admin")
	admin.Post("/competitions", handlers.CreateCompetition)
	admin.Put("/competitions/:id/status", handlers.UpdateCompetitionStatus)
	admin.Post("/competitions/:id/resolve", handlers.ResolveCompetition)
	admin.Post("/competitions/:id/reset", handlers.ResetCompetition)

	log.Println("Starting server on port 8080...")
	log.Fatal(app.Listen(":8080"))
}

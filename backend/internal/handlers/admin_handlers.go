package handlers

import (
	"context"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/extrame/xls"
	"github.com/gofiber/fiber/v2"
	"github.com/vexas-hold-em/backend/internal/db"
	"github.com/vexas-hold-em/backend/internal/models"
)

const adminTimeout = 30 * time.Second

func CreateCompetition(c *fiber.Ctx) error {
	if db.Client == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database not initialized"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), adminTimeout)
	defer cancel()

	compName := c.FormValue("name")
	if compName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Competition name is required"})
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "XLSX file is required"})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to open file"})
	}
	defer file.Close()

	xlFile, err := xls.OpenReader(file, "utf-8")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid XLS file"})
	}

	sheet := xlFile.GetSheet(0)
	if sheet == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "XLS file is empty"})
	}

	compDoc := db.Client.Collection("competitions").Doc(compName)
	_, err = compDoc.Set(ctx, models.Competition{
		Status: "active",
	})
	if err != nil {
		log.Printf("CreateCompetition: Failed to create competition: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create competition"})
	}

	marketsRef := compDoc.Collection("markets")
	successCount := 0
	maxRow := int(sheet.MaxRow)

	colMap := map[string]int{
		"Team":         -1,
		"Team Name":    -1,
		"Division":     -1,
		"Organization": -1,
		"Location":     -1,
	}

	for i := 0; i <= maxRow; i++ {
		row := sheet.Row(i)
		if row == nil {
			continue
		}

		if i == 0 || (colMap["Team"] == -1 && row.Col(0) == "Team") {
			for j := 0; j < 10; j++ {
				colName := row.Col(j)
				if colName == "Team" || colName == "Team Number" {
					colMap["Team"] = j
				} else if colName == "Team Name" {
					colMap["Team Name"] = j
				} else if colName == "Division" || colName == "Division Name" {
					colMap["Division"] = j
				} else if colName == "Organization" || colName == "School" {
					colMap["Organization"] = j
				} else if colName == "Location" || colName == "City" || colName == "State" {
					if colMap["Location"] == -1 {
						colMap["Location"] = j
					}
				}
			}
			continue
		}

		if colMap["Team"] == -1 {
			continue
		}

		teamID := row.Col(colMap["Team"])
		if teamID == "" {
			continue
		}

		teamName := ""
		if colMap["Team Name"] != -1 {
			teamName = row.Col(colMap["Team Name"])
		}

		division := "Default"
		if colMap["Division"] != -1 {
			division = row.Col(colMap["Division"])
		}

		organization := ""
		if colMap["Organization"] != -1 {
			organization = row.Col(colMap["Organization"])
		}

		location := ""
		if colMap["Location"] != -1 {
			location = row.Col(colMap["Location"])
		}

		market := models.Market{
			TeamID:       teamID,
			TeamName:     teamName,
			Division:     division,
			Organization: organization,
			Location:     location,
			YesPool:      100.0,
			NoPool:       100.0,
		}

		_, err := marketsRef.Doc(teamID).Set(ctx, market)
		if err == nil {
			successCount++
		} else {
			log.Printf("CreateCompetition: Failed to create market for %s: %v", teamID, err)
		}
	}

	return c.JSON(fiber.Map{
		"message":        "Competition created successfully",
		"status":         "success",
		"teams_imported": successCount,
	})
}

func UpdateCompetitionStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	if db.Client == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "db not initialized"})
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON"})
	}

	if req.Status != "active" && req.Status != "paused" && req.Status != "closed" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid status"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.Client.Collection("competitions").Doc(id).Update(ctx, []firestore.Update{
		{Path: "status", Value: req.Status},
	})
	if err != nil {
		log.Printf("Failed to update status: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update status"})
	}

	return c.JSON(fiber.Map{
		"message": "Competition " + id + " status updated to " + req.Status,
		"status":  "success",
	})
}

type ResolveRequest struct {
	WinningTeamIDs []string `json:"winningTeamIds"`
}

func ResolveCompetition(c *fiber.Ctx) error {
	id := c.Params("id")
	if db.Client == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "db not initialized"})
	}

	var req ResolveRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
	}

	if len(req.WinningTeamIDs) < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "must provide at least 1 winning team"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), adminTimeout)
	defer cancel()

	compRef := db.Client.Collection("competitions").Doc(id)
	_, err := compRef.Update(ctx, []firestore.Update{
		{Path: "status", Value: "resolved"},
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update competition status"})
	}

	winners := make(map[string]bool)
	for _, w := range req.WinningTeamIDs {
		winners[w] = true
	}

	usersRef := db.Client.Collection("users")
	users, err := usersRef.Documents(ctx).GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch users"})
	}

	batch := db.Client.Batch()
	opsCount := 0

	commitIfFull := func() error {
		if opsCount >= 490 {
			if _, err := batch.Commit(ctx); err != nil {
				return err
			}
			batch = db.Client.Batch()
			opsCount = 0
		}
		return nil
	}

	totalPayouts := 0.0

	for _, userDoc := range users {
		userID := userDoc.Ref.ID
		positions, err := usersRef.Doc(userID).Collection("positions").Where("competitionId", "==", id).Documents(ctx).GetAll()
		if err != nil {
			continue
		}

		userWinnings := 0.0

		for _, posDoc := range positions {
			teamID := posDoc.Ref.ID

			var pos models.Position
			if err := posDoc.DataTo(&pos); err != nil {
				continue
			}

			isWinner := winners[teamID]

			if isWinner && pos.YesShares > 0 {
				userWinnings += pos.YesShares
			} else if !isWinner && pos.NoShares > 0 {
				userWinnings += pos.NoShares
			}

			batch.Delete(posDoc.Ref)
			opsCount++
			if err := commitIfFull(); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to commit position deletion"})
			}
		}

		if userWinnings > 0 {
			batch.Update(userDoc.Ref, []firestore.Update{
				{Path: "balance", Value: firestore.Increment(userWinnings)},
			})
			opsCount++
			totalPayouts += userWinnings

			if err := commitIfFull(); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to commit user payout"})
			}
		}
	}

	if opsCount > 0 {
		if _, err := batch.Commit(ctx); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "final commit failed for payouts"})
		}
	}

	return c.JSON(fiber.Map{
		"message":                 "Competition " + id + " resolved and payouts distributed",
		"status":                  "success",
		"totalPayoutsDistributed": totalPayouts,
	})
}

func ResetCompetition(c *fiber.Ctx) error {
	id := c.Params("id")
	if db.Client == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "db not initialized"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), adminTimeout)
	defer cancel()

	ledgerRef := db.Client.Collection("competitions").Doc(id).Collection("ledger")
	txns, err := ledgerRef.Documents(ctx).GetAll()
	if err != nil {
		log.Printf("ResetCompetition: failed to fetch ledger: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch ledger"})
	}

	userRefunds := make(map[string]float64)
	userPositions := make(map[string]map[string]bool)
	for _, doc := range txns {
		var tx models.Transaction
		if err := doc.DataTo(&tx); err == nil {
			userRefunds[tx.UserID] += tx.AmountSpent

			if userPositions[tx.UserID] == nil {
				userPositions[tx.UserID] = make(map[string]bool)
			}
			userPositions[tx.UserID][tx.TeamID] = true
		}
	}

	batch := db.Client.Batch()
	opsCount := 0

	commitIfFull := func() error {
		if opsCount >= 490 {
			if _, err := batch.Commit(ctx); err != nil {
				return err
			}
			batch = db.Client.Batch()
			opsCount = 0
		}
		return nil
	}

	for userID, refundAmount := range userRefunds {
		if refundAmount != 0 {
			batch.Update(db.Client.Collection("users").Doc(userID), []firestore.Update{
				{Path: "balance", Value: firestore.Increment(refundAmount)},
			})
			opsCount++
			if err := commitIfFull(); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to commit user refunds"})
			}
		}
	}

	for _, doc := range txns {
		batch.Delete(doc.Ref)
		opsCount++
		if err := commitIfFull(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to clear ledger"})
		}
	}

	marketsRef := db.Client.Collection("competitions").Doc(id).Collection("markets")
	markets, err := marketsRef.Documents(ctx).GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch markets"})
	}

	for _, doc := range markets {
		batch.Update(doc.Ref, []firestore.Update{
			{Path: "yesPool", Value: 100.0},
			{Path: "noPool", Value: 100.0},
		})
		opsCount++
		if err := commitIfFull(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to reset markets"})
		}
	}

	for userID, marketsMap := range userPositions {
		for teamID := range marketsMap {
			posRef := db.Client.Collection("users").Doc(userID).Collection("positions").Doc(teamID)
			batch.Delete(posRef)
			opsCount++
			if err := commitIfFull(); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to wipe positions"})
			}
		}
	}

	compRef := db.Client.Collection("competitions").Doc(id)
	batch.Update(compRef, []firestore.Update{
		{Path: "status", Value: "active"},
		{Path: "winningTeamId", Value: firestore.Delete},
	})
	opsCount++

	if opsCount > 0 {
		if _, err := batch.Commit(ctx); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "final commit failed: " + err.Error()})
		}
	}

	return c.JSON(fiber.Map{
		"message": "Competition " + id + " has been completely reset and all VEX refunded",
		"status":  "success",
	})
}

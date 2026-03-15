package handlers

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	"github.com/extrame/xls"
	"github.com/gofiber/fiber/v2"
	"github.com/vexas-hold-em/backend/internal/db"
	"github.com/vexas-hold-em/backend/internal/models"
)

// init a new VEX competition
func CreateCompetition(c *fiber.Ctx) error {
	if db.Client == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database not initialized"})
	}

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

	// parse XLS
	xlFile, err := xls.OpenReader(file, "utf-8")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid XLS file"})
	}

	sheet := xlFile.GetSheet(0)
	if sheet == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "XLS file is empty"})
	}

	// create comp doc
	compDoc := db.Client.Collection("competitions").Doc(compName)
	_, err = compDoc.Set(context.Background(), models.Competition{
		Status: "active",
	})
	if err != nil {
		log.Printf("CreateCompetition: Failed to create competition: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create competition"})
	}

	// process teams -> markets
	marketsRef := compDoc.Collection("markets")
	successCount := 0

	maxRow := int(sheet.MaxRow)

	// dynamically map col indices from header row
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

		// resolve header cols
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
					// grab first location-like col
					if colMap["Location"] == -1 {
						colMap["Location"] = j
					}
				}
			}
			continue
		}

		// skip if team col not found yet
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
			YesPool:      100.0, // base liquidity
			NoPool:       100.0, // base liquidity
		}

		_, err := marketsRef.Doc(teamID).Set(context.Background(), market)
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

// kill switch — pause or close a comp
func UpdateCompetitionStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	// todo: parse requested status ('paused', 'closed', 'active')
	// todo: update status in firestore
	return c.JSON(fiber.Map{
		"message": "Competition " + id + " status updated",
		"status":  "success",
	})
}

// body for payout
type ResolveRequest struct {
	WinningTeamIDs []string `json:"winningTeamIds"` // array of exactly 2 team IDs
}

// finalize event + trigger payout engine
func ResolveCompetition(c *fiber.Ctx) error {
	id := c.Params("id")
	if db.Client == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "db not initialized"})
	}

	var req ResolveRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
	}

	if len(req.WinningTeamIDs) != 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "must provide exactly 2 winning teams"})
	}

	ctx := context.Background()

	// 1. Mark comp as resolved
	compRef := db.Client.Collection("competitions").Doc(id)
	_, err := compRef.Update(ctx, []firestore.Update{
		{Path: "status", Value: "resolved"},
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update competition status"})
	}

	// create a map of winners for fast lookup
	winners := map[string]bool{
		req.WinningTeamIDs[0]: true,
		req.WinningTeamIDs[1]: true,
	}

	// 2. Fetch all users
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

	// 3. for each user, fetch their positions in this comp's mkts and payout
	for _, userDoc := range users {
		userID := userDoc.Ref.ID
		positions, err := usersRef.Doc(userID).Collection("positions").Documents(ctx).GetAll()
		if err != nil {
			continue // skip user if error fetching positions
		}

		userWinnings := 0.0

		for _, posDoc := range positions {
			teamID := posDoc.Ref.ID

			// we only care about positions that are part of this comp's mkts
			// since positions don't currently tag the comp id, we must verify the mkt exists in this comp
			marketRef := compRef.Collection("markets").Doc(teamID)
			marketSnap, err := marketRef.Get(ctx)
			if err != nil || !marketSnap.Exists() {
				// either error or this position belongs to a market from a DIFFERENT competition
				continue
			}

			var pos models.Position
			if err := posDoc.DataTo(&pos); err != nil {
				continue
			}

			// payout math
			// if team is a winner: yes shares pay $1, no shares pay $0
			// if team is a loser: yes shares pay $0, no shares pay $1
			isWinner := winners[teamID]

			if isWinner && pos.YesShares > 0 {
				userWinnings += pos.YesShares
			} else if !isWinner && pos.NoShares > 0 {
				userWinnings += pos.NoShares
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

// clear comp + refund users
func ResetCompetition(c *fiber.Ctx) error {
	id := c.Params("id")
	if db.Client == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "db not initialized"})
	}

	ctx := context.Background()

	// 1. fetch all ledger txns
	ledgerRef := db.Client.Collection("competitions").Doc(id).Collection("ledger")
	txns, err := ledgerRef.Documents(ctx).GetAll()
	if err != nil {
		log.Printf("ResetCompetition: failed to fetch ledger: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch ledger"})
	}

	// 2. calculate net amount spent per user (amountSpent tracks both buys (+) and payouts (-))
	userRefunds := make(map[string]float64)
	for _, doc := range txns {
		var tx models.Transaction
		if err := doc.DataTo(&tx); err == nil {
			userRefunds[tx.UserID] += tx.AmountSpent
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

	// 3. refund users
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

	// 4. delete ledger docs
	for _, doc := range txns {
		batch.Delete(doc.Ref)
		opsCount++
		if err := commitIfFull(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to clear ledger"})
		}
	}

	// 5. reset market pools
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

		// 6. delete user positions for this market
		// (this is heavy, ideally done via Cloud Functions, but doing it in-line for now)
		for userID := range userRefunds {
			posRef := db.Client.Collection("users").Doc(userID).Collection("positions").Doc(doc.Ref.ID)
			batch.Delete(posRef)
			opsCount++
			if err := commitIfFull(); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to wipe positions"})
			}
		}
	}

	// 7. reset comp status
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

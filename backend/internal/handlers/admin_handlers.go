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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	// Parse all markets from the sheet before writing anything to Firestore.
	colMap := map[string]int{
		"Team":         -1,
		"Team Name":    -1,
		"Division":     -1,
		"Organization": -1,
		"Location":     -1,
	}

	var markets []models.Market
	maxRow := int(sheet.MaxRow)

	for i := 0; i <= maxRow; i++ {
		row := sheet.Row(i)
		if row == nil {
			continue
		}

		if i == 0 {
			for j := 0; j <= row.LastCol(); j++ {
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

		markets = append(markets, models.Market{
			TeamID:       teamID,
			TeamName:     teamName,
			Division:     division,
			Organization: organization,
			Location:     location,
		})
	}

	if len(markets) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No teams found in XLS file"})
	}

	// Initial YES probability = 2/numTeams so that the sum of all YES prices = 2
	// (a reasonable prior for a single-winner competition where top teams are favoured).
	// Pool sum = 1000; yesPool = 1000 * p, noPool = 1000 * (1-p).
	numTeams := len(markets)
	yesFrac := 2.0 / float64(numTeams)
	if yesFrac > 0.999 {
		yesFrac = 0.999
	}
	initialYesPool := 500.0 * yesFrac
	initialNoPool := 500.0 - initialYesPool
	initialYesOdds := initialYesPool / (initialYesPool + initialNoPool)
	for i := range markets {
		markets[i].YesPool = initialYesPool
		markets[i].NoPool = initialNoPool
		markets[i].InitialYesOdds = initialYesOdds
	}

	// Create the competition doc; returns an error if it already exists.
	compDoc := db.Client.Collection("competitions").Doc(compName)
	_, err = compDoc.Create(ctx, models.Competition{
		Status: "active",
	})
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Competition already exists"})
		}
		log.Printf("CreateCompetition: Failed to create competition: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create competition"})
	}

	marketsRef := compDoc.Collection("markets")
	successCount := 0

	for _, market := range markets {
		_, err := marketsRef.Doc(market.TeamID).Set(ctx, market)
		if err == nil {
			successCount++
		} else {
			log.Printf("CreateCompetition: Failed to create market for %s: %v", market.TeamID, err)
		}
	}

	if successCount == 0 {
		if _, delErr := compDoc.Delete(ctx); delErr != nil {
			log.Printf("CreateCompetition: Failed to delete orphaned competition %s: %v", compName, delErr)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to import any teams"})
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

	winners := make(map[string]bool)
	for _, w := range req.WinningTeamIDs {
		winners[w] = true
	}

	// Single collection group query instead of one query per user.
	// Requires a Firestore collection group index on competitionId.
	positionDocs, err := db.Client.CollectionGroup("positions").Where("competitionId", "==", id).Documents(ctx).GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch positions"})
	}

	// Group winning shares by market so we can look up each market's reserve.
	type marketWin struct {
		totalShares float64
		userShares  map[string]float64
		userRefs    map[string]*firestore.DocumentRef
	}
	mktWins := make(map[string]*marketWin)

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

	for _, posDoc := range positionDocs {
		userRef := posDoc.Ref.Parent.Parent
		userID := userRef.ID
		teamID := posDoc.Ref.ID

		var pos models.Position
		if err := posDoc.DataTo(&pos); err != nil {
			batch.Delete(posDoc.Ref)
			opsCount++
			if err := commitIfFull(); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to commit position deletion"})
			}
			continue
		}

		isWinner := winners[teamID]
		var winShares float64
		if isWinner {
			winShares = pos.YesShares
		} else {
			winShares = pos.NoShares
		}

		if winShares > 0 {
			if mktWins[teamID] == nil {
				mktWins[teamID] = &marketWin{
					userShares: make(map[string]float64),
					userRefs:   make(map[string]*firestore.DocumentRef),
				}
			}
			mktWins[teamID].totalShares += winShares
			mktWins[teamID].userShares[userID] += winShares
			mktWins[teamID].userRefs[userID] = userRef
		}

		batch.Delete(posDoc.Ref)
		opsCount++
		if err := commitIfFull(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to commit position deletion"})
		}
	}

	// For each winning market, fetch its reserve and compute the per-share payout.
	// Each share pays $1 from the reserve; the reserve is always sufficient because
	// every dollar bet was deposited into it, and CPMM slippage ensures the reserve
	// accumulates at least as many dollars as shares were issued.
	userWinnings := make(map[string]float64)
	userRefs := make(map[string]*firestore.DocumentRef)

	for teamID, mw := range mktWins {
		mktRef := db.Client.Collection("competitions").Doc(id).Collection("markets").Doc(teamID)
		mktDoc, err := mktRef.Get(ctx)
		if err != nil {
			log.Printf("ResolveCompetition: failed to fetch market %s: %v", teamID, err)
			continue
		}
		var market models.Market
		if err := mktDoc.DataTo(&market); err != nil {
			continue
		}

		payoutPerShare := market.Reserve / mw.totalShares

		marketPayout := mw.totalShares * payoutPerShare
		batch.Update(mktRef, []firestore.Update{
			{Path: "reserve", Value: firestore.Increment(-marketPayout)},
		})
		opsCount++
		if err := commitIfFull(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to deduct from market reserve"})
		}

		for userID, shares := range mw.userShares {
			userWinnings[userID] += shares * payoutPerShare
			userRefs[userID] = mw.userRefs[userID]
		}
	}

	totalPayouts := 0.0
	for userID, winnings := range userWinnings {
		if winnings > 0 {
			batch.Update(userRefs[userID], []firestore.Update{
				{Path: "balance", Value: firestore.Increment(winnings)},
			})
			opsCount++
			totalPayouts += winnings
			if err := commitIfFull(); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to commit user payout"})
			}
		}
	}

	// Update competition status last — only marked resolved once all payouts are committed.
	compRef := db.Client.Collection("competitions").Doc(id)
	batch.Update(compRef, []firestore.Update{
		{Path: "status", Value: "resolved"},
		{Path: "winningTeamIds", Value: req.WinningTeamIDs},
	})
	opsCount++

	if _, err := batch.Commit(ctx); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "final commit failed for payouts"})
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

	// Use collection group query to get all positions for this competition directly,
	// rather than reconstructing them from the ledger (which would miss orphaned docs).
	positionDocs, err := db.Client.CollectionGroup("positions").Where("competitionId", "==", id).Documents(ctx).GetAll()
	if err != nil {
		log.Printf("ResetCompetition: failed to fetch positions: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch positions"})
	}

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

	numTeams := len(markets)
	yesFrac := 2.0 / float64(numTeams)
	if yesFrac > 0.999 {
		yesFrac = 0.999
	}
	resetYesPool := 500.0 * yesFrac
	resetNoPool := 500.0 - resetYesPool

	for _, doc := range markets {
		batch.Update(doc.Ref, []firestore.Update{
			{Path: "yesPool", Value: resetYesPool},
			{Path: "noPool", Value: resetNoPool},
			{Path: "reserve", Value: 0.0},
		})
		opsCount++
		if err := commitIfFull(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to reset markets"})
		}
	}

	for _, posDoc := range positionDocs {
		batch.Delete(posDoc.Ref)
		opsCount++
		if err := commitIfFull(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to wipe positions"})
		}
	}

	compRef := db.Client.Collection("competitions").Doc(id)
	batch.Update(compRef, []firestore.Update{
		{Path: "status", Value: "active"},
		{Path: "winningTeamIds", Value: firestore.Delete},
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

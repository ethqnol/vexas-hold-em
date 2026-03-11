package handlers

import (
	"context"
	"log"

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
	// TODO: parse requested status ('paused', 'closed', 'active')
	// TODO: update status in firestore
	return c.JSON(fiber.Map{
		"message": "Competition " + id + " status updated",
		"status":  "success",
	})
}

// finalize event + trigger payout engine
func ResolveCompetition(c *fiber.Ctx) error {
	id := c.Params("id")
	// TODO: identify winning team
	// TODO: run payout engine
	return c.JSON(fiber.Map{
		"message": "Competition " + id + " resolved and payouts distributed",
		"status":  "success",
	})
}

// clear comp + refund users
func ResetCompetition(c *fiber.Ctx) error {
	id := c.Params("id")
	// TODO: reset all AMM pools
	// TODO: reverse txns / refund users
	return c.JSON(fiber.Map{
		"message": "Competition " + id + " has been reset",
		"status":  "success",
	})
}

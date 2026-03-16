package bot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

const competitionURL := fmt.Sprintf("http://localhost:8080/api/v1/competitions")
const marketsURL := fetch(`http://localhost:8080/api/v1/competitions/${id}/markets`)

type User struct {
	Balance     float64 `firestore:"balance"`
	DisplayName string  `firestore:"displayName"`
	Email       string  `firestore:"email"`
}

type Position struct {
	CompetitionID string  `firestore:"competitionId,omitempty"`
	TeamName      string  `firestore:"teamName,omitempty"`
	YesShares float64 `firestore:"yesShares"`
	NoShares  float64 `firestore:"noShares"`
}

type Competition struct {
	Status        string `firestore:"status"` // "active", "closed", "resolved"
	WinningTeamID string `firestore:"winningTeamId,omitempty"`
}

// for buying
type TradeRequest struct {
	CompetitionID string  `json:"competitionId"`
	MarketID      string  `json:"marketId"` // team ID
	UserID        string  `json:"userId"`
	TradeType     string  `json:"tradeType"` // "YES" or "NO"
	Amount        float64 `json:"amount"`    // amt of VEX to spend
}

// for selling
type SellRequest struct {
	CompetitionID string  `json:"competitionId"`
	MarketID      string  `json:"marketId"` // team ID
	UserID        string  `json:"userId"`
	TradeType     string  `json:"tradeType"` // "YES" or "NO"
	Shares        float64 `json:"shares"`    // # of shares to sell
}

func getUser(message string) *discordgo.MessageSend {
	// extract email from message
	match := userCmdRegex.FindStringSubmatch(s)

    if match == nil {
        return &discordgo.MessageSend{
            Content: "Invalid message format",
        }
    }

    email := match[1]

	// CHECK: accurate url?
	userURL := fmt.Sprintf("http://localhost:8080/api/v1/users/%s", email)

	// get the info from the website
	body := goToURL(userURL)

	// convert JSON to User struct
	var user User
	json.Unmarshal([]byte(body), &user)

	// extract user info
	balance := strconv.FormatFloat(user.Balance, 'f', 2, 64)
	displayName := user.DisplayName

	// build discord embed response
	embed := &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{{
			Type: discordgo.EmbedTypeRich,
			Title: "User",
			Description: email,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name: "Balance",
					Value: balance,
					Inline: true,
				},
				{
					Name: "Display Name",
					Value: displayName,
					Inline: true,
				},
			},
		},
	},
	}

	return embed
}

func getPosition(message string) *discordgo.MessageSend {
	// extract email from message
	match := userCmdRegex.FindStringSubmatch(message)

    if match == nil {
        return &discordgo.MessageSend{
            Content: "Invalid message format",
        }
    }

    email := match[1]

	// CHECK: accurate url?
	positionURL := fmt.Sprintf("http://localhost:8080/api/v1/users/%s/positions", email)

	// get the info from the website
	body := goToURL(positionURL)

	// convert JSON to Position struct
	var position Position
	json.Unmarshal([]byte(body), &position)

	// extract position info
	compID := position.CompetitionID
	teamName := position.TeamName
	yesShares := strconv.FormatFloat(position.YesShares, 'f', 2, 64)
	noShares := strconv.FormatFloat(position.NoShares, 'f', 2, 64)

	// build discord embed response
	embed := &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{{
			Type: discordgo.EmbedTypeRich,
			Title: "Position",
			Description: email,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name: "Competition ID",
					Value: compID,
					Inline: true,
				},
				{
					Name: "Team Name",
					Value: teamName,
					Inline: true,
				},
				{
					Name: "Yes Shares",
					Value: yesShares,
					Inline: true,
				},
				{
					Name: "No Shares",
					Value: noShares,
					Inline: true,
				},
			},
		},
	},
	}

	return embed
}

func getCompetition(message string) *discordgo.MessageSend {

	// get the info from the website
	body := goToURL(competitionURL)

	// convert JSON to Competition struct
	var comp Competition
	json.Unmarshal([]byte(body), &comp)

	// extract competition info
	status := comp.Status
	winningTeamID := comp.WinningTeamID

	// build discord embed response
	embed := &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{{
			Type: discordgo.EmbedTypeRich,
			Title: "Competition",
			Description: "Details:",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name: "Status",
					Value: status,
					Inline: true,
				},
				{
					Name: "Winning Team ID",
					Value: winningTeamID,
					Inline: true,
				},
			},
		},
	},
	}

	return embed
}

func goToURL(url string) string {
	// create new http client with timeout
	client := http.Client{Timeout: 5 * time.Second}
	
	// query Vex API
	response, err := client.Get(url)
	if err != nil {
		return &discordgo.MessageSend{
			Content: "Sorry, there was an error trying to reach the URL",
		}
	}

	// open http response body
	body, _ := ioutil.ReadAll(response.Body)
	defer response.Body.Close()

	return body
}


func handleBuy(message string) *discordgo.MessageSend {
	// extract components from message
	var buyCmdRegex = regexp.MustCompile(`^!buy\s+(\S+)\s+(.+)\s+(yes|no)\s+(\d+)$`)
	match := buyCmdRegex.FindStringSubmatch(message)

    if match == nil {
        return &discordgo.MessageSend{
            Content: "Invalid message format",
        }
    }

    email    := match[1]
    teamName := match[2]
    yesNo    := match[3]
    amount   := match[4]

	userID := getUserID(email)
	competitionID := getCompetitionID(teamName)
	marketID := getMarketID(teamName)
	
	tradeRequest := TradeRequest{
		CompetitionID: competitionID,
		MarketID: marketID,
		UserID: userID,
		TradeType: yesNo,
		Amount: amount,
	}

	result := PlaceTrade(tradeRequest)

	return &discordgo.MessageSend{
		Content: result,
	}
}

func handleSell(message string) *discordgo.MessageSend {
	// extract components from message
	var buyCmdRegex = regexp.MustCompile(`^!sell\s+(\S+)\s+(.+)\s+(yes|no)\s+(\d+)$`)
	match := buyCmdRegex.FindStringSubmatch(message)

    if match == nil {
        return &discordgo.MessageSend{
            Content: "Invalid message format",
        }
    }

    email    := match[1]
    teamName := match[2]
    yesNo    := match[3]
    shares   := match[4]

	userID := getUserID(email)
	competitionID := getCompetitionID(teamName)
	marketID := getMarketID(teamName)
	
	sellRequest := SellRequest{
		CompetitionID: competitionID,
		MarketID: marketID,
		UserID: userID,
		TradeType: yesNo,
		Shares: shares,
	}

	result := SellShares(sellRequest)

	return &discordgo.MessageSend{
		Content: result,
	}
}

// gets the IDs for placing trades so user can just type email and team name, which are more convenient
func getUserID(email string) int {

}

func getCompetitionID(teamName string) int {

}

func getMarketID(teamName string) int {

}
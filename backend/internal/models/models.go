package models

type User struct {
	Balance     float64  `firestore:"balance"`
	DisplayName string   `firestore:"displayName"`
	Email       string   `firestore:"email"`
	IsAdmin     bool     `firestore:"isAdmin" json:"isAdmin"`
	TotalLost   float64  `firestore:"totalLost"`
	Title       string   `firestore:"title"`
	Cosmetic    string   `firestore:"cosmetic"`
	OwnedItems  []string `firestore:"ownedItems"`
}

type Position struct {
	CompetitionID string  `firestore:"competitionId,omitempty"`
	TeamName      string  `firestore:"teamName,omitempty"`
	YesShares     float64 `firestore:"yesShares"`
	NoShares      float64 `firestore:"noShares"`
}

type Competition struct {
	Status        string `firestore:"status" json:"status"`
	WinningTeamIDs []string `firestore:"winningTeamIds,omitempty" json:"winningTeamIds,omitempty"`
}

type Market struct {
	TeamID          string  `firestore:"teamId" json:"teamId"`
	TeamName        string  `firestore:"teamName" json:"teamName"`
	Division        string  `firestore:"division" json:"division"`
	Organization    string  `firestore:"organization" json:"organization"`
	Location        string  `firestore:"location" json:"location"`
	YesPool         float64 `firestore:"yesPool" json:"yesPool"`
	NoPool          float64 `firestore:"noPool" json:"noPool"`
	Reserve         float64 `firestore:"reserve" json:"reserve"`
	InitialYesOdds  float64 `firestore:"initialYesOdds" json:"initialYesOdds"`
}

type Transaction struct {
	Timestamp    int64   `firestore:"timestamp"`
	UserID       string  `firestore:"userId"`
	TeamID       string  `firestore:"teamId"`
	TradeType    string  `firestore:"tradeType"`
	AmountSpent  float64 `firestore:"amountSpent"`
	SharesBought float64 `firestore:"sharesBought"`
	YesOdds      float64 `firestore:"yesOdds"`
}

type BlackjackSession struct {
	BetAmount   float64  `firestore:"betAmount"`
	PlayerCards []string `firestore:"playerCards"`
	DealerCards []string `firestore:"dealerCards"`
	Status      string   `firestore:"status"`
}

type LotteryConfig struct {
	Jackpot        float64 `firestore:"jackpot" json:"jackpot"`
	LastWinner     string  `firestore:"lastWinner" json:"lastWinner"`
	LastWinnerName string  `firestore:"lastWinnerName" json:"lastWinnerName"`
	LastWonAt      int64   `firestore:"lastWonAt" json:"lastWonAt"`
}


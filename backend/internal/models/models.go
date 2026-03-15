package models

// firestore user
type User struct {
	Balance     float64 `firestore:"balance"`
	DisplayName string  `firestore:"displayName"`
	Email       string  `firestore:"email"`
}

// user's held shares in a team
type Position struct {
	CompetitionID string  `firestore:"competitionId,omitempty"`
	TeamName      string  `firestore:"teamName,omitempty"`
	YesShares     float64 `firestore:"yesShares"`
	NoShares      float64 `firestore:"noShares"`
}

// entire event
type Competition struct {
	Status        string `firestore:"status"` // "active", "closed", "resolved"
	WinningTeamID string `firestore:"winningTeamId,omitempty"`
}

// amm pool for a team in a comp
type Market struct {
	TeamID       string  `firestore:"teamId" json:"teamId"`
	TeamName     string  `firestore:"teamName" json:"teamName"`
	Division     string  `firestore:"division" json:"division"`
	Organization string  `firestore:"organization" json:"organization"`
	Location     string  `firestore:"location" json:"location"`
	YesPool      float64 `firestore:"yesPool" json:"yesPool"` // amm liquidity
	NoPool       float64 `firestore:"noPool" json:"noPool"`   // amm liquidity
}

// ledger entry
type Transaction struct {
	Timestamp    int64   `firestore:"timestamp"`
	UserID       string  `firestore:"userId"`
	TeamID       string  `firestore:"teamId"`
	TradeType    string  `firestore:"tradeType"` // "YES", "NO", "SELL_YES", "SELL_NO"
	AmountSpent  float64 `firestore:"amountSpent"`
	SharesBought float64 `firestore:"sharesBought"`
	YesOdds      float64 `firestore:"yesOdds"` // implied prob at time of trade
}

package polymarket

// Position represents a user's position from the Data API.
type Position struct {
	ProxyWallet        string  `json:"proxyWallet"`
	Asset              string  `json:"asset"`
	ConditionID        string  `json:"conditionId"`
	Size               float64 `json:"size"`
	AvgPrice           float64 `json:"avgPrice"`
	InitialValue       float64 `json:"initialValue"`
	CurrentValue       float64 `json:"currentValue"`
	CashPnl            float64 `json:"cashPnl"`
	PercentPnl         float64 `json:"percentPnl"`
	TotalBought        float64 `json:"totalBought"`
	RealizedPnl        float64 `json:"realizedPnl"`
	PercentRealizedPnl float64 `json:"percentRealizedPnl"`
	CurPrice           float64 `json:"curPrice"`
	Redeemable         bool    `json:"redeemable"`
	Mergeable          bool    `json:"mergeable"`
	Title              string  `json:"title"`
	Slug               string  `json:"slug"`
	Icon               string  `json:"icon"`
	EventSlug          string  `json:"eventSlug"`
	Outcome            string  `json:"outcome"`
	OutcomeIndex       int     `json:"outcomeIndex"`
	OppositeOutcome    string  `json:"oppositeOutcome"`
	OppositeAsset      string  `json:"oppositeAsset"`
	EndDate            string  `json:"endDate"`
	NegativeRisk       bool    `json:"negativeRisk"`
}

// Trade represents a trade from the Data API.
type Trade struct {
	ProxyWallet           string  `json:"proxyWallet"`
	Side                  string  `json:"side"`
	Asset                 string  `json:"asset"`
	ConditionID           string  `json:"conditionId"`
	Size                  float64 `json:"size"`
	Price                 float64 `json:"price"`
	Timestamp             int64   `json:"timestamp"`
	Title                 string  `json:"title"`
	Slug                  string  `json:"slug"`
	Icon                  string  `json:"icon"`
	EventSlug             string  `json:"eventSlug"`
	Outcome               string  `json:"outcome"`
	OutcomeIndex          int     `json:"outcomeIndex"`
	Name                  string  `json:"name"`
	Pseudonym             string  `json:"pseudonym"`
	Bio                   string  `json:"bio"`
	ProfileImage          string  `json:"profileImage"`
	ProfileImageOptimized string  `json:"profileImageOptimized"`
	TransactionHash       string  `json:"transactionHash"`
}

// Activity represents an activity entry from the Data API.
type Activity struct {
	ProxyWallet           string  `json:"proxyWallet"`
	Timestamp             int64   `json:"timestamp"`
	ConditionID           string  `json:"conditionId"`
	Type                  string  `json:"type"`
	Size                  float64 `json:"size"`
	USDCSize              float64 `json:"usdcSize"`
	TransactionHash       string  `json:"transactionHash"`
	Price                 float64 `json:"price"`
	Asset                 string  `json:"asset"`
	Side                  string  `json:"side"`
	OutcomeIndex          int     `json:"outcomeIndex"`
	Title                 string  `json:"title"`
	Slug                  string  `json:"slug"`
	Icon                  string  `json:"icon"`
	EventSlug             string  `json:"eventSlug"`
	Outcome               string  `json:"outcome"`
	Name                  string  `json:"name"`
	Pseudonym             string  `json:"pseudonym"`
	Bio                   string  `json:"bio"`
	ProfileImage          string  `json:"profileImage"`
	ProfileImageOptimized string  `json:"profileImageOptimized"`
}

// Holder represents a holder of a market token.
type Holder struct {
	ProxyWallet            string  `json:"proxyWallet"`
	Bio                    string  `json:"bio"`
	Asset                  string  `json:"asset"`
	Pseudonym              string  `json:"pseudonym"`
	Amount                 float64 `json:"amount"`
	DisplayUsernamePublic  bool    `json:"displayUsernamePublic"`
	OutcomeIndex           int     `json:"outcomeIndex"`
	Name                   string  `json:"name"`
	ProfileImage           string  `json:"profileImage"`
	ProfileImageOptimized  string  `json:"profileImageOptimized"`
}

// TokenHolders groups holders by token.
type TokenHolders struct {
	Token   string   `json:"token"`
	Holders []Holder `json:"holders"`
}

// TradesParams holds query parameters for the trades endpoint.
type TradesParams struct {
	User         string
	Market       string
	Limit        *int
	Offset       *int
	Side         *Side
	TakerOnly    *bool
	FilterType   string // CASH or TOKENS
	FilterAmount *float64
}

// PositionsParams holds query parameters for the positions endpoint.
type PositionsParams struct {
	User          string
	Market        string
	Limit         *int
	Offset        *int
	SizeThreshold *float64
	Redeemable    *bool
	Mergeable     *bool
	Title         string
	SortBy        string
	SortDirection string
}

// ActivityParams holds query parameters for the activity endpoint.
type ActivityParams struct {
	User          string
	Market        string
	Limit         *int
	Offset        *int
	Type          string // comma-separated: TRADE,SPLIT,MERGE,REDEEM,REWARD,CONVERSION
	Side          *Side
	Start         *int64
	End           *int64
	SortBy        string
	SortDirection string
}

// OpenInterest represents open interest for a market.
type OpenInterest struct {
	Market string  `json:"market"`
	Value  float64 `json:"value"`
}

// HoldersParams holds query parameters for the holders endpoint.
type HoldersParams struct {
	Market string
	Limit  *int
}

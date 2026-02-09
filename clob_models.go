package polymarket

// ClobMarket represents a market returned by the CLOB API.
type ClobMarket struct {
	ConditionID              string    `json:"condition_id"`
	QuestionID               string    `json:"question_id"`
	Question                 string    `json:"question"`
	Description              string    `json:"description"`
	MarketSlug               string    `json:"market_slug"`
	EndDateISO               string    `json:"end_date_iso"`
	GameStartTime            string    `json:"game_start_time"`
	Icon                     string    `json:"icon"`
	Image                    string    `json:"image"`
	FPMM                     string    `json:"fpmm"`
	Active                   bool      `json:"active"`
	Closed                   bool      `json:"closed"`
	Archived                 bool      `json:"archived"`
	AcceptingOrders          bool      `json:"accepting_orders"`
	AcceptingOrderTimestamp   string    `json:"accepting_order_timestamp"`
	EnableOrderBook          bool      `json:"enable_order_book"`
	Is5050Outcome            bool      `json:"is_50_50_outcome"`
	NegRisk                  bool      `json:"neg_risk"`
	NegRiskMarketID          string    `json:"neg_risk_market_id"`
	NegRiskRequestID         string    `json:"neg_risk_request_id"`
	NotificationsEnabled     bool      `json:"notifications_enabled"`
	MinimumOrderSize         string    `json:"minimum_order_size"`
	MinimumTickSize          string    `json:"minimum_tick_size"`
	MakerBaseFee             float64   `json:"maker_base_fee"`
	TakerBaseFee             float64   `json:"taker_base_fee"`
	SecondsDelay             int       `json:"seconds_delay"`
	Tags                     []string  `json:"tags"`
	Tokens                   []Token   `json:"tokens"`
	Rewards                  *Rewards  `json:"rewards"`
}

// Token represents a token within a CLOB market.
type Token struct {
	TokenID  string  `json:"token_id"`
	Outcome  string  `json:"outcome"`
	Price    float64 `json:"price"`
	Winner   bool    `json:"winner"`
}

// Rewards holds reward configuration for a market.
type Rewards struct {
	MinSize   string          `json:"min_size"`
	MaxSpread string          `json:"max_spread"`
	Rates     []RewardRate    `json:"rates"`
}

// RewardRate represents a reward rate tier.
type RewardRate struct {
	AssetAddress string `json:"asset_address"`
	Rewards      string `json:"rewards"`
}

// OrderBook represents the order book for a token.
type OrderBook struct {
	Market         string       `json:"market"`
	AssetID        string       `json:"asset_id"`
	Timestamp      string       `json:"timestamp"`
	Hash           string       `json:"hash"`
	Bids           []OrderLevel `json:"bids"`
	Asks           []OrderLevel `json:"asks"`
	MinOrderSize   string       `json:"min_order_size"`
	TickSize       string       `json:"tick_size"`
	NegRisk        bool         `json:"neg_risk"`
	LastTradePrice string       `json:"last_trade_price"`
}

// OrderLevel represents a single price level in the order book.
type OrderLevel struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

// PriceResponse is returned by the single-token price endpoint.
type PriceResponse struct {
	Price string `json:"price"`
}

// MidpointResponse is returned by the single-token midpoint endpoint.
type MidpointResponse struct {
	Mid string `json:"mid"`
}

// SpreadResponse is returned by the single-token spread endpoint.
type SpreadResponse struct {
	Spread string `json:"spread"`
}

// LastTradePrice is returned by the last trade price endpoint.
type LastTradePrice struct {
	Price string `json:"price"`
	Side  string `json:"side"`
}

// LastTradesPriceEntry is returned by the bulk last trades prices endpoint.
type LastTradesPriceEntry struct {
	Price   string `json:"price"`
	Side    string `json:"side"`
	TokenID string `json:"token_id"`
}

// PriceHistoryParams holds parameters for the price history endpoint.
type PriceHistoryParams struct {
	Market   string
	StartTs  *int64
	EndTs    *int64
	Interval Interval
	Fidelity *int
}

// PriceHistoryResponse is the response from the price history endpoint.
type PriceHistoryResponse struct {
	History []PricePoint `json:"history"`
}

// PricePoint represents a single price point in a price history response.
type PricePoint struct {
	T int64   `json:"t"`
	P float64 `json:"p"`
}

// MarketTradeEvent represents a trade event for a market.
type MarketTradeEvent struct {
	EventType       string                `json:"event_type"`
	Market          MarketTradeEventInfo  `json:"market"`
	User            MarketTradeEventUser  `json:"user"`
	Side            string                `json:"side"`
	Size            string                `json:"size"`
	FeeRateBPS      string                `json:"fee_rate_bps"`
	Price           string                `json:"price"`
	Outcome         string                `json:"outcome"`
	OutcomeIndex    int                   `json:"outcome_index"`
	TransactionHash string                `json:"transaction_hash"`
	Timestamp       string                `json:"timestamp"`
}

// MarketTradeEventInfo contains market info within a trade event.
type MarketTradeEventInfo struct {
	ConditionID string `json:"condition_id"`
	AssetID     string `json:"asset_id"`
	Question    string `json:"question"`
	Icon        string `json:"icon"`
	Slug        string `json:"slug"`
}

// MarketTradeEventUser contains user info within a trade event.
type MarketTradeEventUser struct {
	Address                  string `json:"address"`
	Username                 string `json:"username"`
	ProfilePicture           string `json:"profile_picture"`
	OptimizedProfilePicture  string `json:"optimized_profile_picture"`
	Pseudonym                string `json:"pseudonym"`
}

// SimplifiedMarket is a lightweight market representation.
type SimplifiedMarket struct {
	ConditionID     string   `json:"condition_id"`
	Active          bool     `json:"active"`
	Closed          bool     `json:"closed"`
	Archived        bool     `json:"archived"`
	AcceptingOrders bool     `json:"accepting_orders"`
	Tokens          []Token  `json:"tokens"`
	Rewards         *Rewards `json:"rewards"`
}

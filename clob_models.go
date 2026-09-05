package polymarket

import "encoding/json"

// ClobMarket represents a market returned by the CLOB API.
type ClobMarket struct {
	ConditionID             string     `json:"condition_id"`
	QuestionID              string     `json:"question_id"`
	Question                string     `json:"question"`
	Description             string     `json:"description"`
	MarketSlug              string     `json:"market_slug"`
	EndDateISO              string     `json:"end_date_iso"`
	GameStartTime           string     `json:"game_start_time"`
	Icon                    string     `json:"icon"`
	Image                   string     `json:"image"`
	FPMM                    string     `json:"fpmm"`
	Active                  bool       `json:"active"`
	Closed                  bool       `json:"closed"`
	Archived                bool       `json:"archived"`
	AcceptingOrders         bool       `json:"accepting_orders"`
	AcceptingOrderTimestamp string     `json:"accepting_order_timestamp"`
	EnableOrderBook         bool       `json:"enable_order_book"`
	Is5050Outcome           bool       `json:"is_50_50_outcome"`
	NegRisk                 bool       `json:"neg_risk"`
	NegRiskMarketID         string     `json:"neg_risk_market_id"`
	NegRiskRequestID        string     `json:"neg_risk_request_id"`
	NotificationsEnabled    bool       `json:"notifications_enabled"`
	MinimumOrderSize        FlexString `json:"minimum_order_size"`
	MinimumTickSize         FlexString `json:"minimum_tick_size"`
	MakerBaseFee            float64    `json:"maker_base_fee"`
	TakerBaseFee            float64    `json:"taker_base_fee"`
	SecondsDelay            int        `json:"seconds_delay"`
	Tags                    []string   `json:"tags"`
	Tokens                  []Token    `json:"tokens"`
	Rewards                 *Rewards   `json:"rewards"`
}

// Token represents a token within a CLOB market.
type Token struct {
	TokenID string  `json:"token_id"`
	Outcome string  `json:"outcome"`
	Price   float64 `json:"price"`
	Winner  bool    `json:"winner"`
}

// Rewards holds reward configuration for a market.
type Rewards struct {
	MinSize   FlexString   `json:"min_size"`
	MaxSpread FlexString   `json:"max_spread"`
	Rates     []RewardRate `json:"rates"`
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
	EventType       string               `json:"event_type"`
	Market          MarketTradeEventInfo `json:"market"`
	User            MarketTradeEventUser `json:"user"`
	Side            string               `json:"side"`
	Size            string               `json:"size"`
	FeeRateBPS      string               `json:"fee_rate_bps"`
	Price           string               `json:"price"`
	Outcome         string               `json:"outcome"`
	OutcomeIndex    int                  `json:"outcome_index"`
	TransactionHash string               `json:"transaction_hash"`
	Timestamp       string               `json:"timestamp"`
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
	Address                 string `json:"address"`
	Username                string `json:"username"`
	ProfilePicture          string `json:"profile_picture"`
	OptimizedProfilePicture string `json:"optimized_profile_picture"`
	Pseudonym               string `json:"pseudonym"`
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

// Order represents a full order returned by the CLOB API.
type Order struct {
	ID                string      `json:"id"`
	Market            string      `json:"market"`
	AssetID           string      `json:"asset_id"`
	Side              string      `json:"side"`
	OriginalSize      string      `json:"original_size"`
	SizeMatched       string      `json:"size_matched"`
	Price             string      `json:"price"`
	Outcome           string      `json:"outcome"`
	Owner             string      `json:"owner"`
	Status            string      `json:"status"`
	ExpirationTime    string      `json:"expiration"`
	Type              string      `json:"type"`
	CreatedAt         json.Number `json:"created_at"`
	AssociateTradeIDs []string    `json:"associate_trades"`
}

// OpenOrder represents an open order in a list response.
type OpenOrder struct {
	ID             string      `json:"id"`
	Market         string      `json:"market"`
	AssetID        string      `json:"asset_id"`
	Side           string      `json:"side"`
	OriginalSize   string      `json:"original_size"`
	SizeMatched    string      `json:"size_matched"`
	Price          string      `json:"price"`
	Outcome        string      `json:"outcome"`
	Owner          string      `json:"owner"`
	Status         string      `json:"status"`
	ExpirationTime string      `json:"expiration"`
	Type           string      `json:"type"`
	CreatedAt      json.Number `json:"created_at"`
}

// UserTrade represents a trade in the authenticated user's trade history.
type UserTrade struct {
	TakerOrderID    string       `json:"taker_order_id"`
	TraderSide      string       `json:"trader_side"`
	LastUpdate      string       `json:"last_update"`
	MakerOrders     []MakerOrder `json:"maker_orders"`
	ID              string       `json:"id"`
	Market          string       `json:"market"`
	AssetID         string       `json:"asset_id"`
	Side            string       `json:"side"`
	Size            string       `json:"size"`
	FeeRateBPS      string       `json:"fee_rate_bps"`
	Price           string       `json:"price"`
	Status          string       `json:"status"`
	Outcome         string       `json:"outcome"`
	Owner           string       `json:"owner"`
	MakerAddress    string       `json:"maker_address"`
	MatchTime       string       `json:"match_time"`
	TradeOwner      string       `json:"trader"`
	TransactionHash string       `json:"transaction_hash"`
	BucketIndex     int          `json:"bucket_index"`
	Type            string       `json:"type"`
	CreatedAt       json.Number  `json:"created_at"`
}

// MakerOrder attributes the maker's portion of an execution to its order.
type MakerOrder struct {
	OrderID       string `json:"order_id"`
	Owner         string `json:"owner"`
	MakerAddress  string `json:"maker_address"`
	MatchedAmount string `json:"matched_amount"`
	Price         string `json:"price"`
	FeeRateBPS    string `json:"fee_rate_bps"`
	AssetID       string `json:"asset_id"`
	Outcome       string `json:"outcome"`
	Side          Side   `json:"side,omitempty"`
}

// BuilderTrade is a V2 execution attributed to a public builder code.
type BuilderTrade struct {
	ID              string  `json:"id"`
	TradeType       string  `json:"tradeType"`
	TakerOrderHash  string  `json:"takerOrderHash"`
	Builder         string  `json:"builder"`
	BuilderCode     string  `json:"builderCode"`
	Market          string  `json:"market"`
	AssetID         string  `json:"assetId"`
	Side            string  `json:"side"`
	Size            string  `json:"size"`
	SizeUSDC        string  `json:"sizeUsdc"`
	Price           string  `json:"price"`
	Status          string  `json:"status"`
	Outcome         string  `json:"outcome"`
	OutcomeIndex    int     `json:"outcomeIndex"`
	Owner           string  `json:"owner"`
	Maker           string  `json:"maker"`
	TransactionHash string  `json:"transactionHash"`
	MatchTime       string  `json:"matchTime"`
	BucketIndex     int     `json:"bucketIndex"`
	Fee             string  `json:"fee"`
	FeeUSDC         string  `json:"feeUsdc"`
	BuilderFee      string  `json:"builderFee"`
	ErrorMsg        *string `json:"err_msg"`
	CreatedAt       *string `json:"createdAt"`
	UpdatedAt       *string `json:"updatedAt"`
}

// BalanceAllowance holds a user's balance and allowance information.
type BalanceAllowance struct {
	Allowances map[string]string `json:"allowances"`
	Balance    string            `json:"balance"`
	Allowance  string            `json:"allowance"`
}

// PostOrderOptions controls execution without changing the signed order.
type PostOrderOptions struct {
	PostOnly  bool // Only valid with GTC or GTD
	DeferExec bool
}

// Notification represents a CLOB notification.
type Notification struct {
	Type      string `json:"type"`
	Payload   string `json:"payload"`
	Timestamp string `json:"timestamp"`
}

// OrdersParams holds filter parameters for the GetOrders endpoint.
type OrdersParams struct {
	Market *string
	Asset  *string
}

// ClobTradesParams holds filter parameters for the authenticated CLOB GetTrades endpoint.
type ClobTradesParams struct {
	Market *string
	Asset  *string
}

// BalanceParams holds parameters for the GetBalanceAllowance endpoint.
type BalanceParams struct {
	AssetType     *string
	TokenID       *string
	SignatureType *int
}

// CancelOrderPayload is the request body for canceling a single order.
type CancelOrderPayload struct {
	ID string `json:"orderID"`
}

// CancelOrdersPayload is retained for source compatibility.
// Deprecated: CancelOrders now sends the V2 JSON array directly.
type CancelOrdersPayload struct {
	OrderIDs []string `json:"ids"`
}

// cancelResponse is the response body returned by the cancel endpoints.
type cancelResponse struct {
	Canceled    []string          `json:"canceled"`
	NotCanceled map[string]string `json:"not_canceled"`
}

// postOrderJSON is the JSON representation of a signed order for the POST /order endpoint.
// Field types match the official V2 client:
//   - salt and signatureType are JSON numbers (bare integers)
//   - tokenId, makerAmount, takerAmount, expiration, timestamp are strings
//   - side is "BUY" or "SELL"
type postOrderJSON struct {
	Salt          json.Number `json:"salt"`
	Maker         string      `json:"maker"`
	Signer        string      `json:"signer"`
	TokenID       string      `json:"tokenId"`
	MakerAmount   string      `json:"makerAmount"`
	TakerAmount   string      `json:"takerAmount"`
	Expiration    string      `json:"expiration"`
	Timestamp     string      `json:"timestamp"`
	Metadata      string      `json:"metadata"`
	Builder       string      `json:"builder"`
	Side          string      `json:"side"`
	SignatureType int         `json:"signatureType"`
	Signature     string      `json:"signature"`
}

// postOrderRequest is the request body for POST /order.
// The Owner field must be set to the API key (not the wallet address),
// matching the Python client: order_to_json(order, self.creds.api_key, ...).
type postOrderRequest struct {
	Order     postOrderJSON `json:"order"`
	Owner     string        `json:"owner"`
	OrderType string        `json:"orderType"`
	PostOnly  bool          `json:"postOnly"`
	DeferExec bool          `json:"deferExec"`
}

// PostOrderResponse is the response from the POST /order endpoint.
type PostOrderResponse struct {
	Status            string   `json:"status"`
	MakingAmount      string   `json:"makingAmount"`
	TakingAmount      string   `json:"takingAmount"`
	TradeIDs          []string `json:"tradeIDs"`
	Success           bool     `json:"success"`
	ErrorMsg          string   `json:"errorMsg"`
	OrderID           string   `json:"orderID"`
	TransactionHashes []string `json:"transactionsHashes"`
}

// ApiCredentials holds API key credentials returned by the CLOB API key endpoints.
// These are L2 credentials used for HMAC authentication on subsequent requests.
//
// Use ClobClient.CreateOrDeriveApiCreds to obtain these from a private key.
type ApiCredentials struct {
	ApiKey     string `json:"apiKey"`
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
}

// ToCredentials converts ApiCredentials to a Credentials struct suitable for
// passing to WithCredentials. You must provide the wallet address separately.
func (a *ApiCredentials) ToCredentials(address string) *Credentials {
	return &Credentials{
		Key:        a.ApiKey,
		Secret:     a.Secret,
		Passphrase: a.Passphrase,
		Address:    address,
	}
}

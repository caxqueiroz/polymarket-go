package polymarket

import (
	"encoding/json"
	"strconv"
	"strings"
)

// GammaMarket represents a market from the Gamma API.
type GammaMarket struct {
	ID                 string      `json:"id"`
	Question           string      `json:"question"`
	ConditionID        string      `json:"conditionId"`
	Slug               string      `json:"slug"`
	EndDate            string      `json:"endDate"`
	StartDate          string      `json:"startDate"`
	Description        string      `json:"description"`
	ResolutionSource   string      `json:"resolutionSource"`
	Image              string      `json:"image"`
	Icon               string      `json:"icon"`
	Active             bool        `json:"active"`
	Closed             bool        `json:"closed"`
	Archived           bool        `json:"archived"`
	Featured           bool        `json:"featured"`
	Restricted         bool        `json:"restricted"`
	Liquidity          Number      `json:"liquidity"`
	Volume             Number      `json:"volume"`
	Volume24hr         Number      `json:"volume24hr"`
	OpenInterest       Number      `json:"openInterest"`
	EnableOrderBook    bool        `json:"enableOrderBook"`
	LiquidityClob      Number      `json:"liquidityClob"`
	LiquidityAmm       Number      `json:"liquidityAmm"`
	NegRisk            bool        `json:"negRisk"`
	NegRiskMarketID    string      `json:"negRiskMarketID"`
	NegRiskFeeBips     Number      `json:"negRiskFeeBips"`
	CommentCount       int         `json:"commentCount"`
	ClobTokenIDs       StringSlice `json:"clobTokenIds"`
	Outcomes           StringSlice `json:"outcomes"`
	OutcomePrices      StringSlice `json:"outcomePrices"`
	BestBid            Number      `json:"bestBid"`
	BestAsk            Number      `json:"bestAsk"`
	LastTradePrice     Number      `json:"lastTradePrice"`
	Spread             Number      `json:"spread"`
	OneDayPriceChange  Number      `json:"oneDayPriceChange"`
	Category           string      `json:"category"`
	EventSlug          string      `json:"eventSlug"`
}

// GammaEvent represents an event from the Gamma API.
type GammaEvent struct {
	ID               string        `json:"id"`
	Ticker           string        `json:"ticker"`
	Slug             string        `json:"slug"`
	Title            string        `json:"title"`
	Subtitle         string        `json:"subtitle"`
	Description      string        `json:"description"`
	ResolutionSource string        `json:"resolutionSource"`
	StartDate        string        `json:"startDate"`
	EndDate          string        `json:"endDate"`
	CreationDate     string        `json:"creationDate"`
	Image            string        `json:"image"`
	Icon             string        `json:"icon"`
	Active           bool          `json:"active"`
	Closed           bool          `json:"closed"`
	Archived         bool          `json:"archived"`
	Featured         bool          `json:"featured"`
	Restricted       bool          `json:"restricted"`
	Liquidity        Number        `json:"liquidity"`
	Volume           Number        `json:"volume"`
	OpenInterest     Number        `json:"openInterest"`
	Category         string        `json:"category"`
	Subcategory      string        `json:"subcategory"`
	EnableOrderBook  bool          `json:"enableOrderBook"`
	LiquidityAmm     Number        `json:"liquidityAmm"`
	LiquidityClob    Number        `json:"liquidityClob"`
	NegRisk          bool          `json:"negRisk"`
	CommentCount     int           `json:"commentCount"`
	Volume24hr       Number        `json:"volume24hr"`
	Volume1wk        Number        `json:"volume1wk"`
	Volume1mo        Number        `json:"volume1mo"`
	Volume1yr        Number        `json:"volume1yr"`
	Markets          []GammaMarket `json:"markets"`
	Tags             []GammaTag    `json:"tags"`
}

// GammaTag represents a tag from the Gamma API.
type GammaTag struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Slug      string `json:"slug"`
	ForceShow bool   `json:"forceShow"`
	ForceHide bool   `json:"forceHide"`
}

// SearchResult represents a search result from the Gamma API.
type SearchResult struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Icon        string `json:"icon"`
	Active      bool   `json:"active"`
	Closed      bool   `json:"closed"`
}

// GammaMarketParams holds query parameters for the Gamma markets endpoint.
type GammaMarketParams struct {
	Limit          *int
	Offset         *int
	Order          string
	Ascending      *bool
	ID             string
	Slug           string
	Closed         *bool
	Active         *bool
	TagID          *int
	ExcludeTagID   []int
	RelatedTags    *bool
	Featured       *bool
	EnableOrderBook *bool
}

// GammaEventParams holds query parameters for the Gamma events endpoint.
type GammaEventParams struct {
	Limit        *int
	Offset       *int
	Order        string
	Ascending    *bool
	ID           []int
	Slug         []string
	TagID        *int
	ExcludeTagID []int
	RelatedTags  *bool
	Featured     *bool
	Closed       *bool
	StartDateMin string
	StartDateMax string
	EndDateMin   string
	EndDateMax   string
}

// Number is a float64 that can be unmarshaled from either a JSON number or a
// JSON string. The Gamma API inconsistently returns numeric fields as either
// numbers or strings.
type Number float64

// Float64 returns the Number as a float64.
func (n Number) Float64() float64 { return float64(n) }

// UnmarshalJSON implements json.Unmarshaler for Number.
func (n *Number) UnmarshalJSON(data []byte) error {
	var f float64
	if err := json.Unmarshal(data, &f); err == nil {
		*n = Number(f)
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		*n = 0
		return nil
	}
	if s == "" {
		*n = 0
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		*n = 0
		return nil
	}
	*n = Number(f)
	return nil
}

// MarshalJSON implements json.Marshaler for Number.
func (n Number) MarshalJSON() ([]byte, error) {
	return json.Marshal(float64(n))
}

// StringSlice handles Gamma API fields that can be either a JSON array or a JSON
// string containing a JSON array (e.g. "[\"Yes\",\"No\"]").
type StringSlice []string

// UnmarshalJSON implements json.Unmarshaler for StringSlice.
func (s *StringSlice) UnmarshalJSON(data []byte) error {
	// Try array first.
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		*s = arr
		return nil
	}

	// Try quoted string containing a JSON array.
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		*s = nil
		return nil
	}

	str = strings.TrimSpace(str)
	if str == "" || str == "[]" {
		*s = nil
		return nil
	}

	if err := json.Unmarshal([]byte(str), &arr); err != nil {
		*s = nil
		return nil
	}
	*s = arr
	return nil
}

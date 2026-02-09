package polymarket

import (
	"encoding/json"
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
	Liquidity          float64     `json:"liquidity"`
	Volume             float64     `json:"volume"`
	Volume24hr         float64     `json:"volume24hr"`
	OpenInterest       float64     `json:"openInterest"`
	EnableOrderBook    bool        `json:"enableOrderBook"`
	LiquidityClob      float64     `json:"liquidityClob"`
	LiquidityAmm       float64     `json:"liquidityAmm"`
	NegRisk            bool        `json:"negRisk"`
	NegRiskMarketID    string      `json:"negRiskMarketID"`
	NegRiskFeeBips     float64     `json:"negRiskFeeBips"`
	CommentCount       int         `json:"commentCount"`
	ClobTokenIDs       StringSlice `json:"clobTokenIds"`
	Outcomes           StringSlice `json:"outcomes"`
	OutcomePrices      StringSlice `json:"outcomePrices"`
	BestBid            float64     `json:"bestBid"`
	BestAsk            float64     `json:"bestAsk"`
	LastTradePrice     float64     `json:"lastTradePrice"`
	Spread             float64     `json:"spread"`
	OneDayPriceChange  float64     `json:"oneDayPriceChange"`
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
	Liquidity        float64       `json:"liquidity"`
	Volume           float64       `json:"volume"`
	OpenInterest     float64       `json:"openInterest"`
	Category         string        `json:"category"`
	Subcategory      string        `json:"subcategory"`
	EnableOrderBook  bool          `json:"enableOrderBook"`
	LiquidityAmm     float64       `json:"liquidityAmm"`
	LiquidityClob    float64       `json:"liquidityClob"`
	NegRisk          bool          `json:"negRisk"`
	CommentCount     int           `json:"commentCount"`
	Volume24hr       float64       `json:"volume24hr"`
	Volume1wk        float64       `json:"volume1wk"`
	Volume1mo        float64       `json:"volume1mo"`
	Volume1yr        float64       `json:"volume1yr"`
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

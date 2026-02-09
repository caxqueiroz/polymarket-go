package polymarket

import (
	"encoding/json"
	"strconv"
)

// FlexString is a string that can be unmarshaled from either a JSON string or a
// JSON number. Some Polymarket API fields return numeric values as bare JSON
// numbers rather than quoted strings.
type FlexString string

// String returns the FlexString as a plain string.
func (s FlexString) String() string { return string(s) }

// UnmarshalJSON implements json.Unmarshaler for FlexString.
func (s *FlexString) UnmarshalJSON(data []byte) error {
	// Try string first.
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = FlexString(str)
		return nil
	}

	// Try number — format as string without precision loss.
	var f float64
	if err := json.Unmarshal(data, &f); err == nil {
		*s = FlexString(strconv.FormatFloat(f, 'f', -1, 64))
		return nil
	}

	// Null or other — leave as empty string.
	*s = ""
	return nil
}

// MarshalJSON implements json.Marshaler for FlexString.
func (s FlexString) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

// Side represents buy or sell.
type Side string

const (
	Buy  Side = "BUY"
	Sell Side = "SELL"
)

// Interval for price history.
type Interval string

const (
	Interval1m  Interval = "1m"
	Interval5m  Interval = "5m"
	Interval15m Interval = "15m"
	Interval1h  Interval = "1h"
	Interval4h  Interval = "4h"
	Interval6h  Interval = "6h"
	Interval1d  Interval = "1d"
	Interval1w  Interval = "1w"
)

// TickSize represents the minimum price increment for a market.
type TickSize string

const (
	TickSize0_1  TickSize = "0.1"
	TickSize0_01 TickSize = "0.01"
	TickSize0_001 TickSize = "0.001"
	TickSize0_0001 TickSize = "0.0001"
)

// BookParams holds query parameters for fetching an order book.
type BookParams struct {
	TokenID string
	Side    *Side
}

package polymarket

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

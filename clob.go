package polymarket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ClobClient provides access to the Polymarket CLOB API.
type ClobClient struct {
	base    *baseClient
	baseURL string
}

// Health checks if the CLOB API is operational.
func (c *ClobClient) Health(ctx context.Context) error {
	return c.base.get(ctx, c.baseURL, "/", nil, nil)
}

// ServerTime returns the current server time as a Unix timestamp string.
func (c *ClobClient) ServerTime(ctx context.Context) (string, error) {
	body, err := c.base.getRaw(ctx, c.baseURL, "/time", nil)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(body)), nil
}

// GetMarket retrieves a single market by condition ID.
func (c *ClobClient) GetMarket(ctx context.Context, conditionID string) (*ClobMarket, error) {
	var m ClobMarket
	if err := c.base.get(ctx, c.baseURL, "/markets/"+conditionID, nil, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// ListMarkets returns a single page of markets. Pass an empty cursor for the first page.
func (c *ClobClient) ListMarkets(ctx context.Context, nextCursor string) (*CursorPage[ClobMarket], error) {
	params := url.Values{}
	if nextCursor != "" {
		params.Set("next_cursor", nextCursor)
	}

	body, err := c.base.getRaw(ctx, c.baseURL, "/markets", params)
	if err != nil {
		return nil, err
	}

	var page CursorPage[ClobMarket]
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, fmt.Errorf("polymarket: decoding markets page: %w", err)
	}
	return &page, nil
}

// AllMarkets returns a CursorIterator that iterates over all CLOB markets.
func (c *ClobClient) AllMarkets(ctx context.Context) *CursorIterator[ClobMarket] {
	return newCursorIterator[ClobMarket](c.base, c.baseURL, "/markets", nil)
}

// ListSimplifiedMarkets returns a single page of simplified markets.
func (c *ClobClient) ListSimplifiedMarkets(ctx context.Context, nextCursor string) (*CursorPage[SimplifiedMarket], error) {
	params := url.Values{}
	if nextCursor != "" {
		params.Set("next_cursor", nextCursor)
	}

	body, err := c.base.getRaw(ctx, c.baseURL, "/simplified-markets", params)
	if err != nil {
		return nil, err
	}

	var page CursorPage[SimplifiedMarket]
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, fmt.Errorf("polymarket: decoding simplified markets page: %w", err)
	}
	return &page, nil
}

// GetPrice returns the best price for a token on a given side.
func (c *ClobClient) GetPrice(ctx context.Context, tokenID string, side Side) (string, error) {
	params := url.Values{}
	params.Set("token_id", tokenID)
	params.Set("side", string(side))

	var resp PriceResponse
	if err := c.base.get(ctx, c.baseURL, "/price", params, &resp); err != nil {
		return "", err
	}
	return resp.Price, nil
}

// GetPrices returns prices for multiple tokens via POST. Returns a map of tokenID -> {side: price}.
func (c *ClobClient) GetPrices(ctx context.Context, params []BookParams) (map[string]map[string]string, error) {
	body, err := c.base.postJSONRaw(ctx, c.baseURL, "/prices", buildBookParamsPayload(params))
	if err != nil {
		return nil, err
	}

	var result map[string]map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("polymarket: decoding prices: %w", err)
	}
	return result, nil
}

// GetMidpoint returns the midpoint price for a token.
func (c *ClobClient) GetMidpoint(ctx context.Context, tokenID string) (string, error) {
	params := url.Values{}
	params.Set("token_id", tokenID)

	var resp MidpointResponse
	if err := c.base.get(ctx, c.baseURL, "/midpoint", params, &resp); err != nil {
		return "", err
	}
	return resp.Mid, nil
}

// GetMidpoints returns midpoints for multiple tokens via POST.
func (c *ClobClient) GetMidpoints(ctx context.Context, params []BookParams) (map[string]string, error) {
	body, err := c.base.postJSONRaw(ctx, c.baseURL, "/midpoints", buildBookParamsPayload(params))
	if err != nil {
		return nil, err
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("polymarket: decoding midpoints: %w", err)
	}
	return result, nil
}

// GetSpread returns the spread for a token.
func (c *ClobClient) GetSpread(ctx context.Context, tokenID string) (string, error) {
	params := url.Values{}
	params.Set("token_id", tokenID)

	var resp SpreadResponse
	if err := c.base.get(ctx, c.baseURL, "/spread", params, &resp); err != nil {
		return "", err
	}
	return resp.Spread, nil
}

// GetSpreads returns spreads for multiple tokens via POST.
func (c *ClobClient) GetSpreads(ctx context.Context, params []BookParams) (map[string]string, error) {
	body, err := c.base.postJSONRaw(ctx, c.baseURL, "/spreads", buildBookParamsPayload(params))
	if err != nil {
		return nil, err
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("polymarket: decoding spreads: %w", err)
	}
	return result, nil
}

// GetBook returns the order book for a token.
func (c *ClobClient) GetBook(ctx context.Context, tokenID string) (*OrderBook, error) {
	params := url.Values{}
	params.Set("token_id", tokenID)

	var ob OrderBook
	if err := c.base.get(ctx, c.baseURL, "/book", params, &ob); err != nil {
		return nil, err
	}
	return &ob, nil
}

// GetBooks returns order books for multiple tokens via POST.
func (c *ClobClient) GetBooks(ctx context.Context, bookParams []BookParams) ([]OrderBook, error) {
	body, err := c.base.postJSONRaw(ctx, c.baseURL, "/books", buildBookParamsPayload(bookParams))
	if err != nil {
		return nil, err
	}

	var result []OrderBook
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("polymarket: decoding books: %w", err)
	}
	return result, nil
}

// GetLastTradePrice returns the last trade price for a token.
func (c *ClobClient) GetLastTradePrice(ctx context.Context, tokenID string) (*LastTradePrice, error) {
	params := url.Values{}
	params.Set("token_id", tokenID)

	var resp LastTradePrice
	if err := c.base.get(ctx, c.baseURL, "/last-trade-price", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetLastTradesPrices returns last trade prices for multiple tokens via POST.
func (c *ClobClient) GetLastTradesPrices(ctx context.Context, params []BookParams) ([]LastTradesPriceEntry, error) {
	body, err := c.base.postJSONRaw(ctx, c.baseURL, "/last-trades-prices", buildBookParamsPayload(params))
	if err != nil {
		return nil, err
	}

	var result []LastTradesPriceEntry
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("polymarket: decoding last trades prices: %w", err)
	}
	return result, nil
}

// GetPriceHistory returns historical price data for a token.
func (c *ClobClient) GetPriceHistory(ctx context.Context, p PriceHistoryParams) (*PriceHistoryResponse, error) {
	params := url.Values{}
	params.Set("market", p.Market)
	if p.StartTs != nil {
		params.Set("startTs", strconv.FormatInt(*p.StartTs, 10))
	}
	if p.EndTs != nil {
		params.Set("endTs", strconv.FormatInt(*p.EndTs, 10))
	}
	if p.Interval != "" {
		params.Set("interval", string(p.Interval))
	}
	if p.Fidelity != nil {
		params.Set("fidelity", strconv.Itoa(*p.Fidelity))
	}

	var resp PriceHistoryResponse
	if err := c.base.get(ctx, c.baseURL, "/prices-history", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetMarketTradeEvents returns trade events for a market by condition ID.
func (c *ClobClient) GetMarketTradeEvents(ctx context.Context, conditionID string) ([]MarketTradeEvent, error) {
	var events []MarketTradeEvent
	if err := c.base.get(ctx, c.baseURL, "/live-activity/events/"+conditionID, nil, &events); err != nil {
		return nil, err
	}
	return events, nil
}

// GetTickSize returns the tick size for a token.
func (c *ClobClient) GetTickSize(ctx context.Context, tokenID string) (TickSize, error) {
	params := url.Values{}
	params.Set("token_id", tokenID)

	body, err := c.base.getRaw(ctx, c.baseURL, "/tick-size", params)
	if err != nil {
		return "", err
	}

	// Try wrapped object first: {"minimum_tick_size": "0.01"}
	var wrapped struct {
		MinimumTickSize string `json:"minimum_tick_size"`
	}
	if err := json.Unmarshal(body, &wrapped); err == nil && wrapped.MinimumTickSize != "" {
		return TickSize(wrapped.MinimumTickSize), nil
	}

	// Fall back to bare string: "0.01"
	var ts string
	if err := json.Unmarshal(body, &ts); err == nil {
		return TickSize(ts), nil
	}

	return TickSize(strings.Trim(string(body), "\"\n ")), nil
}

// GetNegRisk returns whether a token is in a neg-risk market.
func (c *ClobClient) GetNegRisk(ctx context.Context, tokenID string) (bool, error) {
	params := url.Values{}
	params.Set("token_id", tokenID)

	body, err := c.base.getRaw(ctx, c.baseURL, "/neg-risk", params)
	if err != nil {
		return false, err
	}

	// Try wrapped object first: {"neg_risk": true}
	var wrapped struct {
		NegRisk bool `json:"neg_risk"`
	}
	if err := json.Unmarshal(body, &wrapped); err == nil {
		return wrapped.NegRisk, nil
	}

	// Fall back to bare boolean
	var result bool
	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Errorf("polymarket: decoding neg-risk: %w", err)
	}
	return result, nil
}

// GetFeeRateBPS returns the fee rate in basis points for a token.
func (c *ClobClient) GetFeeRateBPS(ctx context.Context, tokenID string) (float64, error) {
	params := url.Values{}
	params.Set("token_id", tokenID)

	body, err := c.base.getRaw(ctx, c.baseURL, "/fee-rate-bps", params)
	if err != nil {
		return 0, err
	}

	var result float64
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("polymarket: decoding fee rate: %w", err)
	}
	return result, nil
}

// bookParamEntry is the JSON payload shape for bulk POST endpoints.
type bookParamEntry struct {
	TokenID string `json:"token_id"`
	Side    string `json:"side,omitempty"`
}

// GetOrder retrieves a single order by ID. Requires authentication.
func (c *ClobClient) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	var o Order
	if err := c.base.get(ctx, c.baseURL, "/order/"+orderID, nil, &o); err != nil {
		return nil, err
	}
	return &o, nil
}

// GetOrders returns the authenticated user's orders. Requires authentication.
func (c *ClobClient) GetOrders(ctx context.Context, p *OrdersParams) ([]OpenOrder, error) {
	params := url.Values{}
	if p != nil {
		if p.Market != nil {
			params.Set("market", *p.Market)
		}
		if p.Asset != nil {
			params.Set("asset_id", *p.Asset)
		}
	}

	body, err := c.base.getRaw(ctx, c.baseURL, "/orders", params)
	if err != nil {
		return nil, err
	}

	var orders []OpenOrder
	if err := json.Unmarshal(body, &orders); err != nil {
		return nil, fmt.Errorf("polymarket: decoding orders: %w", err)
	}
	return orders, nil
}

// CancelOrder cancels a single order by ID. Requires authentication.
func (c *ClobClient) CancelOrder(ctx context.Context, orderID string) error {
	_, err := c.base.deleteRaw(ctx, c.baseURL, "/order/"+orderID, nil)
	return err
}

// CancelOrders cancels multiple orders by ID. Returns the IDs of canceled orders.
// Requires authentication.
func (c *ClobClient) CancelOrders(ctx context.Context, orderIDs []string) ([]string, error) {
	body, err := c.base.deleteRaw(ctx, c.baseURL, "/orders", CancelOrdersPayload{OrderIDs: orderIDs})
	if err != nil {
		return nil, err
	}

	var canceled []string
	if err := json.Unmarshal(body, &canceled); err != nil {
		return nil, fmt.Errorf("polymarket: decoding cancel response: %w", err)
	}
	return canceled, nil
}

// CancelAll cancels all open orders. Returns the IDs of canceled orders.
// Requires authentication.
func (c *ClobClient) CancelAll(ctx context.Context) ([]string, error) {
	body, err := c.base.deleteRaw(ctx, c.baseURL, "/cancel-all", nil)
	if err != nil {
		return nil, err
	}

	var canceled []string
	if err := json.Unmarshal(body, &canceled); err != nil {
		return nil, fmt.Errorf("polymarket: decoding cancel-all response: %w", err)
	}
	return canceled, nil
}

// GetTrades returns the authenticated user's trades. Requires authentication.
func (c *ClobClient) GetTrades(ctx context.Context, p *ClobTradesParams) ([]UserTrade, error) {
	params := url.Values{}
	if p != nil {
		if p.Market != nil {
			params.Set("market", *p.Market)
		}
		if p.Asset != nil {
			params.Set("asset_id", *p.Asset)
		}
	}

	body, err := c.base.getRaw(ctx, c.baseURL, "/trades", params)
	if err != nil {
		return nil, err
	}

	var trades []UserTrade
	if err := json.Unmarshal(body, &trades); err != nil {
		return nil, fmt.Errorf("polymarket: decoding trades: %w", err)
	}
	return trades, nil
}

// GetBalanceAllowance returns the user's balance and allowance. Requires authentication.
func (c *ClobClient) GetBalanceAllowance(ctx context.Context, p *BalanceParams) (*BalanceAllowance, error) {
	params := url.Values{}
	if p != nil {
		if p.AssetType != nil {
			params.Set("asset_type", *p.AssetType)
		}
	}

	var ba BalanceAllowance
	if err := c.base.get(ctx, c.baseURL, "/balance-allowance", params, &ba); err != nil {
		return nil, err
	}
	return &ba, nil
}

// GetNotifications returns the user's notifications. Requires authentication.
func (c *ClobClient) GetNotifications(ctx context.Context) ([]Notification, error) {
	var notifications []Notification
	if err := c.base.get(ctx, c.baseURL, "/notifications", nil, &notifications); err != nil {
		return nil, err
	}
	return notifications, nil
}

func buildBookParamsPayload(params []BookParams) []bookParamEntry {
	entries := make([]bookParamEntry, len(params))
	for i, p := range params {
		entries[i] = bookParamEntry{TokenID: p.TokenID}
		if p.Side != nil {
			entries[i].Side = string(*p.Side)
		}
	}
	return entries
}

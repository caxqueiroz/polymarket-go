package polymarket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// timeNow is a replaceable function for testing timestamp generation.
var timeNow = time.Now

// ClobClient provides access to the Polymarket CLOB API.
type ClobClient struct {
	base        *baseClient
	baseURL     string
	builderCode string
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

	body, err := c.base.getRaw(ctx, c.baseURL, "/fee-rate", params)
	if err != nil {
		return 0, err
	}

	// Try wrapped object first: {"base_fee": 0}
	var wrapped struct {
		BaseFee float64 `json:"base_fee"`
	}
	if err := json.Unmarshal(body, &wrapped); err == nil {
		return wrapped.BaseFee, nil
	}

	// Fall back to bare number
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
	if err := c.base.get(ctx, c.baseURL, "/data/order/"+orderID, nil, &o); err != nil {
		return nil, err
	}
	return &o, nil
}

// GetOrders returns the authenticated user's orders. Requires authentication.
// Automatically paginates through all results.
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

	var all []OpenOrder
	cursor := "MA=="
	for cursor != "LTE=" {
		params.Set("next_cursor", cursor)
		body, err := c.base.getRaw(ctx, c.baseURL, "/data/orders", params)
		if err != nil {
			return nil, err
		}

		var page struct {
			Data       []OpenOrder `json:"data"`
			NextCursor string      `json:"next_cursor"`
		}
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, fmt.Errorf("polymarket: decoding orders: %w", err)
		}
		all = append(all, page.Data...)
		cursor = page.NextCursor
		if cursor == "" {
			break
		}
	}
	return all, nil
}

// CancelOrder cancels a single order by ID. Requires authentication.
func (c *ClobClient) CancelOrder(ctx context.Context, orderID string) error {
	body, err := c.base.deleteRaw(ctx, c.baseURL, "/order", CancelOrderPayload{ID: orderID})
	if err != nil {
		return err
	}
	_, err = decodeCancelResponse(body, []string{orderID})
	return err
}

// CancelOrders cancels multiple orders by ID. Returns the IDs of canceled orders.
// Requires authentication.
func (c *ClobClient) CancelOrders(ctx context.Context, orderIDs []string) ([]string, error) {
	if len(orderIDs) == 0 {
		return nil, nil
	}
	body, err := c.base.deleteRaw(ctx, c.baseURL, "/orders", orderIDs)
	if err != nil {
		return nil, err
	}

	return decodeCancelResponse(body, orderIDs)
}

// CancelAll cancels all open orders. Returns the IDs of canceled orders.
// Requires authentication.
func (c *ClobClient) CancelAll(ctx context.Context) ([]string, error) {
	body, err := c.base.deleteRaw(ctx, c.baseURL, "/cancel-all", nil)
	if err != nil {
		return nil, err
	}

	return decodeCancelResponse(body, nil)
}

func decodeCancelResponse(body []byte, requested []string) ([]string, error) {
	var resp cancelResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("polymarket: decoding cancel response: %w", err)
	}
	if resp.Canceled == nil && resp.NotCanceled == nil {
		return nil, fmt.Errorf("polymarket: missing cancellation acknowledgement")
	}
	if resp.NotCanceled == nil {
		resp.NotCanceled = make(map[string]string)
	}
	canceled := make(map[string]bool, len(resp.Canceled))
	for _, id := range resp.Canceled {
		canceled[id] = true
	}
	for _, id := range requested {
		if !canceled[id] {
			if _, ok := resp.NotCanceled[id]; !ok {
				resp.NotCanceled[id] = "cancellation not acknowledged"
			}
		}
	}
	if len(resp.NotCanceled) != 0 {
		return resp.Canceled, &CancelError{NotCanceled: resp.NotCanceled}
	}
	return resp.Canceled, nil
}

// GetTrades returns the authenticated user's trades. Requires authentication.
// Automatically paginates through all results.
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

	var all []UserTrade
	cursor := "MA=="
	for cursor != "LTE=" {
		params.Set("next_cursor", cursor)
		body, err := c.base.getRaw(ctx, c.baseURL, "/data/trades", params)
		if err != nil {
			return nil, err
		}

		var page struct {
			Data       []UserTrade `json:"data"`
			NextCursor string      `json:"next_cursor"`
		}
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, fmt.Errorf("polymarket: decoding trades: %w", err)
		}
		all = append(all, page.Data...)
		cursor = page.NextCursor
		if cursor == "" {
			break
		}
	}
	return all, nil
}

// GetBalanceAllowance returns the user's balance and allowance. Requires authentication.
func (c *ClobClient) GetBalanceAllowance(ctx context.Context, p *BalanceParams) (*BalanceAllowance, error) {
	params := url.Values{}
	if p != nil {
		if p.AssetType != nil {
			params.Set("asset_type", *p.AssetType)
		}
		if p.TokenID != nil {
			params.Set("token_id", *p.TokenID)
		}
		if p.SignatureType != nil {
			params.Set("signature_type", strconv.Itoa(*p.SignatureType))
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

// ---- L1 API Key Management ----
//
// These methods use L1 (EIP-712) authentication via the OrderSigner to manage
// API credentials. The returned credentials can be used for L2 (HMAC) authentication
// on subsequent requests (orders, trades, balances, etc.).
//
// Mirrors the Python client's create_or_derive_api_creds() pattern:
//   client = ClobClient(host, key=private_key, chain_id=chain_id)
//   api_creds = client.create_or_derive_api_creds()
//   client.set_api_creds(api_creds)

// CreateOrDeriveApiCreds creates or derives API credentials using L1 (EIP-712) authentication.
// If credentials already exist for the signer's address, they are returned (derived).
// If not, new credentials are created.
//
// This is the primary method for obtaining API credentials from a private key.
// It mirrors the Python/TypeScript client's create_or_derive_api_creds() / createOrDeriveApiKey().
//
// Usage:
//
//	signer, _ := polymarket.NewOrderSigner(privateKey, polymarket.Polygon)
//	creds, _ := client.Clob.CreateOrDeriveApiCreds(ctx, signer)
//	// Re-create client with L2 credentials:
//	authedClient := polymarket.NewClient(polymarket.WithCredentials(creds.ToCredentials(signer.Address())))
func (c *ClobClient) CreateOrDeriveApiCreds(ctx context.Context, signer *OrderSigner) (*ApiCredentials, error) {
	// Try to create a new key first (POST).
	creds, err := c.CreateApiKey(ctx, signer)
	if err == nil {
		return creds, nil
	}
	// If create failed (e.g. key already exists), derive the existing one.
	return c.DeriveApiKey(ctx, signer)
}

// CreateApiKey creates a new API key using L1 (EIP-712) authentication.
// Unlike CreateOrDeriveApiCreds, this always creates a new key even if one already exists.
func (c *ClobClient) CreateApiKey(ctx context.Context, signer *OrderSigner) (*ApiCredentials, error) {
	l1, err := signer.BuildL1AuthHeaders(
		strconv.FormatInt(timeNow().Unix(), 10),
		0,
	)
	if err != nil {
		return nil, fmt.Errorf("polymarket: building L1 auth headers: %w", err)
	}

	var creds ApiCredentials
	if err := c.base.postL1(ctx, c.baseURL, "/auth/api-key", nil, l1, &creds); err != nil {
		return nil, err
	}
	return &creds, nil
}

// DeriveApiKey derives (retrieves) an existing API key using L1 (EIP-712) authentication.
// Returns an error if no key exists for the signer's address.
func (c *ClobClient) DeriveApiKey(ctx context.Context, signer *OrderSigner) (*ApiCredentials, error) {
	l1, err := signer.BuildL1AuthHeaders(
		strconv.FormatInt(timeNow().Unix(), 10),
		0,
	)
	if err != nil {
		return nil, fmt.Errorf("polymarket: building L1 auth headers: %w", err)
	}

	var creds ApiCredentials
	if err := c.base.getL1(ctx, c.baseURL, "/auth/derive-api-key", nil, l1, &creds); err != nil {
		return nil, err
	}
	return &creds, nil
}

// GetApiKeys returns all API keys for the authenticated user's address.
// Requires L2 (HMAC) authentication.
func (c *ClobClient) GetApiKeys(ctx context.Context) ([]string, error) {
	body, err := c.base.getRaw(ctx, c.baseURL, "/auth/api-keys", nil)
	if err != nil {
		return nil, err
	}

	// API returns {"apiKeys":["key-id-1","key-id-2",...]}.
	var resp struct {
		ApiKeys []string `json:"apiKeys"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("polymarket: decoding api-keys: %w", err)
	}
	return resp.ApiKeys, nil
}

// DeleteApiKey deletes the authenticated user's API key.
// Requires L2 (HMAC) authentication.
func (c *ClobClient) DeleteApiKey(ctx context.Context) error {
	_, err := c.base.deleteRaw(ctx, c.baseURL, "/auth/api-key", nil)
	return err
}

// PostOrderWithOptions submits a signed V2 order to the CLOB. Requires authentication.
//
// The order must be signed using an OrderSigner with the correct EIP-712 signature.
// The orderType controls time-in-force behavior (GTC, GTD, FOK, FAK).
// The legacy IOC constant is translated to FAK. Post-only requires GTC or GTD.
func (c *ClobClient) PostOrderWithOptions(ctx context.Context, signed *SignedOrder, orderType OrderType, opts PostOrderOptions) (*PostOrderResponse, error) {
	if c.base.creds == nil {
		return nil, fmt.Errorf("polymarket: PostOrder requires authentication (use WithCredentials)")
	}
	if signed == nil {
		return nil, fmt.Errorf("polymarket: nil signed order")
	}
	if err := validateV2Order(&signed.Order); err != nil {
		return nil, err
	}
	if orderType == OrderIOC {
		orderType = OrderFAK
	}
	switch orderType {
	case OrderGTC, OrderGTD, OrderFOK, OrderFAK:
	default:
		return nil, fmt.Errorf("polymarket: unsupported order type %q", orderType)
	}
	if opts.PostOnly && orderType != OrderGTC && orderType != OrderGTD {
		return nil, fmt.Errorf("polymarket: post-only requires GTC or GTD")
	}
	metadata, _ := normalizeBytes32(signed.Order.Metadata) // validated above
	builder, _ := normalizeBytes32(signed.Order.Builder)
	expiration := "0"
	if signed.Order.Expiration != nil {
		expiration = signed.Order.Expiration.String()
	}

	req := postOrderRequest{
		Order: postOrderJSON{
			Salt:          json.Number(signed.Order.Salt.String()), // bare JSON number, matching Python client
			Maker:         signed.Order.Maker,
			Signer:        signed.Order.Signer,
			TokenID:       signed.Order.TokenID.String(),
			MakerAmount:   signed.Order.MakerAmount.String(),
			TakerAmount:   signed.Order.TakerAmount.String(),
			Expiration:    expiration,
			Timestamp:     signed.Order.Timestamp.String(),
			Metadata:      metadata,
			Builder:       builder,
			Side:          string(signed.Order.Side),
			SignatureType: int(signed.Order.SignatureType),
			Signature:     signed.Signature,
		},
		Owner:     c.base.creds.Key, // API key, not wallet address (matches Python client)
		OrderType: string(orderType),
		PostOnly:  opts.PostOnly,
		DeferExec: opts.DeferExec,
	}

	var resp PostOrderResponse
	if err := c.base.postJSON(ctx, c.baseURL, "/order", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// PostOrder submits a signed order with default execution options.
func (c *ClobClient) PostOrder(ctx context.Context, signed *SignedOrder, orderType OrderType) (*PostOrderResponse, error) {
	return c.PostOrderWithOptions(ctx, signed, orderType, PostOrderOptions{})
}

// GetBuilderTrades returns trades for the code configured with WithBuilderCode.
// Deprecated: use GetBuilderTradesByCode for the full V2 attribution fields.
func (c *ClobClient) GetBuilderTrades(ctx context.Context, market *string) ([]UserTrade, error) {
	trades, err := c.GetBuilderTradesByCode(ctx, c.builderCode, market)
	if err != nil {
		return nil, err
	}
	out := make([]UserTrade, 0, len(trades))
	for _, tr := range trades {
		out = append(out, UserTrade{ID: tr.ID, Market: tr.Market, AssetID: tr.AssetID,
			Side: tr.Side, Size: tr.Size, Price: tr.Price, Status: tr.Status, Outcome: tr.Outcome,
			Owner: tr.Owner, MakerAddress: tr.Maker, MatchTime: tr.MatchTime,
			TransactionHash: tr.TransactionHash, BucketIndex: tr.BucketIndex,
			Type: tr.TradeType, TakerOrderID: tr.TakerOrderHash})
	}
	return out, nil
}

// GetBuilderTradesByCode retrieves all pages of the public V2 builder feed.
func (c *ClobClient) GetBuilderTradesByCode(ctx context.Context, code string, market *string) ([]BuilderTrade, error) {
	code, err := normalizeBytes32(code)
	if err != nil || code == zeroBytes32 {
		return nil, fmt.Errorf("polymarket: a nonzero bytes32 builder code is required")
	}
	params := url.Values{"builder_code": {code}}
	if market != nil {
		params.Set("market", *market)
	}
	var trades []BuilderTrade
	seen := make(map[string]bool)
	for cursor := "MA=="; cursor != "" && cursor != "LTE="; {
		if seen[cursor] {
			return nil, fmt.Errorf("polymarket: repeated builder trades cursor")
		}
		seen[cursor] = true
		params.Set("next_cursor", cursor)
		var page struct {
			Data       []BuilderTrade `json:"data"`
			NextCursor string         `json:"next_cursor"`
		}
		if err := c.base.get(ctx, c.baseURL, "/builder/trades", params, &page); err != nil {
			return nil, err
		}
		trades = append(trades, page.Data...)
		cursor = page.NextCursor
	}
	return trades, nil
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

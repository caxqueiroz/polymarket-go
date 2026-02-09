# polymarket-go
## Experimental
A Go client library for the [Polymarket](https://polymarket.com) public APIs. Provides read-only access to market data, order books, pricing, events, trades, and positions across all three Polymarket API surfaces.

**Zero external dependencies** — built entirely on the Go standard library.

## Installation

```bash
go get polymarket-go
```

Requires **Go 1.25** or later.

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    polymarket "polymarket-go"
)

func main() {
    client := polymarket.NewClient()
    ctx := context.Background()

    // Search for markets
    results, err := client.Gamma.Search(ctx, "bitcoin")
    if err != nil {
        log.Fatal(err)
    }
    for _, r := range results {
        fmt.Println(r.Title)
    }

    // Get order book depth
    book, err := client.Clob.GetBook(ctx, "TOKEN_ID")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Best bid: %s, Best ask: %s\n", book.Bids[0].Price, book.Asks[0].Price)
}
```

## Architecture

The top-level `Client` provides access to three sub-clients, each targeting a different Polymarket API:

```go
client := polymarket.NewClient()

client.Clob   // Order book, pricing, market parameters
client.Gamma  // Market/event metadata, search, tags
client.Data   // Trades, positions, activity, open interest
```

| Sub-client | Base URL | Focus |
|------------|----------|-------|
| `ClobClient` | `https://clob.polymarket.com` | Order book, prices, spreads, tick sizes |
| `GammaClient` | `https://gamma-api.polymarket.com` | Market/event metadata, search, tags |
| `DataClient` | `https://data-api.polymarket.com` | Trades, positions, activity, holders |

All endpoints are **public and unauthenticated** (read-only market data).

## Configuration

Use functional options to customize the client:

```go
client := polymarket.NewClient(
    polymarket.WithHTTPClient(myHTTPClient),
    polymarket.WithClobBaseURL("https://custom-clob-url.com"),
    polymarket.WithGammaBaseURL("https://custom-gamma-url.com"),
    polymarket.WithDataBaseURL("https://custom-data-url.com"),
)
```

| Option | Description |
|--------|-------------|
| `WithHTTPClient(c)` | Use a custom `*http.Client` (timeouts, transport, retries) |
| `WithClobBaseURL(u)` | Override CLOB API base URL |
| `WithGammaBaseURL(u)` | Override Gamma API base URL |
| `WithDataBaseURL(u)` | Override Data API base URL |

## API Reference

### CLOB Client — Order Book & Pricing

```go
// Health & info
client.Clob.Health(ctx)                                     // error
client.Clob.ServerTime(ctx)                                 // (string, error)

// Markets
client.Clob.GetMarket(ctx, conditionID)                     // (*ClobMarket, error)
client.Clob.ListMarkets(ctx, nextCursor)                    // (*CursorPage[ClobMarket], error)
client.Clob.ListSimplifiedMarkets(ctx, nextCursor)          // (*CursorPage[SimplifiedMarket], error)
client.Clob.AllMarkets(ctx)                                 // *CursorIterator[ClobMarket]

// Single-token pricing
client.Clob.GetPrice(ctx, tokenID, side)                    // (string, error)
client.Clob.GetMidpoint(ctx, tokenID)                       // (string, error)
client.Clob.GetSpread(ctx, tokenID)                         // (string, error)
client.Clob.GetLastTradePrice(ctx, tokenID)                 // (*LastTradePrice, error)

// Bulk pricing (POST)
client.Clob.GetPrices(ctx, []BookParams{...})               // (map[string]map[string]string, error)
client.Clob.GetMidpoints(ctx, []BookParams{...})            // (map[string]string, error)
client.Clob.GetSpreads(ctx, []BookParams{...})              // (map[string]string, error)
client.Clob.GetLastTradesPrices(ctx, []BookParams{...})     // ([]LastTradesPriceEntry, error)

// Order book
client.Clob.GetBook(ctx, tokenID)                           // (*OrderBook, error)
client.Clob.GetBooks(ctx, []BookParams{...})                // ([]OrderBook, error)

// Market info
client.Clob.GetTickSize(ctx, tokenID)                       // (TickSize, error)
client.Clob.GetNegRisk(ctx, tokenID)                        // (bool, error)
client.Clob.GetFeeRateBPS(ctx, tokenID)                     // (float64, error)
client.Clob.GetPriceHistory(ctx, PriceHistoryParams{...})   // (*PriceHistoryResponse, error)
client.Clob.GetMarketTradeEvents(ctx, conditionID)          // ([]MarketTradeEvent, error)
```

### Gamma Client — Market Metadata

```go
// Markets
client.Gamma.ListMarkets(ctx, &GammaMarketParams{...})     // ([]GammaMarket, error)
client.Gamma.GetMarketByID(ctx, id)                         // (*GammaMarket, error)
client.Gamma.GetMarketBySlug(ctx, slug)                     // (*GammaMarket, error)

// Events
client.Gamma.ListEvents(ctx, &GammaEventParams{...})        // ([]GammaEvent, error)
client.Gamma.GetEventByID(ctx, id)                          // (*GammaEvent, error)
client.Gamma.GetEventBySlug(ctx, slug)                      // (*GammaEvent, error)

// Search & tags
client.Gamma.Search(ctx, query)                             // ([]SearchResult, error)
client.Gamma.ListTags(ctx)                                  // ([]GammaTag, error)
client.Gamma.GetTagByID(ctx, id)                            // (*GammaTag, error)
client.Gamma.GetTagBySlug(ctx, slug)                        // (*GammaTag, error)
```

### Data Client — Trades & Positions

```go
client.Data.ListTrades(ctx, &TradesParams{...})             // ([]Trade, error)
client.Data.ListPositions(ctx, &PositionsParams{...})       // ([]Position, error)
client.Data.ListActivity(ctx, &ActivityParams{...})         // ([]Activity, error)
client.Data.ListHolders(ctx, &HoldersParams{...})           // ([]TokenHolders, error)
client.Data.GetOpenInterest(ctx, market)                    // ([]OpenInterest, error)
```

## Pagination

CLOB list endpoints use **cursor-based pagination**. You can page manually or use the iterator:

### Manual paging

```go
page, err := client.Clob.ListMarkets(ctx, "") // first page
for _, m := range page.Data {
    fmt.Println(m.Question)
}

if page.NextCursor != "" {
    page, err = client.Clob.ListMarkets(ctx, page.NextCursor)
}
```

### Iterator (automatic paging)

```go
iter := client.Clob.AllMarkets(ctx)
for iter.Next(ctx) {
    market := iter.Item()
    fmt.Println(market.Question)
}
if err := iter.Err(); err != nil {
    log.Fatal(err)
}
```

Gamma and Data endpoints use **limit/offset pagination** managed via their respective param structs:

```go
limit, offset := 10, 0
markets, err := client.Gamma.ListMarkets(ctx, &polymarket.GammaMarketParams{
    Limit:  &limit,
    Offset: &offset,
})
```

## Error Handling

All API errors are returned as `*polymarket.APIError` with the HTTP status code, status text, and response body:

```go
book, err := client.Clob.GetBook(ctx, "invalid-token")
if err != nil {
    var apiErr *polymarket.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("Status: %d, Body: %s\n", apiErr.StatusCode, apiErr.Body)
    }
}
```

Convenience helpers for common error checks:

```go
if polymarket.IsNotFound(err) {
    // 404 — resource doesn't exist
}

if polymarket.IsRateLimited(err) {
    // 429 — back off and retry
}
```

## Timeouts, Retries & Resilience

The library intentionally does **not** embed timeout, retry, or circuit breaker logic. Instead, use the standard Go approach — pass a configured `*http.Client`:

```go
// Simple timeout
client := polymarket.NewClient(
    polymarket.WithHTTPClient(&http.Client{
        Timeout: 10 * time.Second,
    }),
)

// Per-request timeout via context
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
book, err := client.Clob.GetBook(ctx, tokenID)
```

For retries, circuit breakers, or rate limiting, wrap the HTTP transport:

```go
// Example with hashicorp/go-retryablehttp
retryClient := retryablehttp.NewClient()
retryClient.RetryMax = 3

client := polymarket.NewClient(
    polymarket.WithHTTPClient(retryClient.StandardClient()),
)
```

## Examples

See the [`examples/`](examples/) directory for runnable programs:

| Example | Description |
|---------|-------------|
| [`examples/markets`](examples/markets/main.go) | List active markets with volume and pricing |
| [`examples/prices`](examples/prices/main.go) | Fetch CLOB prices, midpoints, and spreads |
| [`examples/orderbook`](examples/orderbook/main.go) | Display order book depth for a market |

Run an example:

```bash
go run ./examples/markets
```

## Design Notes

- **String prices in CLOB** — The CLOB API returns prices as strings to avoid floating-point precision loss. This library preserves that.
- **`Number` type for Gamma floats** — The Gamma API inconsistently returns numeric fields as either JSON numbers (`123.45`) or JSON strings (`"123.45"`). The `Number` type (underlying `float64`) handles both. Use `float64(m.Liquidity)` or `m.Liquidity.Float64()` when you need a plain float64.
- **Separate market types** — `ClobMarket` and `GammaMarket` are distinct types because the two APIs return different field sets with different JSON naming conventions.
- **`StringSlice` unmarshaler** — The Gamma API inconsistently serializes array fields (like `outcomes`) as either a JSON array `["Yes","No"]` or a JSON string `"[\"Yes\",\"No\"]"`. The custom `StringSlice` type handles both transparently.
- **Pointer fields for optional params** — Query parameter structs use `*int`, `*bool`, etc. to distinguish "not set" from zero values.
- **Generic cursor pagination** — `CursorPage[T]` and `CursorIterator[T]` eliminate boilerplate across CLOB list endpoints.

## Project Structure

```
polymarket-go/
  client.go          # Client, NewClient, functional options, shared HTTP
  errors.go          # APIError, IsNotFound, IsRateLimited
  models.go          # Shared types: Side, Interval, TickSize, BookParams
  pagination.go      # CursorPage[T], CursorIterator[T]
  clob.go            # ClobClient (20 methods)
  clob_models.go     # CLOB request/response types
  gamma.go           # GammaClient (10 methods)
  gamma_models.go    # Gamma types, Number, StringSlice unmarshalers
  data.go            # DataClient (5 methods)
  data_models.go     # Data types
  *_test.go          # Tests (54 total)
  examples/          # Runnable examples
```

## License

MIT

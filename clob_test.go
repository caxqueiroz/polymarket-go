package polymarket

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClobClient(handler http.Handler) *ClobClient {
	srv := httptest.NewServer(handler)
	c := NewClient(
		WithHTTPClient(srv.Client()),
		WithClobBaseURL(srv.URL),
	)
	return c.Clob
}

func newTestClobClientAuth(handler http.Handler) *ClobClient {
	srv := httptest.NewServer(handler)
	c := NewClient(
		WithHTTPClient(srv.Client()),
		WithClobBaseURL(srv.URL),
		WithCredentials(testCreds),
	)
	return c.Clob
}

func TestClobHealth(t *testing.T) {
	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))

	if err := clob.Health(context.Background()); err != nil {
		t.Fatalf("Health() error: %v", err)
	}
}

func TestClobServerTime(t *testing.T) {
	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/time" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte("1700000000"))
	}))

	ts, err := clob.ServerTime(context.Background())
	if err != nil {
		t.Fatalf("ServerTime() error: %v", err)
	}
	if ts != "1700000000" {
		t.Errorf("ServerTime() = %q, want %q", ts, "1700000000")
	}
}

func TestClobGetMarket(t *testing.T) {
	market := ClobMarket{
		ConditionID: "0xabc123",
		Question:    "Will it rain?",
		Active:      true,
		Tokens: []Token{
			{TokenID: "tok1", Outcome: "Yes"},
			{TokenID: "tok2", Outcome: "No"},
		},
	}

	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/markets/0xabc123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(market)
	}))

	got, err := clob.GetMarket(context.Background(), "0xabc123")
	if err != nil {
		t.Fatalf("GetMarket() error: %v", err)
	}
	if got.ConditionID != "0xabc123" {
		t.Errorf("ConditionID = %q, want %q", got.ConditionID, "0xabc123")
	}
	if got.Question != "Will it rain?" {
		t.Errorf("Question = %q, want %q", got.Question, "Will it rain?")
	}
	if len(got.Tokens) != 2 {
		t.Errorf("len(Tokens) = %d, want 2", len(got.Tokens))
	}
}

func TestClobListMarkets(t *testing.T) {
	page := CursorPage[ClobMarket]{
		Data: []ClobMarket{
			{ConditionID: "0x1"},
			{ConditionID: "0x2"},
		},
		NextCursor: "abc123",
	}

	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/markets" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(page)
	}))

	got, err := clob.ListMarkets(context.Background(), "")
	if err != nil {
		t.Fatalf("ListMarkets() error: %v", err)
	}
	if len(got.Data) != 2 {
		t.Errorf("len(Data) = %d, want 2", len(got.Data))
	}
	if got.NextCursor != "abc123" {
		t.Errorf("NextCursor = %q, want %q", got.NextCursor, "abc123")
	}
}

func TestClobGetPrice(t *testing.T) {
	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/price" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("token_id") != "tok1" {
			t.Errorf("token_id = %q, want %q", r.URL.Query().Get("token_id"), "tok1")
		}
		if r.URL.Query().Get("side") != "BUY" {
			t.Errorf("side = %q, want %q", r.URL.Query().Get("side"), "BUY")
		}
		json.NewEncoder(w).Encode(PriceResponse{Price: "0.65"})
	}))

	price, err := clob.GetPrice(context.Background(), "tok1", Buy)
	if err != nil {
		t.Fatalf("GetPrice() error: %v", err)
	}
	if price != "0.65" {
		t.Errorf("GetPrice() = %q, want %q", price, "0.65")
	}
}

func TestClobGetMidpoint(t *testing.T) {
	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/midpoint" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(MidpointResponse{Mid: "0.50"})
	}))

	mid, err := clob.GetMidpoint(context.Background(), "tok1")
	if err != nil {
		t.Fatalf("GetMidpoint() error: %v", err)
	}
	if mid != "0.50" {
		t.Errorf("GetMidpoint() = %q, want %q", mid, "0.50")
	}
}

func TestClobGetSpread(t *testing.T) {
	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(SpreadResponse{Spread: "0.02"})
	}))

	spread, err := clob.GetSpread(context.Background(), "tok1")
	if err != nil {
		t.Fatalf("GetSpread() error: %v", err)
	}
	if spread != "0.02" {
		t.Errorf("GetSpread() = %q, want %q", spread, "0.02")
	}
}

func TestClobGetBook(t *testing.T) {
	book := OrderBook{
		Market:  "0xabc",
		AssetID: "tok1",
		Bids: []OrderLevel{
			{Price: "0.45", Size: "100"},
			{Price: "0.44", Size: "200"},
		},
		Asks: []OrderLevel{
			{Price: "0.55", Size: "150"},
		},
		TickSize:       "0.01",
		LastTradePrice: "0.50",
	}

	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/book" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(book)
	}))

	got, err := clob.GetBook(context.Background(), "tok1")
	if err != nil {
		t.Fatalf("GetBook() error: %v", err)
	}
	if len(got.Bids) != 2 {
		t.Errorf("len(Bids) = %d, want 2", len(got.Bids))
	}
	if len(got.Asks) != 1 {
		t.Errorf("len(Asks) = %d, want 1", len(got.Asks))
	}
	if got.LastTradePrice != "0.50" {
		t.Errorf("LastTradePrice = %q, want %q", got.LastTradePrice, "0.50")
	}
}

func TestClobGetLastTradePrice(t *testing.T) {
	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(LastTradePrice{Price: "0.72", Side: "BUY"})
	}))

	got, err := clob.GetLastTradePrice(context.Background(), "tok1")
	if err != nil {
		t.Fatalf("GetLastTradePrice() error: %v", err)
	}
	if got.Price != "0.72" {
		t.Errorf("Price = %q, want %q", got.Price, "0.72")
	}
	if got.Side != "BUY" {
		t.Errorf("Side = %q, want %q", got.Side, "BUY")
	}
}

func TestClobGetPriceHistory(t *testing.T) {
	resp := PriceHistoryResponse{
		History: []PricePoint{
			{T: 1700000000, P: 0.65},
			{T: 1700003600, P: 0.70},
		},
	}

	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/prices-history" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("market") != "tok1" {
			t.Errorf("market = %q, want %q", r.URL.Query().Get("market"), "tok1")
		}
		if r.URL.Query().Get("interval") != "1h" {
			t.Errorf("interval = %q, want %q", r.URL.Query().Get("interval"), "1h")
		}
		json.NewEncoder(w).Encode(resp)
	}))

	got, err := clob.GetPriceHistory(context.Background(), PriceHistoryParams{
		Market:   "tok1",
		Interval: Interval1h,
	})
	if err != nil {
		t.Fatalf("GetPriceHistory() error: %v", err)
	}
	if len(got.History) != 2 {
		t.Errorf("len(History) = %d, want 2", len(got.History))
	}
}

func TestClobGetMarketTradeEvents(t *testing.T) {
	events := []MarketTradeEvent{
		{
			EventType: "trade",
			Side:      "BUY",
			Size:      "10",
			Price:     "0.65",
			Market:    MarketTradeEventInfo{ConditionID: "0xabc"},
		},
	}

	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/live-activity/events/0xabc" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(events)
	}))

	got, err := clob.GetMarketTradeEvents(context.Background(), "0xabc")
	if err != nil {
		t.Fatalf("GetMarketTradeEvents() error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(events) = %d, want 1", len(got))
	}
	if got[0].Price != "0.65" {
		t.Errorf("Price = %q, want %q", got[0].Price, "0.65")
	}
}

func TestClobGetTickSizeWrapped(t *testing.T) {
	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"minimum_tick_size": "0.01"}`))
	}))

	ts, err := clob.GetTickSize(context.Background(), "tok1")
	if err != nil {
		t.Fatalf("GetTickSize() error: %v", err)
	}
	if ts != TickSize0_01 {
		t.Errorf("GetTickSize() = %q, want %q", ts, TickSize0_01)
	}
}

func TestClobGetTickSizeBare(t *testing.T) {
	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`"0.001"`))
	}))

	ts, err := clob.GetTickSize(context.Background(), "tok1")
	if err != nil {
		t.Fatalf("GetTickSize() error: %v", err)
	}
	if ts != TickSize0_001 {
		t.Errorf("GetTickSize() = %q, want %q", ts, TickSize0_001)
	}
}

func TestClobGetNegRiskWrapped(t *testing.T) {
	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"neg_risk": true}`))
	}))

	nr, err := clob.GetNegRisk(context.Background(), "tok1")
	if err != nil {
		t.Fatalf("GetNegRisk() error: %v", err)
	}
	if !nr {
		t.Errorf("GetNegRisk() = false, want true")
	}
}

func TestClobGetNegRiskBare(t *testing.T) {
	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`true`))
	}))

	nr, err := clob.GetNegRisk(context.Background(), "tok1")
	if err != nil {
		t.Fatalf("GetNegRisk() error: %v", err)
	}
	if !nr {
		t.Errorf("GetNegRisk() = false, want true")
	}
}

func TestClobGetFeeRateBPS(t *testing.T) {
	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`100`))
	}))

	fee, err := clob.GetFeeRateBPS(context.Background(), "tok1")
	if err != nil {
		t.Fatalf("GetFeeRateBPS() error: %v", err)
	}
	if fee != 100 {
		t.Errorf("GetFeeRateBPS() = %f, want 100", fee)
	}
}

func TestClobAPIError(t *testing.T) {
	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}`))
	}))

	_, err := clob.GetMarket(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("IsNotFound() = false, want true for error: %v", err)
	}
}

func TestClobGetPricesPost(t *testing.T) {
	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/prices" {
			t.Errorf("path = %q, want /prices", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var payload []bookParamEntry
		json.Unmarshal(body, &payload)
		if len(payload) != 2 {
			t.Errorf("payload len = %d, want 2", len(payload))
		}
		if payload[0].TokenID != "tok1" {
			t.Errorf("payload[0].TokenID = %q, want %q", payload[0].TokenID, "tok1")
		}
		if payload[0].Side != "BUY" {
			t.Errorf("payload[0].Side = %q, want %q", payload[0].Side, "BUY")
		}

		resp := map[string]map[string]string{
			"tok1": {"BUY": "0.65"},
			"tok2": {"SELL": "0.35"},
		}
		json.NewEncoder(w).Encode(resp)
	}))

	side := Buy
	got, err := clob.GetPrices(context.Background(), []BookParams{
		{TokenID: "tok1", Side: &side},
		{TokenID: "tok2"},
	})
	if err != nil {
		t.Fatalf("GetPrices() error: %v", err)
	}
	if got["tok1"]["BUY"] != "0.65" {
		t.Errorf("tok1 BUY = %q, want %q", got["tok1"]["BUY"], "0.65")
	}
}

func TestClobGetMidpointsPost(t *testing.T) {
	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		resp := map[string]string{
			"tok1": "0.50",
			"tok2": "0.60",
		}
		json.NewEncoder(w).Encode(resp)
	}))

	got, err := clob.GetMidpoints(context.Background(), []BookParams{
		{TokenID: "tok1"},
		{TokenID: "tok2"},
	})
	if err != nil {
		t.Fatalf("GetMidpoints() error: %v", err)
	}
	if got["tok1"] != "0.50" {
		t.Errorf("tok1 = %q, want %q", got["tok1"], "0.50")
	}
}

func TestClobGetSpreadsPost(t *testing.T) {
	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		resp := map[string]string{"tok1": "0.05"}
		json.NewEncoder(w).Encode(resp)
	}))

	got, err := clob.GetSpreads(context.Background(), []BookParams{
		{TokenID: "tok1"},
	})
	if err != nil {
		t.Fatalf("GetSpreads() error: %v", err)
	}
	if got["tok1"] != "0.05" {
		t.Errorf("tok1 = %q, want %q", got["tok1"], "0.05")
	}
}

func TestClobGetBooksPost(t *testing.T) {
	books := []OrderBook{
		{Market: "0x1", AssetID: "tok1", Bids: []OrderLevel{{Price: "0.5", Size: "10"}}},
		{Market: "0x2", AssetID: "tok2", Asks: []OrderLevel{{Price: "0.6", Size: "20"}}},
	}

	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		json.NewEncoder(w).Encode(books)
	}))

	got, err := clob.GetBooks(context.Background(), []BookParams{
		{TokenID: "tok1"},
		{TokenID: "tok2"},
	})
	if err != nil {
		t.Fatalf("GetBooks() error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len(books) = %d, want 2", len(got))
	}
}

func TestClobGetLastTradesPricesPost(t *testing.T) {
	entries := []LastTradesPriceEntry{
		{Price: "0.55", Side: "BUY", TokenID: "tok1"},
	}

	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		json.NewEncoder(w).Encode(entries)
	}))

	got, err := clob.GetLastTradesPrices(context.Background(), []BookParams{
		{TokenID: "tok1"},
	})
	if err != nil {
		t.Fatalf("GetLastTradesPrices() error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Price != "0.55" {
		t.Errorf("Price = %q, want %q", got[0].Price, "0.55")
	}
}

func TestFlexStringUnmarshal(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "json string", input: `"0.01"`, want: "0.01"},
		{name: "json number float", input: `0.01`, want: "0.01"},
		{name: "json number int", input: `15`, want: "15"},
		{name: "json zero", input: `0`, want: "0"},
		{name: "empty string", input: `""`, want: ""},
		{name: "null", input: `null`, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s FlexString
			if err := json.Unmarshal([]byte(tt.input), &s); err != nil {
				t.Fatalf("Unmarshal() error: %v", err)
			}
			if string(s) != tt.want {
				t.Errorf("got %q, want %q", string(s), tt.want)
			}
		})
	}
}

func TestClobMarketNumericFields(t *testing.T) {
	// Real CLOB API returns minimum_order_size and minimum_tick_size as numbers,
	// and rewards.min_size/max_spread as numbers.
	input := `{
		"condition_id": "0xabc",
		"question": "Test market",
		"minimum_order_size": 15,
		"minimum_tick_size": 0.01,
		"maker_base_fee": 0,
		"taker_base_fee": 0,
		"rewards": {
			"min_size": 0,
			"max_spread": 0,
			"rates": null
		},
		"tokens": [
			{"token_id": "tok1", "outcome": "Yes", "price": 0.65, "winner": false}
		]
	}`

	var m ClobMarket
	if err := json.Unmarshal([]byte(input), &m); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}
	if string(m.MinimumOrderSize) != "15" {
		t.Errorf("MinimumOrderSize = %q, want %q", string(m.MinimumOrderSize), "15")
	}
	if string(m.MinimumTickSize) != "0.01" {
		t.Errorf("MinimumTickSize = %q, want %q", string(m.MinimumTickSize), "0.01")
	}
	if string(m.Rewards.MinSize) != "0" {
		t.Errorf("Rewards.MinSize = %q, want %q", string(m.Rewards.MinSize), "0")
	}
}

// requireAuthHeaders verifies all 5 POLY_* headers are present.
func requireAuthHeaders(t *testing.T, r *http.Request) {
	t.Helper()
	for _, h := range []string{"POLY_API_KEY", "POLY_SIGNATURE", "POLY_TIMESTAMP", "POLY_PASSPHRASE", "POLY_ADDRESS"} {
		if r.Header.Get(h) == "" {
			t.Errorf("missing auth header %s", h)
		}
	}
}

func TestClobGetOrder(t *testing.T) {
	order := Order{
		ID:      "order-123",
		Market:  "0xabc",
		AssetID: "tok1",
		Side:    "BUY",
		Price:   "0.65",
		Status:  "LIVE",
	}

	clob := newTestClobClientAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireAuthHeaders(t, r)
		if r.URL.Path != "/data/order/order-123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(order)
	}))

	got, err := clob.GetOrder(context.Background(), "order-123")
	if err != nil {
		t.Fatalf("GetOrder() error: %v", err)
	}
	if got.ID != "order-123" {
		t.Errorf("ID = %q, want %q", got.ID, "order-123")
	}
	if got.Price != "0.65" {
		t.Errorf("Price = %q, want %q", got.Price, "0.65")
	}
}

func TestClobGetOrders(t *testing.T) {
	clob := newTestClobClientAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireAuthHeaders(t, r)
		if r.URL.Path != "/data/orders" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if m := r.URL.Query().Get("market"); m != "0xabc" {
			t.Errorf("market = %q, want %q", m, "0xabc")
		}
		// Return paginated response matching Polymarket API format
		json.NewEncoder(w).Encode(map[string]any{
			"data": []OpenOrder{
				{ID: "o1", Market: "0xabc", Side: "BUY", Price: "0.65"},
				{ID: "o2", Market: "0xabc", Side: "SELL", Price: "0.70"},
			},
			"next_cursor": "LTE=", // END_CURSOR
		})
	}))

	market := "0xabc"
	got, err := clob.GetOrders(context.Background(), &OrdersParams{Market: &market})
	if err != nil {
		t.Fatalf("GetOrders() error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].ID != "o1" {
		t.Errorf("ID = %q, want %q", got[0].ID, "o1")
	}
}

func TestClobGetOrdersNilParams(t *testing.T) {
	clob := newTestClobClientAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireAuthHeaders(t, r)
		json.NewEncoder(w).Encode(map[string]any{
			"data":        []OpenOrder{},
			"next_cursor": "LTE=",
		})
	}))

	got, err := clob.GetOrders(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetOrders() error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}

func TestClobCancelOrder(t *testing.T) {
	clob := newTestClobClientAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireAuthHeaders(t, r)
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		if r.URL.Path != "/order/order-456" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))

	if err := clob.CancelOrder(context.Background(), "order-456"); err != nil {
		t.Fatalf("CancelOrder() error: %v", err)
	}
}

func TestClobCancelOrders(t *testing.T) {
	clob := newTestClobClientAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireAuthHeaders(t, r)
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		if r.URL.Path != "/orders" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var payload CancelOrdersPayload
		json.Unmarshal(body, &payload)
		if len(payload.OrderIDs) != 2 {
			t.Errorf("len(OrderIDs) = %d, want 2", len(payload.OrderIDs))
		}

		json.NewEncoder(w).Encode([]string{"o1", "o2"})
	}))

	got, err := clob.CancelOrders(context.Background(), []string{"o1", "o2"})
	if err != nil {
		t.Fatalf("CancelOrders() error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
}

func TestClobCancelAll(t *testing.T) {
	clob := newTestClobClientAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireAuthHeaders(t, r)
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		if r.URL.Path != "/cancel-all" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode([]string{"o1", "o2", "o3"})
	}))

	got, err := clob.CancelAll(context.Background())
	if err != nil {
		t.Fatalf("CancelAll() error: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("len = %d, want 3", len(got))
	}
}

func TestClobGetTrades(t *testing.T) {
	clob := newTestClobClientAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireAuthHeaders(t, r)
		if r.URL.Path != "/data/trades" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"data": []UserTrade{
				{ID: "t1", Market: "0xabc", Side: "BUY", Price: "0.65", Size: "10"},
			},
			"next_cursor": "LTE=",
		})
	}))

	got, err := clob.GetTrades(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetTrades() error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Price != "0.65" {
		t.Errorf("Price = %q, want %q", got[0].Price, "0.65")
	}
}

func TestClobGetBalanceAllowance(t *testing.T) {
	ba := BalanceAllowance{Balance: "1000.00", Allowance: "5000.00"}

	clob := newTestClobClientAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireAuthHeaders(t, r)
		if r.URL.Path != "/balance-allowance" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(ba)
	}))

	got, err := clob.GetBalanceAllowance(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetBalanceAllowance() error: %v", err)
	}
	if got.Balance != "1000.00" {
		t.Errorf("Balance = %q, want %q", got.Balance, "1000.00")
	}
	if got.Allowance != "5000.00" {
		t.Errorf("Allowance = %q, want %q", got.Allowance, "5000.00")
	}
}

func TestClobGetNotifications(t *testing.T) {
	notifs := []Notification{
		{Type: "order_filled", Payload: "order-123", Timestamp: "1700000000"},
	}

	clob := newTestClobClientAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireAuthHeaders(t, r)
		if r.URL.Path != "/notifications" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(notifs)
	}))

	got, err := clob.GetNotifications(context.Background())
	if err != nil {
		t.Fatalf("GetNotifications() error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Type != "order_filled" {
		t.Errorf("Type = %q, want %q", got[0].Type, "order_filled")
	}
}

func TestClobNoAuthHeadersWithoutCredentials(t *testing.T) {
	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, h := range []string{"POLY_API_KEY", "POLY_SIGNATURE", "POLY_TIMESTAMP", "POLY_PASSPHRASE", "POLY_ADDRESS"} {
			if r.Header.Get(h) != "" {
				t.Errorf("header %s should be empty without credentials, got %q", h, r.Header.Get(h))
			}
		}
		w.WriteHeader(http.StatusOK)
	}))

	clob.Health(context.Background())
}

// ---- L1 Auth API Key Tests ----

// requireL1AuthHeaders checks that L1 (EIP-712) authentication headers are present.
func requireL1AuthHeaders(t *testing.T, r *http.Request) {
	t.Helper()
	for _, h := range []string{"POLY_ADDRESS", "POLY_SIGNATURE", "POLY_TIMESTAMP", "POLY_NONCE"} {
		if r.Header.Get(h) == "" {
			t.Errorf("missing L1 auth header %s", h)
		}
	}
}

func TestClobCreateOrDeriveApiCreds(t *testing.T) {
	t.Run("create succeeds", func(t *testing.T) {
		clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// CreateOrDeriveApiCreds first tries POST /auth/api-key (create)
			if r.Method != http.MethodPost {
				t.Errorf("method = %q, want POST", r.Method)
			}
			if r.URL.Path != "/auth/api-key" {
				t.Errorf("path = %q, want /auth/api-key", r.URL.Path)
			}
			requireL1AuthHeaders(t, r)

			addr := r.Header.Get("POLY_ADDRESS")
			if addr != "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266" {
				t.Errorf("POLY_ADDRESS = %q, want 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266", addr)
			}

			sig := r.Header.Get("POLY_SIGNATURE")
			if len(sig) != 132 {
				t.Errorf("POLY_SIGNATURE length = %d, want 132", len(sig))
			}

			w.Write([]byte(`{
				"apiKey": "test-api-key-uuid",
				"secret": "dGVzdC1zZWNyZXQtYmFzZTY0",
				"passphrase": "test-passphrase"
			}`))
		}))

		signer, _ := NewOrderSigner(testPrivateKey, Polygon)
		creds, err := clob.CreateOrDeriveApiCreds(context.Background(), signer)
		if err != nil {
			t.Fatalf("CreateOrDeriveApiCreds() error: %v", err)
		}
		if creds.ApiKey != "test-api-key-uuid" {
			t.Errorf("ApiKey = %q, want %q", creds.ApiKey, "test-api-key-uuid")
		}
		if creds.Secret != "dGVzdC1zZWNyZXQtYmFzZTY0" {
			t.Errorf("Secret = %q, want %q", creds.Secret, "dGVzdC1zZWNyZXQtYmFzZTY0")
		}
		if creds.Passphrase != "test-passphrase" {
			t.Errorf("Passphrase = %q, want %q", creds.Passphrase, "test-passphrase")
		}
	})

	t.Run("create fails then derive succeeds", func(t *testing.T) {
		calls := 0
		clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			if calls == 1 {
				// First call: CreateApiKey → POST /auth/api-key → 409 (conflict)
				if r.Method != http.MethodPost {
					t.Errorf("call 1: method = %q, want POST", r.Method)
				}
				if r.URL.Path != "/auth/api-key" {
					t.Errorf("call 1: path = %q, want /auth/api-key", r.URL.Path)
				}
				requireL1AuthHeaders(t, r)
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte(`{"error":"key already exists"}`))
				return
			}
			// Second call: DeriveApiKey → GET /auth/derive-api-key
			if r.Method != http.MethodGet {
				t.Errorf("call 2: method = %q, want GET", r.Method)
			}
			if r.URL.Path != "/auth/derive-api-key" {
				t.Errorf("call 2: path = %q, want /auth/derive-api-key", r.URL.Path)
			}
			requireL1AuthHeaders(t, r)
			w.Write([]byte(`{
				"apiKey": "derived-key",
				"secret": "ZGVyaXZlZC1zZWNyZXQ=",
				"passphrase": "derived-pass"
			}`))
		}))

		signer, _ := NewOrderSigner(testPrivateKey, Polygon)
		creds, err := clob.CreateOrDeriveApiCreds(context.Background(), signer)
		if err != nil {
			t.Fatalf("CreateOrDeriveApiCreds() error: %v", err)
		}
		if calls != 2 {
			t.Errorf("calls = %d, want 2 (create + derive)", calls)
		}
		if creds.ApiKey != "derived-key" {
			t.Errorf("ApiKey = %q, want %q", creds.ApiKey, "derived-key")
		}
	})
}

func TestClobCreateApiKey(t *testing.T) {
	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/auth/api-key" {
			t.Errorf("path = %q, want /auth/api-key", r.URL.Path)
		}
		requireL1AuthHeaders(t, r)

		w.Write([]byte(`{
			"apiKey": "new-api-key",
			"secret": "bmV3LXNlY3JldA==",
			"passphrase": "new-pass"
		}`))
	}))

	signer, _ := NewOrderSigner(testPrivateKey, Polygon)
	creds, err := clob.CreateApiKey(context.Background(), signer)
	if err != nil {
		t.Fatalf("CreateApiKey() error: %v", err)
	}
	if creds.ApiKey != "new-api-key" {
		t.Errorf("ApiKey = %q, want %q", creds.ApiKey, "new-api-key")
	}
}

func TestClobDeriveApiKey(t *testing.T) {
	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/auth/derive-api-key" {
			t.Errorf("path = %q, want /auth/derive-api-key", r.URL.Path)
		}
		requireL1AuthHeaders(t, r)

		w.Write([]byte(`{
			"apiKey": "derived-key",
			"secret": "ZGVyaXZlZC1zZWNyZXQ=",
			"passphrase": "derived-pass"
		}`))
	}))

	signer, _ := NewOrderSigner(testPrivateKey, Polygon)
	creds, err := clob.DeriveApiKey(context.Background(), signer)
	if err != nil {
		t.Fatalf("DeriveApiKey() error: %v", err)
	}
	if creds.ApiKey != "derived-key" {
		t.Errorf("ApiKey = %q, want %q", creds.ApiKey, "derived-key")
	}
	if creds.Secret != "ZGVyaXZlZC1zZWNyZXQ=" {
		t.Errorf("Secret = %q, want %q", creds.Secret, "ZGVyaXZlZC1zZWNyZXQ=")
	}
}

func TestClobGetApiKeys(t *testing.T) {
	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/auth/api-keys" {
			t.Errorf("path = %q, want /auth/api-keys", r.URL.Path)
		}
		// GetApiKeys uses L2 (HMAC) auth, not L1
		w.Write([]byte(`{"apiKeys":["key-1","key-2"]}`))
	}))

	keys, err := clob.GetApiKeys(context.Background())
	if err != nil {
		t.Fatalf("GetApiKeys() error: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("len(keys) = %d, want 2", len(keys))
	}
	if keys[0] != "key-1" {
		t.Errorf("keys[0] = %q, want %q", keys[0], "key-1")
	}
	if keys[1] != "key-2" {
		t.Errorf("keys[1] = %q, want %q", keys[1], "key-2")
	}
}

func TestClobDeleteApiKey(t *testing.T) {
	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", r.Method)
		}
		if r.URL.Path != "/auth/api-key" {
			t.Errorf("path = %q, want /auth/api-key", r.URL.Path)
		}
		// DeleteApiKey uses L2 (HMAC) auth, not L1

		w.WriteHeader(http.StatusOK)
	}))

	err := clob.DeleteApiKey(context.Background())
	if err != nil {
		t.Fatalf("DeleteApiKey() error: %v", err)
	}
}

func TestApiCredentialsToCredentials(t *testing.T) {
	creds := &ApiCredentials{
		ApiKey:     "test-key",
		Secret:     "test-secret",
		Passphrase: "test-pass",
	}
	addr := "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"

	c := creds.ToCredentials(addr)

	if c.Key != "test-key" {
		t.Errorf("Key = %q, want %q", c.Key, "test-key")
	}
	if c.Secret != "test-secret" {
		t.Errorf("Secret = %q, want %q", c.Secret, "test-secret")
	}
	if c.Passphrase != "test-pass" {
		t.Errorf("Passphrase = %q, want %q", c.Passphrase, "test-pass")
	}
	if c.Address != addr {
		t.Errorf("Address = %q, want %q", c.Address, addr)
	}
}

func TestClobL1AuthSignatureRecoverable(t *testing.T) {
	// End-to-end test: build L1 headers and verify the signature can be recovered
	signer, _ := NewOrderSigner(testPrivateKey, Polygon)

	var capturedAddr, capturedSig, capturedTs, capturedNonce string

	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAddr = r.Header.Get("POLY_ADDRESS")
		capturedSig = r.Header.Get("POLY_SIGNATURE")
		capturedTs = r.Header.Get("POLY_TIMESTAMP")
		capturedNonce = r.Header.Get("POLY_NONCE")

		w.Write([]byte(`{"apiKey":"k","secret":"s","passphrase":"p"}`))
	}))

	_, err := clob.CreateOrDeriveApiCreds(context.Background(), signer)
	if err != nil {
		t.Fatal(err)
	}

	// Verify the captured signature is recoverable
	if capturedAddr != signer.Address() {
		t.Errorf("captured address = %s, want %s", capturedAddr, signer.Address())
	}

	sigBytes, _ := hex.DecodeString(capturedSig[2:])
	if len(sigBytes) != 65 {
		t.Fatalf("sig bytes length = %d, want 65", len(sigBytes))
	}

	v := sigBytes[64]
	var r2, s2 [32]byte
	copy(r2[:], sigBytes[:32])
	copy(s2[:], sigBytes[32:64])

	digest := clobAuthDigest(capturedAddr, capturedTs, 0, 137)
	recovered, err := ecRecover(digest[:], r2, s2, v-27)
	if err != nil {
		t.Fatalf("ecRecover error: %v", err)
	}
	if recovered != signer.Address() {
		t.Errorf("recovered address = %s, want %s", recovered, signer.Address())
	}

	if capturedNonce != "0" {
		t.Errorf("nonce = %s, want 0", capturedNonce)
	}
}

func TestClobMarketCursorPageNumericFields(t *testing.T) {
	// Simulates the actual /markets response that was failing
	clob := newTestClobClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{
			"data": [{
				"condition_id": "0xabc",
				"question": "Test",
				"minimum_order_size": 15,
				"minimum_tick_size": 0.01,
				"maker_base_fee": 0,
				"taker_base_fee": 0,
				"tokens": [],
				"rewards": {"min_size": 0, "max_spread": 0, "rates": null}
			}],
			"next_cursor": ""
		}`))
	}))

	page, err := clob.ListMarkets(context.Background(), "")
	if err != nil {
		t.Fatalf("ListMarkets() error: %v", err)
	}
	if len(page.Data) != 1 {
		t.Fatalf("len(Data) = %d, want 1", len(page.Data))
	}
	if string(page.Data[0].MinimumOrderSize) != "15" {
		t.Errorf("MinimumOrderSize = %q, want %q", string(page.Data[0].MinimumOrderSize), "15")
	}
}

func newTestClobClientAuthWithBuilder(handler http.Handler) *ClobClient {
	srv := httptest.NewServer(handler)
	c := NewClient(
		WithHTTPClient(srv.Client()),
		WithClobBaseURL(srv.URL),
		WithCredentials(testCreds),
		WithBuilderCredentials(testBuilderCreds),
	)
	return c.Clob
}

func requireBuilderHeaders(t *testing.T, r *http.Request) {
	t.Helper()
	for _, h := range []string{"POLY_BUILDER_API_KEY", "POLY_BUILDER_SIGNATURE", "POLY_BUILDER_TIMESTAMP", "POLY_BUILDER_PASSPHRASE"} {
		if r.Header.Get(h) == "" {
			t.Errorf("missing builder header %s", h)
		}
	}
}

func requireNoBuilderHeaders(t *testing.T, r *http.Request) {
	t.Helper()
	for _, h := range []string{"POLY_BUILDER_API_KEY", "POLY_BUILDER_SIGNATURE", "POLY_BUILDER_TIMESTAMP", "POLY_BUILDER_PASSPHRASE"} {
		if r.Header.Get(h) != "" {
			t.Errorf("unexpected builder header %s = %q", h, r.Header.Get(h))
		}
	}
}

func makeTestOrder() OrderData {
	return OrderData{
		Salt:          big.NewInt(123),
		Maker:         "0xMaker",
		Signer:        "0xSigner",
		Taker:         "0x0000000000000000000000000000000000000000",
		TokenID:       big.NewInt(456),
		MakerAmount:   big.NewInt(100),
		TakerAmount:   big.NewInt(50),
		Expiration:    big.NewInt(0),
		Nonce:         big.NewInt(0),
		FeeRateBPS:    big.NewInt(0),
		Side:          Buy,
		SignatureType: SignatureTypeEOA,
	}
}

func TestPostOrderWithBuilderHeaders(t *testing.T) {
	clob := newTestClobClientAuthWithBuilder(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireAuthHeaders(t, r)
		requireBuilderHeaders(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success":true,"orderID":"order-123"}`))
	}))

	signed := &SignedOrder{
		Order:     makeTestOrder(),
		Signature: "0xdeadbeef",
	}
	resp, err := clob.PostOrder(context.Background(), signed, OrderGTC)
	if err != nil {
		t.Fatalf("PostOrder() error: %v", err)
	}
	if !resp.Success {
		t.Error("expected success")
	}
}

func TestPostOrderWithoutBuilderHeaders(t *testing.T) {
	clob := newTestClobClientAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requireAuthHeaders(t, r)
		requireNoBuilderHeaders(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success":true,"orderID":"order-456"}`))
	}))

	signed := &SignedOrder{
		Order:     makeTestOrder(),
		Signature: "0xdeadbeef",
	}
	_, err := clob.PostOrder(context.Background(), signed, OrderGTC)
	if err != nil {
		t.Fatalf("PostOrder() error: %v", err)
	}
}

package polymarket

import (
	"context"
	"encoding/json"
	"io"
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

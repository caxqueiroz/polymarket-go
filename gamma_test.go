package polymarket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestGammaClient(handler http.Handler) *GammaClient {
	srv := httptest.NewServer(handler)
	c := NewClient(
		WithHTTPClient(srv.Client()),
		WithGammaBaseURL(srv.URL),
	)
	return c.Gamma
}

func TestGammaListMarkets(t *testing.T) {
	markets := []GammaMarket{
		{ID: "1", Question: "Market 1", Active: true},
		{ID: "2", Question: "Market 2", Active: false},
	}

	gamma := newTestGammaClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/markets" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(markets)
	}))

	got, err := gamma.ListMarkets(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListMarkets() error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len(markets) = %d, want 2", len(got))
	}
}

func TestGammaListMarketsWithParams(t *testing.T) {
	gamma := newTestGammaClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "5" {
			t.Errorf("limit = %q, want %q", r.URL.Query().Get("limit"), "5")
		}
		if r.URL.Query().Get("active") != "true" {
			t.Errorf("active = %q, want %q", r.URL.Query().Get("active"), "true")
		}
		json.NewEncoder(w).Encode([]GammaMarket{})
	}))

	limit := 5
	active := true
	_, err := gamma.ListMarkets(context.Background(), &GammaMarketParams{
		Limit:  &limit,
		Active: &active,
	})
	if err != nil {
		t.Fatalf("ListMarkets() error: %v", err)
	}
}

func TestGammaGetMarketByID(t *testing.T) {
	gamma := newTestGammaClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]GammaMarket{
			{ID: "42", Question: "Found it"},
		})
	}))

	got, err := gamma.GetMarketByID(context.Background(), "42")
	if err != nil {
		t.Fatalf("GetMarketByID() error: %v", err)
	}
	if got.ID != "42" {
		t.Errorf("ID = %q, want %q", got.ID, "42")
	}
}

func TestGammaGetMarketByIDNotFound(t *testing.T) {
	gamma := newTestGammaClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]GammaMarket{})
	}))

	_, err := gamma.GetMarketByID(context.Background(), "999")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("IsNotFound() = false, want true")
	}
}

func TestGammaGetMarketBySlug(t *testing.T) {
	gamma := newTestGammaClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("slug") != "will-it-rain" {
			t.Errorf("slug = %q, want %q", r.URL.Query().Get("slug"), "will-it-rain")
		}
		json.NewEncoder(w).Encode([]GammaMarket{
			{ID: "1", Slug: "will-it-rain"},
		})
	}))

	got, err := gamma.GetMarketBySlug(context.Background(), "will-it-rain")
	if err != nil {
		t.Fatalf("GetMarketBySlug() error: %v", err)
	}
	if got.Slug != "will-it-rain" {
		t.Errorf("Slug = %q, want %q", got.Slug, "will-it-rain")
	}
}

func TestGammaListEvents(t *testing.T) {
	events := []GammaEvent{
		{ID: "1", Title: "Event 1"},
	}

	gamma := newTestGammaClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/events" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(events)
	}))

	got, err := gamma.ListEvents(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListEvents() error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len(events) = %d, want 1", len(got))
	}
}

func TestGammaGetEventByID(t *testing.T) {
	event := GammaEvent{ID: "10", Title: "My Event"}

	gamma := newTestGammaClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/events/10" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(event)
	}))

	got, err := gamma.GetEventByID(context.Background(), "10")
	if err != nil {
		t.Fatalf("GetEventByID() error: %v", err)
	}
	if got.Title != "My Event" {
		t.Errorf("Title = %q, want %q", got.Title, "My Event")
	}
}

func TestGammaGetEventBySlug(t *testing.T) {
	gamma := newTestGammaClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]GammaEvent{
			{ID: "1", Slug: "rain-event"},
		})
	}))

	got, err := gamma.GetEventBySlug(context.Background(), "rain-event")
	if err != nil {
		t.Fatalf("GetEventBySlug() error: %v", err)
	}
	if got.Slug != "rain-event" {
		t.Errorf("Slug = %q, want %q", got.Slug, "rain-event")
	}
}

func TestGammaSearch(t *testing.T) {
	results := []SearchResult{
		{ID: "1", Title: "Bitcoin prediction"},
	}

	gamma := newTestGammaClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/public-search" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("q") != "bitcoin" {
			t.Errorf("q = %q, want %q", r.URL.Query().Get("q"), "bitcoin")
		}
		json.NewEncoder(w).Encode(results)
	}))

	got, err := gamma.Search(context.Background(), "bitcoin")
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len(results) = %d, want 1", len(got))
	}
}

func TestGammaListTags(t *testing.T) {
	tags := []GammaTag{
		{ID: "1", Label: "Politics", Slug: "politics"},
		{ID: "2", Label: "Sports", Slug: "sports"},
	}

	gamma := newTestGammaClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(tags)
	}))

	got, err := gamma.ListTags(context.Background())
	if err != nil {
		t.Fatalf("ListTags() error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len(tags) = %d, want 2", len(got))
	}
}

func TestGammaGetTagByID(t *testing.T) {
	gamma := newTestGammaClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tags/1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(GammaTag{ID: "1", Label: "Politics"})
	}))

	got, err := gamma.GetTagByID(context.Background(), "1")
	if err != nil {
		t.Fatalf("GetTagByID() error: %v", err)
	}
	if got.Label != "Politics" {
		t.Errorf("Label = %q, want %q", got.Label, "Politics")
	}
}

func TestGammaGetTagBySlug(t *testing.T) {
	gamma := newTestGammaClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]GammaTag{
			{ID: "1", Slug: "politics", Label: "Politics"},
		})
	}))

	got, err := gamma.GetTagBySlug(context.Background(), "politics")
	if err != nil {
		t.Fatalf("GetTagBySlug() error: %v", err)
	}
	if got.Slug != "politics" {
		t.Errorf("Slug = %q, want %q", got.Slug, "politics")
	}
}

func TestStringSliceUnmarshal(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "json array",
			input: `["Yes","No"]`,
			want:  []string{"Yes", "No"},
		},
		{
			name:  "string containing json array",
			input: `"[\"Yes\",\"No\"]"`,
			want:  []string{"Yes", "No"},
		},
		{
			name:  "empty array",
			input: `[]`,
			want:  nil,
		},
		{
			name:  "empty string",
			input: `""`,
			want:  nil,
		},
		{
			name:  "string containing empty array",
			input: `"[]"`,
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s StringSlice
			if err := json.Unmarshal([]byte(tt.input), &s); err != nil {
				t.Fatalf("Unmarshal() error: %v", err)
			}
			if len(s) != len(tt.want) {
				t.Errorf("len = %d, want %d", len(s), len(tt.want))
				return
			}
			for i := range s {
				if s[i] != tt.want[i] {
					t.Errorf("s[%d] = %q, want %q", i, s[i], tt.want[i])
				}
			}
		})
	}
}

func TestStringSliceInStruct(t *testing.T) {
	// Test that StringSlice works correctly when embedded in a struct
	input := `{
		"id": "1",
		"outcomes": "[\"Yes\",\"No\"]",
		"outcomePrices": ["0.65", "0.35"],
		"clobTokenIds": "[\"tok1\",\"tok2\"]"
	}`

	var m GammaMarket
	if err := json.Unmarshal([]byte(input), &m); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}
	if len(m.Outcomes) != 2 {
		t.Errorf("len(Outcomes) = %d, want 2", len(m.Outcomes))
	}
	if len(m.OutcomePrices) != 2 {
		t.Errorf("len(OutcomePrices) = %d, want 2", len(m.OutcomePrices))
	}
	if m.Outcomes[0] != "Yes" {
		t.Errorf("Outcomes[0] = %q, want %q", m.Outcomes[0], "Yes")
	}
}

func TestNumberUnmarshal(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  float64
	}{
		{name: "json number", input: `123.45`, want: 123.45},
		{name: "json string", input: `"123.45"`, want: 123.45},
		{name: "json integer", input: `100`, want: 100},
		{name: "json string integer", input: `"100"`, want: 100},
		{name: "empty string", input: `""`, want: 0},
		{name: "null", input: `null`, want: 0},
		{name: "zero", input: `0`, want: 0},
		{name: "zero string", input: `"0"`, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var n Number
			if err := json.Unmarshal([]byte(tt.input), &n); err != nil {
				t.Fatalf("Unmarshal() error: %v", err)
			}
			if float64(n) != tt.want {
				t.Errorf("got %f, want %f", float64(n), tt.want)
			}
		})
	}
}

func TestGammaMarketStringNumericFields(t *testing.T) {
	// Gamma API returns numeric fields as strings
	input := `{
		"id": "123",
		"question": "Will BTC hit 100k?",
		"liquidity": "50000.50",
		"volume": "1234567.89",
		"volume24hr": "45000",
		"openInterest": "200000",
		"bestBid": "0.65",
		"bestAsk": "0.67",
		"spread": "0.02",
		"outcomes": "[\"Yes\",\"No\"]",
		"outcomePrices": "[\"0.65\",\"0.35\"]"
	}`

	var m GammaMarket
	if err := json.Unmarshal([]byte(input), &m); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}
	if m.Liquidity != 50000.50 {
		t.Errorf("Liquidity = %f, want 50000.50", float64(m.Liquidity))
	}
	if m.Volume != 1234567.89 {
		t.Errorf("Volume = %f, want 1234567.89", float64(m.Volume))
	}
	if m.BestBid != 0.65 {
		t.Errorf("BestBid = %f, want 0.65", float64(m.BestBid))
	}
}

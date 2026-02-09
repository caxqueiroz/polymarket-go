package polymarket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestDataClient(handler http.Handler) *DataClient {
	srv := httptest.NewServer(handler)
	c := NewClient(
		WithHTTPClient(srv.Client()),
		WithDataBaseURL(srv.URL),
	)
	return c.Data
}

func TestDataListTrades(t *testing.T) {
	trades := []Trade{
		{
			ProxyWallet: "0xuser1",
			Side:        "BUY",
			ConditionID: "0xmarket1",
			Size:        100,
			Price:       0.65,
			Timestamp:   1700000000,
			Title:       "Will it rain?",
			Outcome:     "Yes",
		},
	}

	data := newTestDataClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/trades" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(trades)
	}))

	got, err := data.ListTrades(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTrades() error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len(trades) = %d, want 1", len(got))
	}
	if got[0].Side != "BUY" {
		t.Errorf("Side = %q, want %q", got[0].Side, "BUY")
	}
}

func TestDataListTradesWithParams(t *testing.T) {
	data := newTestDataClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("user") != "0xuser1" {
			t.Errorf("user = %q, want %q", r.URL.Query().Get("user"), "0xuser1")
		}
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("limit = %q, want %q", r.URL.Query().Get("limit"), "10")
		}
		if r.URL.Query().Get("market") != "0xmarket1" {
			t.Errorf("market = %q, want %q", r.URL.Query().Get("market"), "0xmarket1")
		}
		json.NewEncoder(w).Encode([]Trade{})
	}))

	limit := 10
	_, err := data.ListTrades(context.Background(), &TradesParams{
		User:   "0xuser1",
		Market: "0xmarket1",
		Limit:  &limit,
	})
	if err != nil {
		t.Fatalf("ListTrades() error: %v", err)
	}
}

func TestDataListPositions(t *testing.T) {
	positions := []Position{
		{
			ProxyWallet:  "0xuser1",
			ConditionID:  "0xmarket1",
			Size:         50,
			AvgPrice:     0.60,
			CurrentValue: 35.0,
			CashPnl:      5.0,
			Title:        "Will it rain?",
			Outcome:      "Yes",
		},
	}

	data := newTestDataClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/positions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(positions)
	}))

	got, err := data.ListPositions(context.Background(), &PositionsParams{
		User: "0xuser1",
	})
	if err != nil {
		t.Fatalf("ListPositions() error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len(positions) = %d, want 1", len(got))
	}
	if got[0].Size != 50 {
		t.Errorf("Size = %f, want 50", got[0].Size)
	}
}

func TestDataListActivity(t *testing.T) {
	activity := []Activity{
		{
			ProxyWallet: "0xuser1",
			Timestamp:   1700000000,
			Type:        "TRADE",
			Size:        10,
			Side:        "BUY",
			Title:       "Will it rain?",
		},
	}

	data := newTestDataClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/activity" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(activity)
	}))

	got, err := data.ListActivity(context.Background(), &ActivityParams{
		User: "0xuser1",
	})
	if err != nil {
		t.Fatalf("ListActivity() error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len(activity) = %d, want 1", len(got))
	}
	if got[0].Type != "TRADE" {
		t.Errorf("Type = %q, want %q", got[0].Type, "TRADE")
	}
}

func TestDataListHolders(t *testing.T) {
	holders := []TokenHolders{
		{
			Token: "tok1",
			Holders: []Holder{
				{ProxyWallet: "0xuser1", Amount: 500, Pseudonym: "whale"},
				{ProxyWallet: "0xuser2", Amount: 200, Pseudonym: "fish"},
			},
		},
	}

	data := newTestDataClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/holders" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("market") != "0xmarket1" {
			t.Errorf("market = %q, want %q", r.URL.Query().Get("market"), "0xmarket1")
		}
		json.NewEncoder(w).Encode(holders)
	}))

	got, err := data.ListHolders(context.Background(), &HoldersParams{
		Market: "0xmarket1",
	})
	if err != nil {
		t.Fatalf("ListHolders() error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len(holders) = %d, want 1", len(got))
	}
	if len(got[0].Holders) != 2 {
		t.Errorf("len(holders[0].Holders) = %d, want 2", len(got[0].Holders))
	}
}

func TestDataGetOpenInterest(t *testing.T) {
	oi := []OpenInterest{
		{Market: "0xmarket1", Value: 1500000},
	}

	data := newTestDataClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oi" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("market") != "0xmarket1" {
			t.Errorf("market = %q, want %q", r.URL.Query().Get("market"), "0xmarket1")
		}
		json.NewEncoder(w).Encode(oi)
	}))

	got, err := data.GetOpenInterest(context.Background(), "0xmarket1")
	if err != nil {
		t.Fatalf("GetOpenInterest() error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len(oi) = %d, want 1", len(got))
	}
	if got[0].Value != 1500000 {
		t.Errorf("Value = %f, want 1500000", got[0].Value)
	}
}

func TestDataAPIError(t *testing.T) {
	data := newTestDataClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":"rate limited"}`))
	}))

	_, err := data.ListTrades(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsRateLimited(err) {
		t.Errorf("IsRateLimited() = false, want true for error: %v", err)
	}
}

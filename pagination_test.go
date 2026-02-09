package polymarket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCursorIteratorMultiPage(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var page CursorPage[ClobMarket]

		switch callCount {
		case 1:
			if r.URL.Query().Get("next_cursor") != "" {
				t.Errorf("first call should not have next_cursor, got %q", r.URL.Query().Get("next_cursor"))
			}
			page = CursorPage[ClobMarket]{
				Data: []ClobMarket{
					{ConditionID: "0x1"},
					{ConditionID: "0x2"},
				},
				NextCursor: "cursor1",
			}
		case 2:
			if r.URL.Query().Get("next_cursor") != "cursor1" {
				t.Errorf("second call next_cursor = %q, want %q", r.URL.Query().Get("next_cursor"), "cursor1")
			}
			page = CursorPage[ClobMarket]{
				Data: []ClobMarket{
					{ConditionID: "0x3"},
				},
				NextCursor: "LTE=",
			}
		default:
			t.Fatal("unexpected call")
		}

		json.NewEncoder(w).Encode(page)
	}))

	c := NewClient(
		WithHTTPClient(srv.Client()),
		WithClobBaseURL(srv.URL),
	)

	iter := c.Clob.AllMarkets(context.Background())
	var markets []ClobMarket
	ctx := context.Background()

	for iter.Next(ctx) {
		markets = append(markets, iter.Item())
	}

	if err := iter.Err(); err != nil {
		t.Fatalf("iterator error: %v", err)
	}

	if len(markets) != 3 {
		t.Errorf("collected %d markets, want 3", len(markets))
	}

	if callCount != 2 {
		t.Errorf("server called %d times, want 2", callCount)
	}

	expected := []string{"0x1", "0x2", "0x3"}
	for i, m := range markets {
		if m.ConditionID != expected[i] {
			t.Errorf("markets[%d].ConditionID = %q, want %q", i, m.ConditionID, expected[i])
		}
	}
}

func TestCursorIteratorEmptyPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := CursorPage[ClobMarket]{
			Data:       []ClobMarket{},
			NextCursor: "",
		}
		json.NewEncoder(w).Encode(page)
	}))

	c := NewClient(
		WithHTTPClient(srv.Client()),
		WithClobBaseURL(srv.URL),
	)

	iter := c.Clob.AllMarkets(context.Background())
	ctx := context.Background()

	if iter.Next(ctx) {
		t.Error("Next() returned true for empty page")
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("iterator error: %v", err)
	}
}

func TestCursorIteratorError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"server error"}`))
	}))

	c := NewClient(
		WithHTTPClient(srv.Client()),
		WithClobBaseURL(srv.URL),
	)

	iter := c.Clob.AllMarkets(context.Background())
	ctx := context.Background()

	if iter.Next(ctx) {
		t.Error("Next() returned true on error")
	}
	if err := iter.Err(); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCursorIteratorSinglePage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := CursorPage[ClobMarket]{
			Data: []ClobMarket{
				{ConditionID: "0x1"},
			},
			NextCursor: "",
		}
		json.NewEncoder(w).Encode(page)
	}))

	c := NewClient(
		WithHTTPClient(srv.Client()),
		WithClobBaseURL(srv.URL),
	)

	iter := c.Clob.AllMarkets(context.Background())
	ctx := context.Background()
	var markets []ClobMarket

	for iter.Next(ctx) {
		markets = append(markets, iter.Item())
	}

	if err := iter.Err(); err != nil {
		t.Fatalf("iterator error: %v", err)
	}
	if len(markets) != 1 {
		t.Errorf("collected %d markets, want 1", len(markets))
	}
}

package polymarket

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestV2PostOrderWire(t *testing.T) {
	for _, tc := range []struct {
		name      string
		orderType OrderType
		options   PostOrderOptions
		wantType  string
	}{
		{"default", OrderGTC, PostOrderOptions{}, "GTC"},
		{"maker", OrderGTD, PostOrderOptions{PostOnly: true, DeferExec: true}, "GTD"},
		{"immediate", OrderFAK, PostOrderOptions{}, "FAK"},
		{"legacy immediate alias", OrderIOC, PostOrderOptions{}, "FAK"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" || r.URL.Path != "/order" {
					t.Errorf("request: %s %s", r.Method, r.URL.Path)
				}
				requireAuthHeaders(t, r)
				requireNoBuilderHeaders(t, r)
				var body struct {
					Order               map[string]json.RawMessage
					Owner, OrderType    string
					PostOnly, DeferExec bool
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
					return
				}
				for _, key := range []string{"nonce", "feeRateBps", "taker"} {
					if _, ok := body.Order[key]; ok {
						t.Errorf("legacy field %s sent", key)
					}
				}
				for key, want := range map[string]string{"timestamp": `"1780449126930"`, "expiration": `"1900000000"`, "side": `"BUY"`, "signatureType": "0", "makerAmount": `"5000000"`, "takerAmount": `"10000000"`, "metadata": `"` + zeroBytes32 + `"`, "builder": `"0x` + strings.Repeat("22", 32) + `"`} {
					if string(body.Order[key]) != want {
						t.Errorf("%s=%s, want %s", key, body.Order[key], want)
					}
				}
				if body.Owner != testCreds.Key || body.OrderType != tc.wantType || body.PostOnly != tc.options.PostOnly || body.DeferExec != tc.options.DeferExec {
					t.Errorf("incorrect envelope: %+v", body)
				}
				if len(body.Order["salt"]) == 0 || body.Order["salt"][0] == '"' {
					t.Error("salt must be a JSON number")
				}
				_, _ = io.WriteString(w, `{"success":true,"orderID":"order-1","status":"matched","makingAmount":"5","takingAmount":"10","tradeIDs":["trade-1"],"transactionsHashes":["0xabc"]}`)
			}))
			defer srv.Close()
			client := NewClient(WithHTTPClient(srv.Client()), WithClobBaseURL(srv.URL), WithCredentials(testCreds), WithBuilderCredentials(testBuilderCreds))
			signer, err := NewOrderSigner(testPrivateKey, Polygon)
			if err != nil {
				t.Fatal(err)
			}
			signed, err := signer.CreateOrder(CreateOrderParams{TokenID: "123", Price: .5, Size: 10, Side: Buy, Timestamp: 1780449126930, Expiration: 1900000000, Builder: "0x" + strings.Repeat("22", 32)})
			if err != nil {
				t.Fatal(err)
			}
			var resp *PostOrderResponse
			if tc.name == "default" {
				resp, err = client.Clob.PostOrder(t.Context(), signed, tc.orderType)
			} else {
				resp, err = client.Clob.PostOrderWithOptions(t.Context(), signed, tc.orderType, tc.options)
			}
			if err != nil {
				t.Fatal(err)
			}
			if resp.Status != "matched" || resp.MakingAmount != "5" || resp.TakingAmount != "10" || !reflect.DeepEqual(resp.TradeIDs, []string{"trade-1"}) {
				t.Fatalf("lost execution data: %+v", resp)
			}
		})
	}
}

func TestV2PostOnlyRejectsImmediateOrders(t *testing.T) {
	client := NewClient(WithCredentials(testCreds))
	signer, err := NewOrderSigner(testPrivateKey, Polygon)
	if err != nil {
		t.Fatal(err)
	}
	signed, err := signer.CreateOrder(CreateOrderParams{TokenID: "123", Price: .5, Size: 10, Side: Buy})
	if err != nil {
		t.Fatal(err)
	}
	for _, typ := range []OrderType{OrderFOK, OrderFAK, OrderIOC, "unknown"} {
		if _, err := client.Clob.PostOrderWithOptions(t.Context(), signed, typ, PostOrderOptions{PostOnly: true}); err == nil {
			t.Errorf("accepted %s", typ)
		}
	}
	if _, err := client.Clob.PostOrder(t.Context(), nil, OrderGTC); err == nil {
		t.Error("nil signed order accepted")
	}
}

func TestV2CancellationContract(t *testing.T) {
	for _, tc := range []struct {
		name, path, wantBody, response string
		ids                            []string
		wantCanceled                   []string
		wantErr                        bool
	}{
		{"single", "/order", `{"orderID":"one"}`, `{"canceled":["one"],"not_canceled":{}}`, []string{"one"}, nil, false},
		{"rejected", "/order", `{"orderID":"one"}`, `{"canceled":[],"not_canceled":{"one":"already matched"}}`, []string{"one"}, nil, true},
		{"missing acknowledgement", "/order", `{"orderID":"one"}`, `{"canceled":[],"not_canceled":{}}`, []string{"one"}, nil, true},
		{"batch", "/orders", `["one","two"]`, `{"canceled":["one","two"],"not_canceled":{}}`, []string{"one", "two"}, []string{"one", "two"}, false},
		{"partial batch", "/orders", `["one","two"]`, `{"canceled":["one"],"not_canceled":{"two":"already matched"}}`, []string{"one", "two"}, []string{"one"}, true},
		{"missing batch acknowledgement", "/orders", `["one","two"]`, `{"canceled":["one"],"not_canceled":{}}`, []string{"one", "two"}, []string{"one"}, true},
		{"all", "/cancel-all", "", `{"canceled":["one"],"not_canceled":{}}`, nil, []string{"one"}, false},
		{"partial all", "/cancel-all", "", `{"canceled":["one"],"not_canceled":{"two":"retry"}}`, nil, []string{"one"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				b, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
				}
				if r.Method != "DELETE" || r.URL.Path != tc.path || string(b) != tc.wantBody {
					t.Errorf("unexpected request: %s %s %s", r.Method, r.URL.Path, b)
				}
				_, _ = io.WriteString(w, tc.response)
			}))
			defer srv.Close()
			clob := NewClient(WithClobBaseURL(srv.URL), WithHTTPClient(srv.Client()), WithCredentials(testCreds)).Clob
			var canceled []string
			var err error
			switch tc.path {
			case "/order":
				err = clob.CancelOrder(t.Context(), tc.ids[0])
			case "/orders":
				canceled, err = clob.CancelOrders(t.Context(), tc.ids)
			default:
				canceled, err = clob.CancelAll(t.Context())
			}
			if (err != nil) != tc.wantErr {
				t.Fatalf("error=%v", err)
			}
			if !reflect.DeepEqual(canceled, tc.wantCanceled) {
				t.Errorf("canceled=%v, want %v", canceled, tc.wantCanceled)
			}
			if tc.wantErr {
				var ce *CancelError
				if !errors.As(err, &ce) || len(ce.NotCanceled) == 0 {
					t.Errorf("missing cancellation details: %v", err)
				}
			}
		})
	}
}

func TestV2CancelAllRejectsMalformedAcknowledgement(t *testing.T) {
	for _, body := range []string{`{}`, `null`, `{"error":"unavailable"}`, `not json`, ``} {
		t.Run(body, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, body) }))
			defer srv.Close()
			clob := NewClient(WithClobBaseURL(srv.URL), WithHTTPClient(srv.Client()), WithCredentials(testCreds)).Clob
			if _, err := clob.CancelAll(t.Context()); err == nil {
				t.Error("malformed acknowledgement accepted")
			}
		})
	}
}

func TestV2BuilderTrades(t *testing.T) {
	code := "0x" + strings.Repeat("22", 32)
	market := "condition-1"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/builder/trades" || r.URL.Query().Get("builder_code") != code || r.URL.Query().Get("market") != market {
			t.Errorf("incorrect builder query %s", r.URL)
		}
		if r.URL.Query().Get("next_cursor") == "MA==" {
			_, _ = io.WriteString(w, `{"data":[{"id":"one","assetId":"123","maker":"0xabc","takerOrderHash":"order-1","side":"BUY","price":"0.5","size":"2","builderFee":"0.01"}],"next_cursor":"next"}`)
		} else {
			_, _ = io.WriteString(w, `{"data":[{"id":"two"}],"next_cursor":"LTE="}`)
		}
	}))
	defer srv.Close()
	clob := NewClient(WithHTTPClient(srv.Client()), WithClobBaseURL(srv.URL), WithBuilderCode(code)).Clob
	trades, err := clob.GetBuilderTradesByCode(t.Context(), code, &market)
	if err != nil {
		t.Fatal(err)
	}
	if len(trades) != 2 || trades[0].AssetID != "123" || trades[0].BuilderFee != "0.01" {
		t.Fatalf("builder trades: %+v", trades)
	}
	legacy, err := clob.GetBuilderTrades(t.Context(), &market)
	if err != nil {
		t.Fatal(err)
	}
	if len(legacy) != 2 || legacy[0].TakerOrderID != "order-1" || legacy[0].MakerAddress != "0xabc" {
		t.Fatalf("compatibility fields: %+v", legacy)
	}
	for _, bad := range []string{"", zeroBytes32, "bad"} {
		if _, err := clob.GetBuilderTradesByCode(t.Context(), bad, nil); err == nil {
			t.Errorf("accepted builder %q", bad)
		}
	}
}

func TestV2BuilderTradeNullableFields(t *testing.T) {
	for _, body := range []string{
		`{"err_msg":"settlement failed","createdAt":"2026-09-04T12:00:00Z","updatedAt":"2026-09-04T12:01:00Z"}`,
		`{"err_msg":null,"createdAt":null,"updatedAt":null}`,
	} {
		var trade BuilderTrade
		if err := json.Unmarshal([]byte(body), &trade); err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(trade)
		if err != nil {
			t.Fatal(err)
		}
		var want, got map[string]json.RawMessage
		if err := json.Unmarshal([]byte(body), &want); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(encoded, &got); err != nil {
			t.Fatal(err)
		}
		for key, value := range want {
			if string(got[key]) != string(value) {
				t.Errorf("%s=%s, want %s", key, got[key], value)
			}
		}
	}
}

func TestV2TradeAttributionAndBalances(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/data/trades":
			if r.URL.Query().Get("next_cursor") == "MA==" {
				_, _ = io.WriteString(w, `{"data":[{"id":"trade-1","taker_order_id":"taker-1","trader_side":"MAKER","status":"CONFIRMED","last_update":"1780449126","maker_orders":[{"order_id":"maker-1","owner":"key-1","maker_address":"0xabc","matched_amount":"2","price":"0.5","asset_id":"123","outcome":"Yes","side":"SELL","fee_rate_bps":"0"}]}],"next_cursor":"page2"}`)
			} else {
				_, _ = io.WriteString(w, `{"data":[{"id":"trade-2","taker_order_id":"taker-2","trader_side":"TAKER"}],"next_cursor":"LTE="}`)
			}
		case "/balance-allowance":
			_, _ = io.WriteString(w, `{"balance":"10000000","allowances":{"0xexchange":"5000000"}}`)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()
	clob := NewClient(WithClobBaseURL(srv.URL), WithHTTPClient(srv.Client()), WithCredentials(testCreds)).Clob
	trades, err := clob.GetTrades(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(trades) != 2 || trades[0].TakerOrderID != "taker-1" || trades[0].TraderSide != "MAKER" || len(trades[0].MakerOrders) != 1 || trades[0].MakerOrders[0].OrderID != "maker-1" || trades[0].MakerOrders[0].MatchedAmount != "2" || trades[1].TakerOrderID != "taker-2" {
		t.Fatalf("lost attribution: %+v", trades)
	}
	ba, err := clob.GetBalanceAllowance(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if ba.Allowances["0xexchange"] != "5000000" {
		t.Fatalf("lost allowances: %+v", ba)
	}
}

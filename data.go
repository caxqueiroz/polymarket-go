package polymarket

import (
	"context"
	"net/url"
	"strconv"
)

// DataClient provides access to the Polymarket Data API.
type DataClient struct {
	base    *baseClient
	baseURL string
}

// ListTrades returns trades from the Data API.
func (d *DataClient) ListTrades(ctx context.Context, p *TradesParams) ([]Trade, error) {
	params := buildTradesParams(p)

	var trades []Trade
	if err := d.base.get(ctx, d.baseURL, "/trades", params, &trades); err != nil {
		return nil, err
	}
	return trades, nil
}

// ListPositions returns positions for a user from the Data API.
func (d *DataClient) ListPositions(ctx context.Context, p *PositionsParams) ([]Position, error) {
	params := buildPositionsParams(p)

	var positions []Position
	if err := d.base.get(ctx, d.baseURL, "/positions", params, &positions); err != nil {
		return nil, err
	}
	return positions, nil
}

// ListActivity returns activity for a user from the Data API.
func (d *DataClient) ListActivity(ctx context.Context, p *ActivityParams) ([]Activity, error) {
	params := buildActivityParams(p)

	var activity []Activity
	if err := d.base.get(ctx, d.baseURL, "/activity", params, &activity); err != nil {
		return nil, err
	}
	return activity, nil
}

// ListHolders returns top holders for a market from the Data API.
func (d *DataClient) ListHolders(ctx context.Context, p *HoldersParams) ([]TokenHolders, error) {
	params := url.Values{}
	if p != nil {
		params.Set("market", p.Market)
		if p.Limit != nil {
			params.Set("limit", strconv.Itoa(*p.Limit))
		}
	}

	var holders []TokenHolders
	if err := d.base.get(ctx, d.baseURL, "/holders", params, &holders); err != nil {
		return nil, err
	}
	return holders, nil
}

// GetOpenInterest returns open interest for one or more markets (comma-separated condition IDs).
func (d *DataClient) GetOpenInterest(ctx context.Context, market string) ([]OpenInterest, error) {
	params := url.Values{}
	if market != "" {
		params.Set("market", market)
	}

	var oi []OpenInterest
	if err := d.base.get(ctx, d.baseURL, "/oi", params, &oi); err != nil {
		return nil, err
	}
	return oi, nil
}

func buildTradesParams(p *TradesParams) url.Values {
	if p == nil {
		return nil
	}
	params := url.Values{}
	if p.User != "" {
		params.Set("user", p.User)
	}
	if p.Market != "" {
		params.Set("market", p.Market)
	}
	if p.Limit != nil {
		params.Set("limit", strconv.Itoa(*p.Limit))
	}
	if p.Offset != nil {
		params.Set("offset", strconv.Itoa(*p.Offset))
	}
	if p.Side != nil {
		params.Set("side", string(*p.Side))
	}
	if p.TakerOnly != nil {
		params.Set("takerOnly", strconv.FormatBool(*p.TakerOnly))
	}
	if p.FilterType != "" {
		params.Set("filterType", p.FilterType)
	}
	if p.FilterAmount != nil {
		params.Set("filterAmount", strconv.FormatFloat(*p.FilterAmount, 'f', -1, 64))
	}
	return params
}

func buildPositionsParams(p *PositionsParams) url.Values {
	if p == nil {
		return nil
	}
	params := url.Values{}
	if p.User != "" {
		params.Set("user", p.User)
	}
	if p.Market != "" {
		params.Set("market", p.Market)
	}
	if p.Limit != nil {
		params.Set("limit", strconv.Itoa(*p.Limit))
	}
	if p.Offset != nil {
		params.Set("offset", strconv.Itoa(*p.Offset))
	}
	if p.SizeThreshold != nil {
		params.Set("sizeThreshold", strconv.FormatFloat(*p.SizeThreshold, 'f', -1, 64))
	}
	if p.Redeemable != nil {
		params.Set("redeemable", strconv.FormatBool(*p.Redeemable))
	}
	if p.Mergeable != nil {
		params.Set("mergeable", strconv.FormatBool(*p.Mergeable))
	}
	if p.Title != "" {
		params.Set("title", p.Title)
	}
	if p.SortBy != "" {
		params.Set("sortBy", p.SortBy)
	}
	if p.SortDirection != "" {
		params.Set("sortDirection", p.SortDirection)
	}
	return params
}

func buildActivityParams(p *ActivityParams) url.Values {
	if p == nil {
		return nil
	}
	params := url.Values{}
	if p.User != "" {
		params.Set("user", p.User)
	}
	if p.Market != "" {
		params.Set("market", p.Market)
	}
	if p.Limit != nil {
		params.Set("limit", strconv.Itoa(*p.Limit))
	}
	if p.Offset != nil {
		params.Set("offset", strconv.Itoa(*p.Offset))
	}
	if p.Type != "" {
		params.Set("type", p.Type)
	}
	if p.Side != nil {
		params.Set("side", string(*p.Side))
	}
	if p.Start != nil {
		params.Set("start", strconv.FormatInt(*p.Start, 10))
	}
	if p.End != nil {
		params.Set("end", strconv.FormatInt(*p.End, 10))
	}
	if p.SortBy != "" {
		params.Set("sortBy", p.SortBy)
	}
	if p.SortDirection != "" {
		params.Set("sortDirection", p.SortDirection)
	}
	return params
}

package polymarket

import (
	"context"
	"net/url"
	"strconv"
	"strings"
)

// GammaClient provides access to the Polymarket Gamma API.
type GammaClient struct {
	base    *baseClient
	baseURL string
}

// ListMarkets returns markets from the Gamma API. Pass nil for default params.
func (g *GammaClient) ListMarkets(ctx context.Context, p *GammaMarketParams) ([]GammaMarket, error) {
	params := buildGammaMarketParams(p)

	var markets []GammaMarket
	if err := g.base.get(ctx, g.baseURL, "/markets", params, &markets); err != nil {
		return nil, err
	}
	return markets, nil
}

// GetMarketByID retrieves a single market by its numeric ID.
func (g *GammaClient) GetMarketByID(ctx context.Context, id string) (*GammaMarket, error) {
	params := url.Values{}
	params.Set("id", id)

	var markets []GammaMarket
	if err := g.base.get(ctx, g.baseURL, "/markets", params, &markets); err != nil {
		return nil, err
	}
	if len(markets) == 0 {
		return nil, &APIError{StatusCode: 404, Status: "404 Not Found", Body: "market not found"}
	}
	return &markets[0], nil
}

// GetMarketByConditionID retrieves a single market by its on-chain condition ID.
func (g *GammaClient) GetMarketByConditionID(ctx context.Context, conditionID string) (*GammaMarket, error) {
	params := url.Values{}
	params.Set("condition_id", conditionID)

	var markets []GammaMarket
	if err := g.base.get(ctx, g.baseURL, "/markets", params, &markets); err != nil {
		return nil, err
	}
	if len(markets) == 0 {
		return nil, &APIError{StatusCode: 404, Status: "404 Not Found", Body: "market not found"}
	}
	return &markets[0], nil
}

// GetMarketBySlug retrieves a single market by its slug.
func (g *GammaClient) GetMarketBySlug(ctx context.Context, slug string) (*GammaMarket, error) {
	params := url.Values{}
	params.Set("slug", slug)

	var markets []GammaMarket
	if err := g.base.get(ctx, g.baseURL, "/markets", params, &markets); err != nil {
		return nil, err
	}
	if len(markets) == 0 {
		return nil, &APIError{StatusCode: 404, Status: "404 Not Found", Body: "market not found"}
	}
	return &markets[0], nil
}

// ListEvents returns events from the Gamma API. Pass nil for default params.
func (g *GammaClient) ListEvents(ctx context.Context, p *GammaEventParams) ([]GammaEvent, error) {
	params := buildGammaEventParams(p)

	var events []GammaEvent
	if err := g.base.get(ctx, g.baseURL, "/events", params, &events); err != nil {
		return nil, err
	}
	return events, nil
}

// GetEventByID retrieves a single event by its numeric ID.
func (g *GammaClient) GetEventByID(ctx context.Context, id string) (*GammaEvent, error) {
	var event GammaEvent
	if err := g.base.get(ctx, g.baseURL, "/events/"+id, nil, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

// GetEventBySlug retrieves a single event by its slug.
func (g *GammaClient) GetEventBySlug(ctx context.Context, slug string) (*GammaEvent, error) {
	params := url.Values{}
	params.Set("slug", slug)

	var events []GammaEvent
	if err := g.base.get(ctx, g.baseURL, "/events", params, &events); err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, &APIError{StatusCode: 404, Status: "404 Not Found", Body: "event not found"}
	}
	return &events[0], nil
}

// Search performs a text search across markets and events.
func (g *GammaClient) Search(ctx context.Context, query string) (SearchResponse, error) {
	params := url.Values{}
	params.Set("q", query)

	var results SearchResponse
	if err := g.base.get(ctx, g.baseURL, "/public-search", params, &results); err != nil {
		return SearchResponse{}, err
	}
	return results, nil
}

// ListTags returns all tags.
func (g *GammaClient) ListTags(ctx context.Context) ([]GammaTag, error) {
	var tags []GammaTag
	if err := g.base.get(ctx, g.baseURL, "/tags", nil, &tags); err != nil {
		return nil, err
	}
	return tags, nil
}

// GetTagByID retrieves a single tag by its numeric ID.
func (g *GammaClient) GetTagByID(ctx context.Context, id string) (*GammaTag, error) {
	var tag GammaTag
	if err := g.base.get(ctx, g.baseURL, "/tags/"+id, nil, &tag); err != nil {
		return nil, err
	}
	return &tag, nil
}

// GetTagBySlug retrieves a tag by its slug.
func (g *GammaClient) GetTagBySlug(ctx context.Context, slug string) (*GammaTag, error) {
	params := url.Values{}
	params.Set("slug", slug)

	var tags []GammaTag
	if err := g.base.get(ctx, g.baseURL, "/tags", params, &tags); err != nil {
		return nil, err
	}
	if len(tags) == 0 {
		return nil, &APIError{StatusCode: 404, Status: "404 Not Found", Body: "tag not found"}
	}
	return &tags[0], nil
}

func buildGammaMarketParams(p *GammaMarketParams) url.Values {
	if p == nil {
		return nil
	}
	params := url.Values{}
	if p.Limit != nil {
		params.Set("limit", strconv.Itoa(*p.Limit))
	}
	if p.Offset != nil {
		params.Set("offset", strconv.Itoa(*p.Offset))
	}
	if p.Order != "" {
		params.Set("order", p.Order)
	}
	if p.Ascending != nil {
		params.Set("ascending", strconv.FormatBool(*p.Ascending))
	}
	if p.ID != "" {
		params.Set("id", p.ID)
	}
	if p.Slug != "" {
		params.Set("slug", p.Slug)
	}
	if p.Closed != nil {
		params.Set("closed", strconv.FormatBool(*p.Closed))
	}
	if p.Active != nil {
		params.Set("active", strconv.FormatBool(*p.Active))
	}
	if p.TagID != nil {
		params.Set("tag_id", strconv.Itoa(*p.TagID))
	}
	if len(p.ExcludeTagID) > 0 {
		ids := make([]string, len(p.ExcludeTagID))
		for i, id := range p.ExcludeTagID {
			ids[i] = strconv.Itoa(id)
		}
		params.Set("exclude_tag_id", strings.Join(ids, ","))
	}
	if p.RelatedTags != nil {
		params.Set("related_tags", strconv.FormatBool(*p.RelatedTags))
	}
	if p.Featured != nil {
		params.Set("featured", strconv.FormatBool(*p.Featured))
	}
	if p.EnableOrderBook != nil {
		params.Set("enable_order_book", strconv.FormatBool(*p.EnableOrderBook))
	}
	return params
}

func buildGammaEventParams(p *GammaEventParams) url.Values {
	if p == nil {
		return nil
	}
	params := url.Values{}
	if p.Limit != nil {
		params.Set("limit", strconv.Itoa(*p.Limit))
	}
	if p.Offset != nil {
		params.Set("offset", strconv.Itoa(*p.Offset))
	}
	if p.Order != "" {
		params.Set("order", p.Order)
	}
	if p.Ascending != nil {
		params.Set("ascending", strconv.FormatBool(*p.Ascending))
	}
	if len(p.ID) > 0 {
		ids := make([]string, len(p.ID))
		for i, id := range p.ID {
			ids[i] = strconv.Itoa(id)
		}
		params.Set("id", strings.Join(ids, ","))
	}
	if len(p.Slug) > 0 {
		params.Set("slug", strings.Join(p.Slug, ","))
	}
	if p.TagID != nil {
		params.Set("tag_id", strconv.Itoa(*p.TagID))
	}
	if len(p.ExcludeTagID) > 0 {
		ids := make([]string, len(p.ExcludeTagID))
		for i, id := range p.ExcludeTagID {
			ids[i] = strconv.Itoa(id)
		}
		params.Set("exclude_tag_id", strings.Join(ids, ","))
	}
	if p.RelatedTags != nil {
		params.Set("related_tags", strconv.FormatBool(*p.RelatedTags))
	}
	if p.Featured != nil {
		params.Set("featured", strconv.FormatBool(*p.Featured))
	}
	if p.Closed != nil {
		params.Set("closed", strconv.FormatBool(*p.Closed))
	}
	if p.StartDateMin != "" {
		params.Set("start_date_min", p.StartDateMin)
	}
	if p.StartDateMax != "" {
		params.Set("start_date_max", p.StartDateMax)
	}
	if p.EndDateMin != "" {
		params.Set("end_date_min", p.EndDateMin)
	}
	if p.EndDateMax != "" {
		params.Set("end_date_max", p.EndDateMax)
	}
	return params
}

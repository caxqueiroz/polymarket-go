package polymarket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// CursorPage represents a page of results from the CLOB API using cursor pagination.
type CursorPage[T any] struct {
	Data       []T    `json:"data"`
	NextCursor string `json:"next_cursor"`
}

// CursorIterator iterates over all pages of a CLOB cursor-paginated endpoint.
type CursorIterator[T any] struct {
	base    *baseClient
	baseURL string
	path    string
	params  url.Values

	page    []T
	pos     int
	cursor  string
	started bool
	done    bool
	err     error
}

func newCursorIterator[T any](base *baseClient, baseURL, path string, params url.Values) *CursorIterator[T] {
	if params == nil {
		params = url.Values{}
	}
	return &CursorIterator[T]{
		base:    base,
		baseURL: baseURL,
		path:    path,
		params:  params,
	}
}

// Next advances the iterator and returns true if there is another item.
func (it *CursorIterator[T]) Next(ctx context.Context) bool {
	if it.done || it.err != nil {
		return false
	}

	if it.started && it.pos < len(it.page) {
		return true
	}

	if it.started && it.cursor == "" {
		it.done = true
		return false
	}

	it.err = it.fetchPage(ctx)
	if it.err != nil {
		return false
	}

	return len(it.page) > 0
}

// Item returns the current item.
func (it *CursorIterator[T]) Item() T {
	item := it.page[it.pos]
	it.pos++
	return item
}

// Err returns any error encountered during iteration.
func (it *CursorIterator[T]) Err() error {
	return it.err
}

func (it *CursorIterator[T]) fetchPage(ctx context.Context) error {
	params := url.Values{}
	for k, v := range it.params {
		params[k] = v
	}
	if it.started && it.cursor != "" {
		params.Set("next_cursor", it.cursor)
	}
	it.started = true

	body, err := it.base.getRaw(ctx, it.baseURL, it.path, params)
	if err != nil {
		return fmt.Errorf("polymarket: fetching page: %w", err)
	}

	var page CursorPage[T]
	if err := json.Unmarshal(body, &page); err != nil {
		return fmt.Errorf("polymarket: decoding page: %w", err)
	}

	it.page = page.Data
	it.pos = 0
	it.cursor = page.NextCursor
	if it.cursor == "LTE=" {
		it.cursor = ""
	}

	return nil
}

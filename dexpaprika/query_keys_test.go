package dexpaprika

import (
	"context"
	"sort"
	"strings"
	"testing"
)

// From 23 November 2026 the API answers 400 to a query key a route does not
// accept, and to the same key sent twice. Until then it ignores both, so a
// misspelled filter looks like an unfiltered page. These are the routes'
// accepted keys as the API defines them (dexpaprika-go
// internal/doc/all/static/api-rest.yml, 16 Nov 2026).
var (
	searchCommonKeys = []string{
		"created_after", "created_before", "cursor", "detailed", "limit", "order_by", "sort",
		"liquidity_usd_max", "liquidity_usd_min", "price_usd_max", "price_usd_min",
		"txns_24h_max", "txns_24h_min",
		"volume_usd_24h_max", "volume_usd_24h_min", "volume_usd_7d_max", "volume_usd_7d_min",
		"volume_usd_30d_max", "volume_usd_30d_min",
	}
	poolSearchExtra  = []string{"dex_name", "token_address"}
	tokenSearchExtra = []string{"fdv_usd_max", "fdv_usd_min", "price_change_percentage_24h_max", "price_change_percentage_24h_min"}
	transactionKeys  = []string{"cursor", "from", "limit", "page", "to"}
)

func keySet(lists ...[]string) map[string]bool {
	out := map[string]bool{}
	for _, l := range lists {
		for _, k := range l {
			out[k] = true
		}
	}
	return out
}

func rejectedKeys(t *testing.T, call func(c *Client) error, allowed map[string]bool) []string {
	t.Helper()
	q := queryFor(t, `{"results":[],"has_next_page":false,"transactions":[],"page_info":{}}`, call)
	var bad []string
	for k, v := range q {
		if !allowed[k] {
			bad = append(bad, k)
		} else if len(v) > 1 {
			bad = append(bad, k+" (sent twice)")
		}
	}
	sort.Strings(bad)
	return bad
}

func TestFilterSendsOnlyAcceptedKeys(t *testing.T) {
	f := func(v float64) *float64 { return &v }
	n := 1
	priceWindows := []string{}
	for _, w := range []string{"5m", "1h", "6h", "24h", "48h"} {
		priceWindows = append(priceWindows, "price_change_percentage_"+w+"_min", "price_change_percentage_"+w+"_max")
	}

	bad := rejectedKeys(t, func(c *Client) error {
		_, err := c.Pools.Filter(context.Background(), "ethereum", &PoolFilterOptions{
			Page: 2, Limit: 5, SortBy: "volume_usd_24h", SortDir: "asc", Cursor: "c",
			Volume24hMin: f(1), Volume24hMax: f(2), Volume7dMin: f(1), Volume7dMax: f(2),
			LiquidityMin: f(1), LiquidityMax: f(2), Txns24hMin: &n, CreatedAfter: "-1d", CreatedBefore: "-1h",
			PriceChange24hMin: f(1), PriceChange24hMax: f(2), PriceChange6hMin: f(1), PriceChange6hMax: f(2),
			PriceChange1hMin: f(1), PriceChange1hMax: f(2), PriceChange5mMin: f(1), PriceChange5mMax: f(2),
			PriceChange48hMin: f(1), PriceChange48hMax: f(2),
		})
		return err
	}, keySet(searchCommonKeys, poolSearchExtra, priceWindows))
	if len(bad) > 0 {
		t.Errorf("Pools.Filter sends keys pools/search rejects: %s", strings.Join(bad, ", "))
	}

	bad = rejectedKeys(t, func(c *Client) error {
		_, err := c.Tokens.Filter(context.Background(), "ethereum", &TokenFilterOptions{
			Page: 2, Limit: 5, SortBy: "volume_usd_24h", SortDir: "asc", Cursor: "c",
			Volume24hMin: f(1), Volume24hMax: f(2), LiquidityMin: f(1), FDVMin: f(1), FDVMax: f(2),
			Txns24hMin: &n, CreatedAfter: "-1d", CreatedBefore: "-1h",
			PriceChange24hMin: f(1), PriceChange24hMax: f(2),
		})
		return err
	}, keySet(searchCommonKeys, tokenSearchExtra))
	if len(bad) > 0 {
		t.Errorf("Tokens.Filter sends keys tokens/search rejects: %s", strings.Join(bad, ", "))
	}
}

func TestTransactionOptionsSendEachKeyOnce(t *testing.T) {
	bad := rejectedKeys(t, func(c *Client) error {
		_, err := c.Pools.GetTransactions(context.Background(), "ethereum", testPool, 0, 10, "",
			WithFrom("-1h"), WithFromTimestamp(1790000000), WithToTimestamp(1790000100), WithTo("-5m"))
		return err
	}, keySet(transactionKeys))
	if len(bad) > 0 {
		t.Errorf("GetTransactions sends keys the route rejects: %s", strings.Join(bad, ", "))
	}
}

func TestSearchEncodesQueryOnce(t *testing.T) {
	q := queryFor(t, `{"tokens":[],"pools":[],"dexes":[]}`, func(c *Client) error {
		_, err := c.Search.Search(context.Background(), "usd+coin")
		return err
	})
	if got := q.Get("query"); got != "usd+coin" {
		t.Errorf("query reached the API as %q, want %q", got, "usd+coin")
	}
}

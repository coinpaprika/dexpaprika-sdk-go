package dexpaprika

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestPools_AdvancedSearch_SortTranslation is the guard against the #1 trap on this
// endpoint: callers pass canonical SortBy/SortDir, but the wire must carry order_by/sort.
// It also confirms the raw backend names are NEVER leaked onto the query string.
func TestPools_AdvancedSearch_SortTranslation(t *testing.T) {
	var gotQuery map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/frontend/v1/pools" {
			t.Errorf("expected path /frontend/v1/pools, got %s", r.URL.Path)
		}
		gotQuery = map[string]string{}
		for k := range r.URL.Query() {
			gotQuery[k] = r.URL.Query().Get(k)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"results":[],"has_next_page":false,"next_cursor":null,"query":{}}`))
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithRetryConfig(0, time.Millisecond, time.Millisecond),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	minPrice := 0.5
	_, err := client.Pools.AdvancedSearch(ctx, &AdvancedSearchOptions{
		Limit:       3,
		SortBy:      "volume_usd_24h",
		SortDir:     "desc",
		PriceUSDMin: &minPrice,
		DexName:     "uniswap_v3",
		Detailed:    true,
	})
	if err != nil {
		t.Fatalf("AdvancedSearch returned error: %v", err)
	}

	// The canonical names must be translated to the wire names.
	if gotQuery["order_by"] != "volume_usd_24h" {
		t.Errorf("expected order_by=volume_usd_24h on the wire, got %q", gotQuery["order_by"])
	}
	if gotQuery["sort"] != "desc" {
		t.Errorf("expected sort=desc on the wire, got %q", gotQuery["sort"])
	}
	// The raw canonical names must NOT leak onto the wire.
	if _, ok := gotQuery["sort_by"]; ok {
		t.Errorf("sort_by leaked onto the wire; it must be translated to order_by")
	}
	if _, ok := gotQuery["sort_dir"]; ok {
		t.Errorf("sort_dir leaked onto the wire; it must be translated to sort")
	}
	// Filters pass through unchanged.
	if gotQuery["price_usd_min"] != "0.5" {
		t.Errorf("expected price_usd_min=0.5, got %q", gotQuery["price_usd_min"])
	}
	if gotQuery["dex_name"] != "uniswap_v3" {
		t.Errorf("expected dex_name=uniswap_v3, got %q", gotQuery["dex_name"])
	}
	if gotQuery["detailed"] != "true" {
		t.Errorf("expected detailed=true, got %q", gotQuery["detailed"])
	}
	if gotQuery["limit"] != "3" {
		t.Errorf("expected limit=3, got %q", gotQuery["limit"])
	}
}

// TestPools_AdvancedSearch_Live hits the real /frontend/v1/pools endpoint.
func TestPools_AdvancedSearch_Live(t *testing.T) {
	client := NewClient(WithRetryConfig(2, time.Second, 3*time.Second))
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	minPrice := 0.5
	res, err := client.Pools.AdvancedSearch(ctx, &AdvancedSearchOptions{
		Limit:       5,
		SortBy:      "volume_usd_24h",
		SortDir:     "desc",
		PriceUSDMin: &minPrice,
		Detailed:    true,
	})
	if err != nil {
		t.Fatalf("AdvancedSearch live returned error: %v", err)
	}
	if res == nil {
		t.Fatal("AdvancedSearch returned nil response")
	}
	if len(res.Results) == 0 {
		t.Fatal("AdvancedSearch returned zero results; expected a non-empty page")
	}

	// The price_usd_min filter must actually be honored by the backend.
	for i, row := range res.Results {
		if row.ID == "" {
			t.Errorf("result %d has empty pool id", i)
		}
		if row.PriceUSD != nil && *row.PriceUSD < minPrice {
			t.Errorf("result %d price_usd %.6f is below the requested minimum %.2f", i, *row.PriceUSD, minPrice)
		}
	}

	// detailed=true should attach timeframe blocks to at least one token.
	foundDetailed := false
	for _, row := range res.Results {
		for _, tok := range row.Tokens {
			if tok.Day != nil || tok.Hour1 != nil {
				foundDetailed = true
			}
		}
	}
	if !foundDetailed {
		t.Error("detailed=true did not produce any per-timeframe token metric blocks")
	}

	// The query echo should reflect the translated wire name.
	if len(res.Query) > 0 {
		var echo map[string]any
		if err := json.Unmarshal(res.Query, &echo); err != nil {
			t.Errorf("failed to decode query echo: %v", err)
		} else if echo["order_by"] != "volume_usd_24h" {
			t.Errorf("query echo order_by = %v, expected volume_usd_24h", echo["order_by"])
		}
	}

	t.Logf("AdvancedSearch returned %d pools, has_next_page=%v", len(res.Results), res.HasNextPage)
}

// TestPools_AdvancedSearchByNetwork_Live hits /frontend/v1/networks/{network}/pools
// and verifies cursor pagination plus the per-network chain constraint.
func TestPools_AdvancedSearchByNetwork_Live(t *testing.T) {
	client := NewClient(WithRetryConfig(2, time.Second, 3*time.Second))
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	const network = "ethereum"
	res, err := client.Pools.AdvancedSearchByNetwork(ctx, network, &AdvancedSearchOptions{
		Limit:   3,
		SortBy:  "volume_usd_24h",
		SortDir: "desc",
	})
	if err != nil {
		t.Fatalf("AdvancedSearchByNetwork live returned error: %v", err)
	}
	if len(res.Results) == 0 {
		t.Fatal("AdvancedSearchByNetwork returned zero results")
	}
	for i, row := range res.Results {
		if row.Chain != network {
			t.Errorf("result %d chain = %q, expected %q", i, row.Chain, network)
		}
	}

	// Volume-desc ordering means each successive volume_usd_24h should not increase.
	var prev *float64
	for _, row := range res.Results {
		if row.VolumeUSD24h == nil {
			continue
		}
		if prev != nil && *row.VolumeUSD24h > *prev {
			t.Errorf("results not sorted desc by volume_usd_24h: %.2f came after %.2f", *row.VolumeUSD24h, *prev)
		}
		prev = row.VolumeUSD24h
	}

	// Cursor pagination: if there is a next page, fetching it should yield fresh pools.
	if res.HasNextPage && res.NextCursor != nil {
		next, err := client.Pools.AdvancedSearchByNetwork(ctx, network, &AdvancedSearchOptions{
			Limit:   3,
			SortBy:  "volume_usd_24h",
			SortDir: "desc",
			Cursor:  *res.NextCursor,
		})
		if err != nil {
			t.Fatalf("AdvancedSearchByNetwork (page 2) returned error: %v", err)
		}
		if len(next.Results) == 0 {
			t.Error("second page returned zero results despite has_next_page=true")
		}
		if len(next.Results) > 0 && len(res.Results) > 0 && next.Results[0].ID == res.Results[0].ID {
			t.Error("cursor pagination returned the same first pool as page one")
		}
	}
}

// TestPools_AdvancedSearch_Validation checks the per-network variant rejects an
// empty network ID.
func TestPools_AdvancedSearch_Validation(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	_, err := client.Pools.AdvancedSearchByNetwork(ctx, "", &AdvancedSearchOptions{})
	if err == nil || err.Error() != "network ID is required" {
		t.Errorf("expected 'network ID is required' error, got: %v", err)
	}
}

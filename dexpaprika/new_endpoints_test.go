package dexpaprika

import (
	"context"
	"testing"
)

func TestPoolsFilter(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	vol := 100000.0
	result, err := client.Pools.Filter(ctx, "ethereum", &PoolFilterOptions{
		Limit:        5,
		Volume24hMin: &vol,
	})
	if err != nil {
		t.Fatalf("Filter() error: %v", err)
	}
	if len(result.Results) == 0 {
		t.Fatal("Filter() returned no results")
	}
	pool := result.Results[0]
	if pool.ID == "" {
		t.Fatal("Pool missing ID")
	}
	if len(pool.Tokens) == 0 {
		t.Fatal("Pool missing tokens")
	}
	// Guard against the silent data loss this fix addresses: the filter endpoint
	// returns timeframe-split volume, so VolumeUSD24h must be populated.
	if pool.VolumeUSD24h == nil {
		t.Fatal("Pool missing volume_usd_24h (filter endpoint should return it)")
	}
	t.Logf("Got %d filtered pools, first: %s (vol24h=%.0f)", len(result.Results), pool.ID, *pool.VolumeUSD24h)
}

func TestPoolsFilterMultipleParams(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	vol := 50000.0
	txns := 10
	result, err := client.Pools.Filter(ctx, "ethereum", &PoolFilterOptions{
		Limit:        3,
		Volume24hMin: &vol,
		Txns24hMin:   &txns,
		SortBy:       "volume_24h",
		SortDir:      "desc",
	})
	if err != nil {
		t.Fatalf("Filter() error: %v", err)
	}
	if result == nil {
		t.Fatal("Filter() returned nil")
	}
	t.Logf("Got %d pools with multi-filter", len(result.Results))
}

func TestTokensGetTop(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	result, err := client.Tokens.GetTop(ctx, "ethereum", &TopTokensOptions{
		Limit: 5,
	})
	if err != nil {
		t.Fatalf("GetTop() error: %v", err)
	}
	if len(result.Tokens) == 0 {
		t.Fatal("GetTop() returned no tokens")
	}
	token := result.Tokens[0]
	if token.Address == "" {
		t.Fatal("Token missing address")
	}
	if token.Symbol == "" {
		t.Fatal("Token missing symbol")
	}
	t.Logf("Top token: %s at $%.4f", token.Symbol, *token.PriceUSD)
}

func TestTokensGetTopWithSort(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	result, err := client.Tokens.GetTop(ctx, "ethereum", &TopTokensOptions{
		Limit:   3,
		OrderBy: "volume_24h",
		Sort:    "asc",
	})
	if err != nil {
		t.Fatalf("GetTop() error: %v", err)
	}
	if result == nil {
		t.Fatal("GetTop() returned nil")
	}
	t.Logf("Got %d tokens (asc sort)", len(result.Tokens))
}

func TestTokensFilter(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	vol := 100000.0
	result, err := client.Tokens.Filter(ctx, "ethereum", &TokenFilterOptions{
		Limit:        5,
		Volume24hMin: &vol,
	})
	if err != nil {
		t.Fatalf("Filter() error: %v", err)
	}
	if len(result.Results) == 0 {
		t.Fatal("Filter() returned no results")
	}
	token := result.Results[0]
	if token.Address == "" {
		t.Fatal("Token missing address")
	}
	if token.Chain == "" {
		t.Fatal("Token missing chain")
	}
	t.Logf("Got %d filtered tokens, first: %s", len(result.Results), token.Address[:10])
}

func TestTokensFilterWithFDV(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	vol := 100000.0
	fdv := 1000000.0
	result, err := client.Tokens.Filter(ctx, "ethereum", &TokenFilterOptions{
		Limit:        3,
		Volume24hMin: &vol,
		FDVMin:       &fdv,
	})
	if err != nil {
		t.Fatalf("Filter() error: %v", err)
	}
	if result == nil {
		t.Fatal("Filter() returned nil")
	}
	t.Logf("Got %d tokens with FDV filter", len(result.Results))
}

func TestTokensGetMultiPrices(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	prices, err := client.Tokens.GetMultiPrices(ctx, "ethereum", []string{
		"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2", // WETH
		"0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", // USDC
	})
	if err != nil {
		t.Fatalf("GetMultiPrices() error: %v", err)
	}
	if len(prices) != 2 {
		t.Fatalf("Expected 2 prices, got %d", len(prices))
	}
	for _, p := range prices {
		if p.ID == "" {
			t.Fatal("Price missing ID")
		}
		if p.PriceUSD == nil {
			t.Fatalf("Price missing price_usd for %s", p.ID)
		}
		t.Logf("%s: $%.4f", p.ID[:10], *p.PriceUSD)
	}
}

func TestTokensGetMultiPricesSingle(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	prices, err := client.Tokens.GetMultiPrices(ctx, "ethereum", []string{
		"0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
	})
	if err != nil {
		t.Fatalf("GetMultiPrices() error: %v", err)
	}
	if len(prices) != 1 {
		t.Fatalf("Expected 1 price, got %d", len(prices))
	}
}

func TestTokensGetMultiPricesValidation(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	_, err := client.Tokens.GetMultiPrices(ctx, "ethereum", []string{})
	if err == nil {
		t.Fatal("Expected error for empty tokens list")
	}
}

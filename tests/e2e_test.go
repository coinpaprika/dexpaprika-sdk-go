package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/coinpaprika/dexpaprika-sdk-go/dexpaprika"
)

func setupTest(t *testing.T) (context.Context, *dexpaprika.Client) {
	t.Helper()
	client := dexpaprika.NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	t.Cleanup(cancel)
	return ctx, client
}

func TestNetworksService(t *testing.T) {
	ctx, client := setupTest(t)

	tests := []struct {
		name string
		fn   func(t *testing.T, ctx context.Context, client *dexpaprika.Client)
	}{
		{
			name: "List",
			fn: func(t *testing.T, ctx context.Context, client *dexpaprika.Client) {
				networks, err := client.Networks.List(ctx)
				if err != nil {
					t.Errorf("List() error = %v", err)
					return
				}
				if len(networks) <= 0 {
					t.Error("List() returned empty networks list")
				}
			},
		},
		{
			name: "ListDexes",
			fn: func(t *testing.T, ctx context.Context, client *dexpaprika.Client) {
				networks, err := client.Networks.List(ctx)
				if err != nil {
					t.Errorf("List() error = %v", err)
					return
				}

				if len(networks) <= 0 {
					t.Errorf("List() returned %d networks list", len(networks))
					return
				}
				networkID := networks[0].ID

				dexes, err := client.Networks.ListDexes(ctx, networkID, 0, 5)
				if err != nil {
					t.Errorf("ListDexes() error = %v", err)
					return
				}
				if dexes == nil || dexes.Dexes == nil {
					t.Error("ListDexes() returned nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn(t, ctx, client)
		})
	}
}

func TestPoolsService(t *testing.T) {
	ctx, client := setupTest(t)
	poolsOpts := &dexpaprika.ListOptions{Limit: 5}

	tests := []struct {
		name string
		fn   func(t *testing.T, ctx context.Context, client *dexpaprika.Client)
	}{
		{
			name: "List",
			fn: func(t *testing.T, ctx context.Context, client *dexpaprika.Client) {
				pools, err := client.Pools.List(ctx, poolsOpts) //nolint:staticcheck // Testing deprecated endpoint behavior
				// Expect a 410 Gone error due to API deprecation
				if err != nil {
					var apiErr *dexpaprika.APIError
					if errors.As(err, &apiErr) {
						if apiErr.StatusCode == 410 {
							// This is expected - the endpoint has been deprecated
							t.Logf("Expected deprecation error received: %s", apiErr.Message)
							return
						}
					}
					t.Errorf("List() error = %v", err)
					return
				}
				// If we somehow get here without an error, warn but don't fail
				if pools == nil || len(pools.Pools) == 0 {
					t.Error("List() returned empty pools list")
				} else {
					t.Logf("Warning: List() still works but is deprecated. Returned %d pools", len(pools.Pools))
				}
			},
		},
		{
			name: "ListByNetwork",
			fn: func(t *testing.T, ctx context.Context, client *dexpaprika.Client) {
				pools, err := client.Pools.ListByNetwork(ctx, "ethereum", poolsOpts)
				if err != nil {
					t.Errorf("ListByNetwork() error = %v", err)
					return
				}
				if len(pools.Pools) <= 0 {
					t.Errorf("ListByNetwork() returned %d pools list", len(pools.Pools))
				}

				networkPools, err := client.Pools.ListByNetwork(ctx, pools.Pools[0].Chain, poolsOpts)
				if err != nil {
					t.Errorf("ListByNetwork() error = %v", err)
					return
				}
				if networkPools == nil || len(networkPools.Pools) == 0 {
					t.Error("ListByNetwork() returned empty pools list")
				}
			},
		},
		{
			name: "GetDetails",
			fn: func(t *testing.T, ctx context.Context, client *dexpaprika.Client) {
				pools, err := client.Pools.ListByNetwork(ctx, "ethereum", poolsOpts)
				if err != nil || len(pools.Pools) == 0 {
					t.Skip("no pools available for testing")
				}

				pool := pools.Pools[0]
				details, err := client.Pools.GetDetails(ctx, pool.Chain, pool.ID, false)
				if err != nil {
					t.Errorf("GetDetails() error = %v", err)
					return
				}
				if details == nil {
					t.Error("GetDetails() returned nil")
				}
			},
		},
		{
			name: "GetOHLCV",
			fn: func(t *testing.T, ctx context.Context, client *dexpaprika.Client) {
				pools, err := client.Pools.ListByNetwork(ctx, "ethereum", poolsOpts)
				if err != nil || len(pools.Pools) == 0 {
					t.Skip("no pools available for testing")
				}

				now := time.Now()
				yesterday := now.Add(-24 * time.Hour)
				ohlcvOpts := &dexpaprika.OHLCVOptions{
					Start:    yesterday.Format("2006-01-02"),
					End:      now.Format("2006-01-02"),
					Interval: "1h",
					Limit:    24,
				}

				// Try several pools rather than only the first. Not every pool
				// has OHLCV coverage: on 2026-08-07 the top-ranked ethereum
				// pool was a Uniswap V4 pool (a 64-hex pool id, not a pair
				// address) and returned zero candles, while USDC/WETH on V3
				// returned 24 for the same window. Pinning the test to
				// whichever pool happens to rank first makes it fail on market
				// movement rather than on a real regression.
				var lastErr error
				for i, pool := range pools.Pools {
					if i >= 5 {
						break
					}
					ohlcv, err := client.Pools.GetOHLCV(ctx, pool.Chain, pool.ID, ohlcvOpts)
					if err != nil {
						lastErr = err
						continue
					}
					if len(ohlcv) > 0 {
						return
					}
				}
				if lastErr != nil {
					t.Errorf("GetOHLCV() error = %v", lastErr)
					return
				}
				t.Skip("none of the top pools reported OHLCV for the last 24h")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn(t, ctx, client)
		})
	}
}

func TestTokensService(t *testing.T) {
	ctx, client := setupTest(t)

	tests := []struct {
		name string
		fn   func(t *testing.T, ctx context.Context, client *dexpaprika.Client)
	}{
		{
			name: "GetDetails",
			fn: func(t *testing.T, ctx context.Context, client *dexpaprika.Client) {
				// Using WETH as fallback
				tokenChain := "ethereum"
				tokenAddress := "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2" // #nosec

				details, err := client.Tokens.GetDetails(ctx, tokenChain, tokenAddress)
				if err != nil {
					t.Errorf("GetDetails() error = %v", err)
					return
				}
				if details == nil || details.Symbol == "" {
					t.Error("GetDetails() returned invalid data")
				}
			},
		},
		{
			name: "GetPools",
			fn: func(t *testing.T, ctx context.Context, client *dexpaprika.Client) {
				tokenChain := "ethereum"
				tokenAddress := "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2" // #nosec

				tokenPoolsOpts := &dexpaprika.TokenPoolsOptions{
					ListOptions: &dexpaprika.ListOptions{
						Limit:   5,
						OrderBy: "volume_usd",
						Sort:    "desc",
					},
				}

				pools, err := client.Tokens.GetPools(ctx, tokenChain, tokenAddress, tokenPoolsOpts)
				if err != nil {
					t.Errorf("GetPools() error = %v", err)
					return
				}
				if pools == nil || pools.Pools == nil {
					t.Error("GetPools() returned nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn(t, ctx, client)
		})
	}
}

func TestSearchService(t *testing.T) {
	ctx, client := setupTest(t)

	tests := []struct {
		name  string
		query string
		fn    func(t *testing.T, ctx context.Context, client *dexpaprika.Client, query string)
	}{
		{
			name:  "Search",
			query: "ethereum",
			fn: func(t *testing.T, ctx context.Context, client *dexpaprika.Client, query string) {
				results, err := client.Search.Search(ctx, query)
				if err != nil {
					t.Errorf("Search() error = %v", err)
					return
				}
				if results == nil {
					t.Error("Search() returned nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn(t, ctx, client, tt.query)
		})
	}
}

func TestUtilsService(t *testing.T) {
	ctx, client := setupTest(t)

	tests := []struct {
		name string
		fn   func(t *testing.T, ctx context.Context, client *dexpaprika.Client)
	}{
		{
			name: "GetStats",
			fn: func(t *testing.T, ctx context.Context, client *dexpaprika.Client) {
				stats, err := client.Utils.GetStats(ctx)
				if err != nil {
					t.Errorf("GetStats() error = %v", err)
					return
				}
				if stats == nil {
					t.Error("GetStats() returned nil")
					return
				}
				if stats.Chains <= 0 {
					t.Error("GetStats() returned invalid chains count")
				}
				if stats.Pools <= 0 {
					t.Error("GetStats() returned invalid pools count")
				}
				if stats.Tokens <= 0 {
					t.Error("GetStats() returned invalid tokens count")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn(t, ctx, client)
		})
	}
}

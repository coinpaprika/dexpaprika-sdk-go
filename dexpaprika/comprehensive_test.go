package dexpaprika

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestAllEndpoints tests every endpoint defined in the OpenAPI specification
func TestAllEndpoints(t *testing.T) {
	// Create a mock server that handles all endpoints
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle each endpoint defined in the OpenAPI spec
		switch r.URL.Path {
		// Networks endpoints
		case "/networks":
			writeTestJSON(w, map[string]interface{}{
				"networks": []map[string]string{
					{"id": "ethereum", "name": "Ethereum", "symbol": "ETH"},
					{"id": "solana", "name": "Solana", "symbol": "SOL"},
					{"id": "bitcoin", "name": "Bitcoin", "symbol": "BTC"},
				},
			})

		// DEXes endpoints
		case "/networks/ethereum/dexes":
			writeTestJSON(w, map[string]interface{}{
				"dexes": []map[string]interface{}{
					{
						"id":          "uniswap_v2",
						"name":        "Uniswap V2",
						"website_url": "https://uniswap.org",
						"pool_count":  1500,
					},
					{
						"id":          "sushiswap",
						"name":        "SushiSwap",
						"website_url": "https://sushi.com",
						"pool_count":  800,
					},
				},
				"page_info": map[string]interface{}{
					"total_count":    50,
					"page":           0,
					"items_on_page":  2,
					"items_per_page": 100,
				},
			})

		// Pools endpoints
		case "/pools":
			writeTestJSON(w, map[string]interface{}{
				"pools": []map[string]interface{}{
					createMockPool("0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc", "ethereum"),
					createMockPool("0x0d4a11d5eeaac28ec3f61d100daf4d40471f1852", "ethereum"),
				},
				"page_info": createPageInfo(100, 0, 2, 100),
			})

		case "/networks/ethereum/pools":
			writeTestJSON(w, map[string]interface{}{
				"pools": []map[string]interface{}{
					createMockPool("0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc", "ethereum"),
					createMockPool("0x0d4a11d5eeaac28ec3f61d100daf4d40471f1852", "ethereum"),
				},
				"page_info": createPageInfo(50, 0, 2, 100),
			})

		case "/networks/ethereum/dexes/uniswap_v2/pools":
			writeTestJSON(w, map[string]interface{}{
				"pools": []map[string]interface{}{
					createMockPool("0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc", "ethereum"),
				},
				"page_info": createPageInfo(25, 0, 1, 100),
			})

		case "/networks/ethereum/pools/0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc":
			writeTestJSON(w, createMockPoolDetails("0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc", "ethereum"))

		case "/networks/ethereum/pools/0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc/ohlcv":
			writeTestJSON(w, []map[string]interface{}{
				{
					"time_open":  "2023-10-27T10:00:00Z",
					"time_close": "2023-10-27T11:00:00Z",
					"open":       1.23,
					"high":       1.25,
					"low":        1.20,
					"close":      1.22,
					"volume":     1000000,
				},
				{
					"time_open":  "2023-10-27T11:00:00Z",
					"time_close": "2023-10-27T12:00:00Z",
					"open":       1.22,
					"high":       1.27,
					"low":        1.21,
					"close":      1.26,
					"volume":     1200000,
				},
			})

		case "/networks/ethereum/pools/0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc/transactions":
			writeTestJSON(w, map[string]interface{}{
				"transactions": []map[string]interface{}{
					{
						"id":                      "0x1234567890abcdef1234567890abcdef",
						"log_index":               0,
						"transaction_index":       5,
						"pool_id":                 "0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc",
						"sender":                  "0x1234567890123456789012345678901234567890",
						"recipient":               "0x0987654321098765432109876543210987654321",
						"token_0":                 "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
						"token_1":                 "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
						"amount_0":                "100000000",
						"amount_1":                "50000000000000000000",
						"created_at_block_number": 12345678,
					},
				},
				"page_info": createPageInfo(100, 0, 1, 100),
			})

		// Tokens endpoints
		case "/networks/ethereum/tokens/0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48":
			writeTestJSON(w, createMockToken("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", "ethereum"))

		case "/networks/ethereum/tokens/0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48/pools":
			writeTestJSON(w, map[string]interface{}{
				"pools": []map[string]interface{}{
					createMockPool("0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc", "ethereum"),
				},
				"page_info": createPageInfo(20, 0, 1, 100),
			})

		// Search endpoint
		case "/search":
			query := r.URL.Query().Get("query")
			resp := map[string]interface{}{
				"tokens": []map[string]interface{}{},
				"pools":  []map[string]interface{}{},
				"dexes":  []map[string]interface{}{},
			}

			if query == "eth" {
				resp["tokens"] = []map[string]interface{}{
					{
						"id":       "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
						"name":     "Wrapped Ether",
						"symbol":   "WETH",
						"chain":    "ethereum",
						"decimals": 18,
					},
				}
				resp["dexes"] = []map[string]interface{}{
					{
						"id":       "uniswap-v3",
						"dex_id":   "uniswap-v3",
						"dex_name": "Uniswap V3",
						"chain":    "ethereum",
					},
				}
			}

			writeTestJSON(w, resp)

		// Stats endpoint
		case "/stats":
			writeTestJSON(w, map[string]interface{}{
				"chains":    15,
				"factories": 150,
				"pools":     2500,
				"tokens":    3500,
			})

		default:
			// If endpoint not handled, return 404
			w.WriteHeader(http.StatusNotFound)
			writeTestJSON(w, map[string]interface{}{
				"error": "Not found",
			})
		}
	}))
	defer server.Close()

	// Create client using our mock server
	client := NewClient()
	err := client.SetBaseURL(server.URL)
	if err != nil {
		t.Fatalf("Failed to set base URL: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test all endpoints one by one
	t.Run("GetNetworks", func(t *testing.T) {
		testGetNetworks(t, ctx, client)
	})

	t.Run("GetNetworkDexes", func(t *testing.T) {
		testGetNetworkDexes(t, ctx, client)
	})

	t.Run("GetTopPools", func(t *testing.T) {
		testGetTopPools(t, ctx, client)
	})

	t.Run("GetNetworkPools", func(t *testing.T) {
		testGetNetworkPools(t, ctx, client)
	})

	t.Run("GetDexPools", func(t *testing.T) {
		testGetDexPools(t, ctx, client)
	})

	t.Run("GetPoolDetails", func(t *testing.T) {
		testGetPoolDetails(t, ctx, client)
	})

	t.Run("GetPoolOHLCV", func(t *testing.T) {
		testGetPoolOHLCV(t, ctx, client)
	})

	t.Run("GetPoolTransactions", func(t *testing.T) {
		testGetPoolTransactions(t, ctx, client)
	})

	t.Run("GetTokenDetails", func(t *testing.T) {
		testGetTokenDetails(t, ctx, client)
	})

	t.Run("GetTokenPools", func(t *testing.T) {
		testGetTokenPools(t, ctx, client)
	})

	t.Run("Search", func(t *testing.T) {
		testSearch(t, ctx, client)
	})

	t.Run("GetStats", func(t *testing.T) {
		testGetStats(t, ctx, client)
	})
}

// Helper functions to test each endpoint
func testGetNetworks(t *testing.T, ctx context.Context, client *Client) {
	req, err := client.NewRequest(http.MethodGet, "/networks", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	var resp map[string]interface{}
	_, err = client.Do(ctx, req, &resp)
	if err != nil {
		t.Fatalf("Failed to get networks: %v", err)
	}

	networks, ok := resp["networks"].([]interface{})
	if !ok || len(networks) != 3 {
		t.Errorf("Expected 3 networks, got %v", networks)
	}
}

func testGetNetworkDexes(t *testing.T, ctx context.Context, client *Client) {
	req, err := client.NewRequest(http.MethodGet, "/networks/ethereum/dexes", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	var resp map[string]interface{}
	_, err = client.Do(ctx, req, &resp)
	if err != nil {
		t.Fatalf("Failed to get network dexes: %v", err)
	}

	dexes, ok := resp["dexes"].([]interface{})
	if !ok || len(dexes) != 2 {
		t.Errorf("Expected 2 dexes, got %v", dexes)
	}
}

func testGetTopPools(t *testing.T, ctx context.Context, client *Client) {
	req, err := client.NewRequest(http.MethodGet, "/pools", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	var resp map[string]interface{}
	_, err = client.Do(ctx, req, &resp)
	if err != nil {
		t.Fatalf("Failed to get top pools: %v", err)
	}

	pools, ok := resp["pools"].([]interface{})
	if !ok || len(pools) != 2 {
		t.Errorf("Expected 2 pools, got %v", pools)
	}
}

func testGetNetworkPools(t *testing.T, ctx context.Context, client *Client) {
	req, err := client.NewRequest(http.MethodGet, "/networks/ethereum/pools", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	var resp map[string]interface{}
	_, err = client.Do(ctx, req, &resp)
	if err != nil {
		t.Fatalf("Failed to get network pools: %v", err)
	}

	pools, ok := resp["pools"].([]interface{})
	if !ok || len(pools) != 2 {
		t.Errorf("Expected 2 pools, got %v", pools)
	}
}

func testGetDexPools(t *testing.T, ctx context.Context, client *Client) {
	req, err := client.NewRequest(http.MethodGet, "/networks/ethereum/dexes/uniswap_v2/pools", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	var resp map[string]interface{}
	_, err = client.Do(ctx, req, &resp)
	if err != nil {
		t.Fatalf("Failed to get dex pools: %v", err)
	}

	pools, ok := resp["pools"].([]interface{})
	if !ok || len(pools) != 1 {
		t.Errorf("Expected 1 pool, got %v", pools)
	}
}

func testGetPoolDetails(t *testing.T, ctx context.Context, client *Client) {
	poolAddress := "0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc"
	req, err := client.NewRequest(http.MethodGet, "/networks/ethereum/pools/"+poolAddress, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	var resp map[string]interface{}
	_, err = client.Do(ctx, req, &resp)
	if err != nil {
		t.Fatalf("Failed to get pool details: %v", err)
	}

	id, ok := resp["id"].(string)
	if !ok || id != poolAddress {
		t.Errorf("Expected pool ID %s, got %v", poolAddress, id)
	}
}

func testGetPoolOHLCV(t *testing.T, ctx context.Context, client *Client) {
	poolAddress := "0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc"
	req, err := client.NewRequest(http.MethodGet, "/networks/ethereum/pools/"+poolAddress+"/ohlcv?start=1741507640", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	var resp []map[string]interface{}
	_, err = client.Do(ctx, req, &resp)
	if err != nil {
		t.Fatalf("Failed to get pool OHLCV: %v", err)
	}

	if len(resp) != 2 {
		t.Errorf("Expected 2 OHLCV records, got %d", len(resp))
	}
}

func testGetPoolTransactions(t *testing.T, ctx context.Context, client *Client) {
	poolAddress := "0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc"
	req, err := client.NewRequest(http.MethodGet, "/networks/ethereum/pools/"+poolAddress+"/transactions", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	var resp map[string]interface{}
	_, err = client.Do(ctx, req, &resp)
	if err != nil {
		t.Fatalf("Failed to get pool transactions: %v", err)
	}

	transactions, ok := resp["transactions"].([]interface{})
	if !ok || len(transactions) != 1 {
		t.Errorf("Expected 1 transaction, got %v", transactions)
	}
}

func testGetTokenDetails(t *testing.T, ctx context.Context, client *Client) {
	tokenAddress := "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"
	req, err := client.NewRequest(http.MethodGet, "/networks/ethereum/tokens/"+tokenAddress, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	var resp map[string]interface{}
	_, err = client.Do(ctx, req, &resp)
	if err != nil {
		t.Fatalf("Failed to get token details: %v", err)
	}

	id, ok := resp["id"].(string)
	if !ok || id != tokenAddress {
		t.Errorf("Expected token ID %s, got %v", tokenAddress, id)
	}
}

func testGetTokenPools(t *testing.T, ctx context.Context, client *Client) {
	tokenAddress := "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"
	req, err := client.NewRequest(http.MethodGet, "/networks/ethereum/tokens/"+tokenAddress+"/pools", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	var resp map[string]interface{}
	_, err = client.Do(ctx, req, &resp)
	if err != nil {
		t.Fatalf("Failed to get token pools: %v", err)
	}

	pools, ok := resp["pools"].([]interface{})
	if !ok || len(pools) != 1 {
		t.Errorf("Expected 1 pool, got %v", pools)
	}
}

func testSearch(t *testing.T, ctx context.Context, client *Client) {
	req, err := client.NewRequest(http.MethodGet, "/search?query=eth", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	var resp SearchResult
	_, err = client.Do(ctx, req, &resp)
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	if len(resp.Tokens) == 0 && len(resp.Pools) == 0 && len(resp.Dexes) == 0 {
		t.Errorf("Expected at least one result in tokens, pools, or dexes, but all were empty")
	}
}

func testGetStats(t *testing.T, ctx context.Context, client *Client) {
	req, err := client.NewRequest(http.MethodGet, "/stats", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	var resp map[string]interface{}
	_, err = client.Do(ctx, req, &resp)
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	chains, ok := resp["chains"].(float64)
	if !ok || int(chains) != 15 {
		t.Errorf("Expected 15 chains, got %v", chains)
	}
}

// Helper functions to create mock data objects
func createMockPool(id string, chain string) map[string]interface{} {
	return map[string]interface{}{
		"id":                        id,
		"dex_id":                    "uniswap-v2",
		"dex_name":                  "Uniswap V2",
		"chain":                     chain,
		"volume_usd":                1500000.0,
		"created_at":                "2021-05-05T21:42:11.000Z",
		"created_at_block_number":   12376729,
		"transactions":              8808,
		"price_usd":                 1.000038420407992,
		"last_price_change_usd_5m":  0.2817344833094529,
		"last_price_change_usd_1h":  -0.11886943575265935,
		"last_price_change_usd_24h": 0.06894442872697064,
		"fee":                       0.003,
		"tokens": []map[string]interface{}{
			{
				"id":       "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
				"name":     "USD Coin",
				"symbol":   "USDC",
				"chain":    chain,
				"decimals": 6,
				"added_at": "2024-12-02T13:00:16.000Z",
			},
			{
				"id":       "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
				"name":     "Wrapped Ether",
				"symbol":   "WETH",
				"chain":    chain,
				"decimals": 18,
				"added_at": "2024-12-02T13:00:16.000Z",
			},
		},
	}
}

func createMockPoolDetails(id string, chain string) map[string]interface{} {
	return map[string]interface{}{
		"id":                      id,
		"created_at_block_number": 12376729,
		"chain":                   chain,
		"created_at":              "2021-05-05T21:42:11.000Z",
		"factory_id":              "0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f",
		"dex_id":                  "uniswap-v2",
		"dex_name":                "Uniswap V2",
		"tokens": []map[string]interface{}{
			{
				"id":       "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
				"name":     "USD Coin",
				"symbol":   "USDC",
				"chain":    chain,
				"decimals": 6,
				"added_at": "2024-12-02T13:00:16.000Z",
				"fdv":      42000000000,
			},
			{
				"id":       "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
				"name":     "Wrapped Ether",
				"symbol":   "WETH",
				"chain":    chain,
				"decimals": 18,
				"added_at": "2024-12-02T13:00:16.000Z",
				"fdv":      350000000000,
			},
		},
		"last_price":     0.06,
		"last_price_usd": 0.06,
		"fee":            0.003,
		"price_time":     "2023-06-15T10:00:00Z",
		"24h":            createMockTimeIntervalMetrics(1500000.0),
		"6h":             createMockTimeIntervalMetrics(500000.0),
		"1h":             createMockTimeIntervalMetrics(100000.0),
		"30m":            createMockTimeIntervalMetrics(50000.0),
		"15m":            createMockTimeIntervalMetrics(25000.0),
		"5m":             createMockTimeIntervalMetrics(10000.0),
	}
}

func createMockTimeIntervalMetrics(volumeUsd float64) map[string]interface{} {
	buyUsd := volumeUsd / 2
	sellUsd := volumeUsd / 2
	buys := int(volumeUsd / 2000)
	sells := int(volumeUsd / 2000)

	return map[string]interface{}{
		"last_price_usd_change": 0.15,
		"volume_usd":            volumeUsd,
		"buy_usd":               buyUsd,
		"sell_usd":              sellUsd,
		"sells":                 sells,
		"buys":                  buys,
		"txns":                  buys + sells,
	}
}

func createMockToken(id string, chain string) map[string]interface{} {
	return map[string]interface{}{
		"id":           id,
		"name":         "USD Coin",
		"symbol":       "USDC",
		"chain":        chain,
		"decimals":     6,
		"total_supply": 39000000000000000,
		"description":  "USD Coin (USDC) is a stablecoin redeemable on a 1:1 basis for US dollars.",
		"website":      "https://www.centre.io/usdc",
		"explorer":     "https://etherscan.io/token/0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
		"added_at":     "2024-09-11T04:37:20Z",
		"summary":      createMockTokenSummary(),
		"last_updated": "2025-02-25T13:44:45.699686371Z",
	}
}

func createMockTokenSummary() map[string]interface{} {
	return map[string]interface{}{
		"price_usd":     1.0,
		"fdv":           39000000000,
		"liquidity_usd": 25796064.003077608,
		"24h":           createMockTimeIntervalMetrics(84000000.0),
		"6h":            createMockTimeIntervalMetrics(20000000.0),
		"1h":            createMockTimeIntervalMetrics(1500000.0),
		"30m":           createMockTimeIntervalMetrics(700000.0),
		"15m":           createMockTimeIntervalMetrics(280000.0),
		"5m":            createMockTimeIntervalMetrics(50000.0),
	}
}

func createPageInfo(totalItems int, page int, itemsOnPage int, itemsPerPage int) map[string]interface{} {
	return map[string]interface{}{
		"total_count":    totalItems,
		"page":           page,
		"items_on_page":  itemsOnPage,
		"items_per_page": itemsPerPage,
	}
}

func writeTestJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

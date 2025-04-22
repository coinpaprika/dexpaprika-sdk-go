package dexpaprika

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPools_List(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test getting top pools
	poolsOpts := &ListOptions{
		Limit:   5,
		OrderBy: "volume_usd",
		Sort:    "desc",
	}
	pools, err := client.Pools.List(ctx, poolsOpts)
	if err != nil {
		t.Fatalf("Pools.List returned error: %v", err)
	}

	if pools == nil {
		t.Fatal("Pools.List returned nil, expected a PoolList")
	}

	if len(pools.Pools) == 0 {
		t.Error("Pools.List returned empty list, expected some pools")
	}

	// Check basic properties of pools
	for _, pool := range pools.Pools {
		if pool.ID == "" {
			t.Error("Pools.List returned pool with empty ID")
		}
		if pool.Chain == "" {
			t.Error("Pools.List returned pool with empty Chain")
		}
		if pool.DexName == "" {
			t.Error("Pools.List returned pool with empty DexName")
		}
	}
}

func TestPools_ListByNetwork(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test getting network-specific pools (using Ethereum as example)
	networkID := "ethereum"
	poolsOpts := &ListOptions{
		Limit:   5,
		OrderBy: "volume_usd",
		Sort:    "desc",
	}

	pools, err := client.Pools.ListByNetwork(ctx, networkID, poolsOpts)
	if err != nil {
		t.Fatalf("Pools.ListByNetwork returned error: %v", err)
	}

	if pools == nil {
		t.Fatal("Pools.ListByNetwork returned nil, expected a PoolList")
	}

	// All pools should be on the specified network
	for _, pool := range pools.Pools {
		if pool.Chain != networkID {
			t.Errorf("Pools.ListByNetwork returned pool with wrong chain: got %s, want %s", pool.Chain, networkID)
		}
	}
}

func TestPools_ListByDex(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test getting DEX-specific pools (using Uniswap V3 on Ethereum as example)
	networkID := "ethereum"
	dexID := "uniswap_v3"
	poolsOpts := &ListOptions{
		Limit:   5,
		OrderBy: "volume_usd",
		Sort:    "desc",
	}

	pools, err := client.Pools.ListByDex(ctx, networkID, dexID, poolsOpts)
	if err != nil {
		t.Fatalf("Pools.ListByDex returned error: %v", err)
	}

	if pools == nil {
		t.Fatal("Pools.ListByDex returned nil, expected a PoolList")
	}

	// All pools should be on the specified network and DEX
	for _, pool := range pools.Pools {
		if pool.Chain != networkID {
			t.Errorf("Pools.ListByDex returned pool with wrong chain: got %s, want %s", pool.Chain, networkID)
		}
		// DEX names might not exactly match the DEX ID, so we don't check this strictly
	}
}

func TestPools_GetDetails(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use a well-known Ethereum USDC-WETH Uniswap V3 pool
	networkID := "ethereum"
	poolID := "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640"

	// Test getting pool details
	details, err := client.Pools.GetDetails(ctx, networkID, poolID, false)
	if err != nil {
		t.Fatalf("Pools.GetDetails returned error: %v", err)
	}

	if details == nil {
		t.Fatal("Pools.GetDetails returned nil, expected pool details")
	}

	// Check the pool details contain relevant information
	if details.ID != poolID {
		t.Errorf("Pools.GetDetails returned wrong pool ID: got %s, want %s", details.ID, poolID)
	}
	if details.Chain != networkID {
		t.Errorf("Pools.GetDetails returned wrong chain: got %s, want %s", details.Chain, networkID)
	}

	// Verify tokens
	hasUSDC := false
	hasWETH := false
	for _, token := range details.Tokens {
		if token.Symbol == "USDC" {
			hasUSDC = true
		}
		if token.Symbol == "WETH" {
			hasWETH = true
		}
	}

	if !hasUSDC {
		t.Error("Pool details missing USDC token")
	}
	if !hasWETH {
		t.Error("Pool details missing WETH token")
	}
}

func TestPools_GetOHLCV(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use a well-known Ethereum USDC-WETH Uniswap V3 pool
	networkID := "ethereum"
	poolID := "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640"

	// Use specific date format for the API
	// Get yesterday and today in the format YYYY-MM-DD
	today := time.Now()
	yesterday := today.AddDate(0, 0, -1)

	// Format the dates properly
	startDate := yesterday.Format("2006-01-02")
	endDate := today.Format("2006-01-02")

	t.Logf("Getting OHLCV data from %s to %s", startDate, endDate)

	// Test getting OHLCV data with specific date format
	ohlcvOpts := &OHLCVOptions{
		Start:    startDate,
		End:      endDate,
		Interval: "1h",
		Limit:    3,
	}

	ohlcv, err := client.Pools.GetOHLCV(ctx, networkID, poolID, ohlcvOpts)
	if err != nil {
		t.Fatalf("Pools.GetOHLCV returned error: %v", err)
	}

	// The API should return data for this timeframe
	if len(ohlcv) == 0 {
		t.Log("Warning: No OHLCV data available for the selected timeframe")
	} else {
		t.Logf("Retrieved %d OHLCV records", len(ohlcv))

		// Check basic properties of OHLCV data
		for _, record := range ohlcv {
			if record.TimeOpen == "" {
				t.Error("Pools.GetOHLCV returned record with empty TimeOpen")
			}
		}
	}
}

// TestPools_ListWithMock tests the List method with a mock server
func TestPools_ListWithMock(t *testing.T) {
	// Define test cases
	tests := []struct {
		name        string
		opts        *ListOptions
		response    string
		statusCode  int
		expectError bool
		poolCount   int
	}{
		{
			name:        "successful response with default options",
			opts:        &ListOptions{},
			response:    `{"pools": [{"id": "0x123", "dex_id": "uniswap", "chain": "ethereum"}, {"id": "0x456", "dex_id": "uniswap", "chain": "ethereum"}]}`,
			statusCode:  http.StatusOK,
			expectError: false,
			poolCount:   2,
		},
		{
			name:        "successful response with pagination",
			opts:        &ListOptions{Page: 2, Limit: 5},
			response:    `{"pools": [{"id": "0x789", "dex_id": "uniswap", "chain": "ethereum"}]}`,
			statusCode:  http.StatusOK,
			expectError: false,
			poolCount:   1,
		},
		{
			name:        "successful response with sorting",
			opts:        &ListOptions{OrderBy: "volume_usd", Sort: "desc"},
			response:    `{"pools": [{"id": "0xabc", "dex_id": "uniswap", "chain": "ethereum", "volume_usd": 1000000}]}`,
			statusCode:  http.StatusOK,
			expectError: false,
			poolCount:   1,
		},
		{
			name:        "server error",
			opts:        &ListOptions{},
			response:    `{"error": "Internal server error"}`,
			statusCode:  http.StatusInternalServerError,
			expectError: true,
			poolCount:   0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check that the request is for the pools endpoint
				if r.URL.Path != "/pools" {
					t.Errorf("Expected request to '/pools', got '%s'", r.URL.Path)
				}

				// Check query parameters if options are provided
				if tc.opts != nil {
					// Check page parameter
					if tc.opts.Page > 0 {
						page := r.URL.Query().Get("page")
						expectedPage := fmt.Sprintf("%d", tc.opts.Page)
						if page != expectedPage {
							t.Errorf("Expected page parameter to be '%s', got '%s'", expectedPage, page)
						}
					}

					// Check limit parameter
					if tc.opts.Limit > 0 {
						limit := r.URL.Query().Get("limit")
						expectedLimit := fmt.Sprintf("%d", tc.opts.Limit)
						if limit != expectedLimit {
							t.Errorf("Expected limit parameter to be '%s', got '%s'", expectedLimit, limit)
						}
					}

					// Check orderBy parameter
					if tc.opts.OrderBy != "" {
						orderBy := r.URL.Query().Get("order_by")
						if orderBy != tc.opts.OrderBy {
							t.Errorf("Expected order_by parameter to be '%s', got '%s'", tc.opts.OrderBy, orderBy)
						}
					}

					// Check sort parameter
					if tc.opts.Sort != "" {
						sort := r.URL.Query().Get("sort")
						if sort != tc.opts.Sort {
							t.Errorf("Expected sort parameter to be '%s', got '%s'", tc.opts.Sort, sort)
						}
					}
				}

				// Set response headers
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.statusCode)
				fmt.Fprintln(w, tc.response)
			}))
			defer server.Close()

			// Create a client that uses the test server
			client := NewClient(
				WithBaseURL(server.URL),
				WithRetryConfig(0, 1*time.Millisecond, 1*time.Millisecond), // No retries for faster tests
			)

			// Create a context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Call the List method
			poolsResp, err := client.Pools.List(ctx, tc.opts)

			// Check error
			if tc.expectError && err == nil {
				t.Error("Expected an error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// If we don't expect an error, check the results
			if !tc.expectError && err == nil {
				if poolsResp == nil {
					t.Fatal("Expected non-nil pools response but got nil")
				}

				if len(poolsResp.Pools) != tc.poolCount {
					t.Errorf("Expected %d pools, got %d", tc.poolCount, len(poolsResp.Pools))
				}
			}
		})
	}
}

// TestPools_GetTransactionsWithMock tests the GetTransactions method with a mock server
func TestPools_GetTransactionsWithMock(t *testing.T) {
	// Define test cases
	tests := []struct {
		name             string
		network          string
		poolAddress      string
		page             int
		limit            int
		cursor           string
		response         string
		statusCode       int
		expectError      bool
		transactionCount int
	}{
		{
			name:             "successful response with default options",
			network:          "ethereum",
			poolAddress:      "0x123456789abcdef",
			page:             0,
			limit:            10,
			cursor:           "",
			response:         `{"transactions": [{"id": "0xabc1", "pool_id": "0x123456789abcdef"}, {"id": "0xabc2", "pool_id": "0x123456789abcdef"}]}`,
			statusCode:       http.StatusOK,
			expectError:      false,
			transactionCount: 2,
		},
		{
			name:             "successful response with pagination",
			network:          "ethereum",
			poolAddress:      "0x123456789abcdef",
			page:             2,
			limit:            10,
			cursor:           "",
			response:         `{"transactions": [{"id": "0xdef1", "pool_id": "0x123456789abcdef"}]}`,
			statusCode:       http.StatusOK,
			expectError:      false,
			transactionCount: 1,
		},
		{
			name:             "successful response with cursor",
			network:          "ethereum",
			poolAddress:      "0x123456789abcdef",
			page:             0,
			limit:            10,
			cursor:           "0xabc2",
			response:         `{"transactions": [{"id": "0xdef1", "pool_id": "0x123456789abcdef"}]}`,
			statusCode:       http.StatusOK,
			expectError:      false,
			transactionCount: 1,
		},
		{
			name:             "network not found",
			network:          "invalid",
			poolAddress:      "0x123456789abcdef",
			page:             0,
			limit:            10,
			cursor:           "",
			response:         `{"error": "Network not found"}`,
			statusCode:       http.StatusNotFound,
			expectError:      true,
			transactionCount: 0,
		},
		{
			name:             "pool not found",
			network:          "ethereum",
			poolAddress:      "0xinvalid",
			page:             0,
			limit:            10,
			cursor:           "",
			response:         `{"error": "Pool not found"}`,
			statusCode:       http.StatusNotFound,
			expectError:      true,
			transactionCount: 0,
		},
		{
			name:             "server error",
			network:          "ethereum",
			poolAddress:      "0x123456789abcdef",
			page:             0,
			limit:            10,
			cursor:           "",
			response:         `{"error": "Internal server error"}`,
			statusCode:       http.StatusInternalServerError,
			expectError:      true,
			transactionCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check that the request is for the correct network and pool endpoint
				expectedPath := fmt.Sprintf("/networks/%s/pools/%s/transactions", tc.network, tc.poolAddress)
				if r.URL.Path != expectedPath {
					t.Errorf("Expected request to '%s', got '%s'", expectedPath, r.URL.Path)
				}

				// Check query parameters
				// Check page parameter
				if tc.page > 0 {
					page := r.URL.Query().Get("page")
					expectedPage := fmt.Sprintf("%d", tc.page)
					if page != expectedPage {
						t.Errorf("Expected page parameter to be '%s', got '%s'", expectedPage, page)
					}
				}

				// Check limit parameter
				if tc.limit > 0 {
					limit := r.URL.Query().Get("limit")
					expectedLimit := fmt.Sprintf("%d", tc.limit)
					if limit != expectedLimit {
						t.Errorf("Expected limit parameter to be '%s', got '%s'", expectedLimit, limit)
					}
				}

				// Check cursor parameter
				if tc.cursor != "" {
					cursor := r.URL.Query().Get("cursor")
					if cursor != tc.cursor {
						t.Errorf("Expected cursor parameter to be '%s', got '%s'", tc.cursor, cursor)
					}
				}

				// Set response headers
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.statusCode)
				fmt.Fprintln(w, tc.response)
			}))
			defer server.Close()

			// Create a client that uses the test server
			client := NewClient(
				WithBaseURL(server.URL),
				WithRetryConfig(0, 1*time.Millisecond, 1*time.Millisecond), // No retries for faster tests
			)

			// Create a context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Call the GetTransactions method
			transactionsResp, err := client.Pools.GetTransactions(ctx, tc.network, tc.poolAddress, tc.page, tc.limit, tc.cursor)

			// Check error
			if tc.expectError && err == nil {
				t.Error("Expected an error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// If we don't expect an error, check the results
			if !tc.expectError && err == nil {
				if transactionsResp == nil {
					t.Fatal("Expected non-nil transactions response but got nil")
				}

				if len(transactionsResp.Transactions) != tc.transactionCount {
					t.Errorf("Expected %d transactions, got %d", tc.transactionCount, len(transactionsResp.Transactions))
				}
			}
		})
	}
}

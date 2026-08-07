package dexpaprika

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPools_List(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(3, 1*time.Second, 8*time.Second),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test getting top pools - this should now return a 410 Gone error
	poolsOpts := &ListOptions{
		Limit:   5,
		OrderBy: "volume_usd",
		Sort:    "desc",
	}
	pools, err := client.Pools.List(ctx, poolsOpts)

	// Expect a 410 Gone error due to API deprecation
	if err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) {
			if apiErr.StatusCode == 410 {
				// This is expected - the endpoint has been deprecated
				t.Logf("Expected deprecation error received: %s", apiErr.Message)
				return
			}
		}
		t.Fatalf("Pools.List returned unexpected error: %v", err)
	}

	// If we somehow get here without an error, warn but don't fail
	// (in case the endpoint is temporarily still working)
	if pools != nil {
		t.Logf("Warning: Pools.List still works but is deprecated. Returned %d pools", len(pools.Pools))
	}
}

func TestPools_ListByNetwork(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(3, 1*time.Second, 8*time.Second),
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
		WithRetryConfig(3, 1*time.Second, 8*time.Second),
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

	// /networks/{network}/dexes/{dex}/pools was removed and answers 410. This
	// test asserted a successful listing until 2026-08-07, when it started
	// failing on main with no change on our side. The SDK behaviour is right:
	// it surfaces the removal as a typed error carrying the replacement path.
	// The expectation was what went stale.
	pools, err := client.Pools.ListByDex(ctx, networkID, dexID, poolsOpts)
	var deprecated *APIError
	if !errors.As(err, &deprecated) || deprecated.StatusCode != 410 {
		t.Fatalf("Pools.ListByDex error = %v, want a 410 APIError", err)
	}
	if !strings.Contains(deprecated.Replacement, "pools/search") {
		t.Errorf("replacement = %q, want it to point at pools/search", deprecated.Replacement)
	}
	if pools != nil {
		t.Errorf("Pools.ListByDex returned %v alongside the removal error, want nil", pools)
	}
}

func TestPools_GetDetails(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(3, 1*time.Second, 8*time.Second),
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
		WithRetryConfig(3, 1*time.Second, 8*time.Second),
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
func TestPools_ListReturnsGoneClientSide(t *testing.T) {
	// List must fail client-side without issuing any HTTP request: the global
	// /pools endpoint was removed (HTTP 410) in API v1.3.0.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("List should not issue an HTTP request, but it hit %s", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithRetryConfig(0, 1*time.Millisecond, 1*time.Millisecond),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.Pools.List(ctx, &ListOptions{Limit: 5})
	if err == nil {
		t.Fatal("expected a deprecation error from List, got nil")
	}
	if resp != nil {
		t.Errorf("expected a nil response, got %+v", resp)
	}
	if !errors.Is(err, ErrGone) {
		t.Errorf("expected errors.Is(err, ErrGone) to be true, got %v", err)
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected an *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusGone {
		t.Errorf("expected status code 410, got %d", apiErr.StatusCode)
	}
	if apiErr.Replacement == "" {
		t.Error("expected a replacement endpoint hint on the error")
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

func TestPools_ValidationErrors(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Test ListByNetwork with empty network ID
	_, err := client.Pools.ListByNetwork(ctx, "", nil)
	if err == nil || err.Error() != "network ID is required" {
		t.Errorf("Expected 'network ID is required' error, got: %v", err)
	}

	// Test ListByDex with empty network ID
	_, err = client.Pools.ListByDex(ctx, "", "uniswap_v3", nil)
	if err == nil || err.Error() != "network ID is required" {
		t.Errorf("Expected 'network ID is required' error, got: %v", err)
	}

	// Test ListByDex with empty dex ID
	_, err = client.Pools.ListByDex(ctx, "ethereum", "", nil)
	if err == nil || err.Error() != "dex ID is required" {
		t.Errorf("Expected 'dex ID is required' error, got: %v", err)
	}

	// Test GetDetails with empty network ID
	_, err = client.Pools.GetDetails(ctx, "", "0xpool", false)
	if err == nil || err.Error() != "network ID is required" {
		t.Errorf("Expected 'network ID is required' error, got: %v", err)
	}

	// Test GetDetails with empty pool address
	_, err = client.Pools.GetDetails(ctx, "ethereum", "", false)
	if err == nil || err.Error() != "pool address is required" {
		t.Errorf("Expected 'pool address is required' error, got: %v", err)
	}

	// Test GetOHLCV with empty network ID
	_, err = client.Pools.GetOHLCV(ctx, "", "0xpool", nil)
	if err == nil || err.Error() != "network ID is required" {
		t.Errorf("Expected 'network ID is required' error, got: %v", err)
	}

	// Test GetOHLCV with empty pool address
	_, err = client.Pools.GetOHLCV(ctx, "ethereum", "", nil)
	if err == nil || err.Error() != "pool address is required" {
		t.Errorf("Expected 'pool address is required' error, got: %v", err)
	}

	// Test GetTransactions with empty network ID
	_, err = client.Pools.GetTransactions(ctx, "", "0xpool", 0, 10, "")
	if err == nil || err.Error() != "network ID is required" {
		t.Errorf("Expected 'network ID is required' error, got: %v", err)
	}

	// Test GetTransactions with empty pool address
	_, err = client.Pools.GetTransactions(ctx, "ethereum", "", 0, 10, "")
	if err == nil || err.Error() != "pool address is required" {
		t.Errorf("Expected 'pool address is required' error, got: %v", err)
	}
}

// TestPools_FilterSendsPriceChangeBounds pins the wire format of the
// price-change filters. They are worth a test because a mistake here is
// invisible from the response: /pools/search ignores query parameters it does
// not recognize and answers 200, so a misspelled or dropped bound returns a
// full, plausible, unfiltered result set.
func TestPools_FilterSendsPriceChangeBounds(t *testing.T) {
	var got string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"results":[],"has_next_page":false}`)
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithRetryConfig(0, 1*time.Millisecond, 1*time.Millisecond),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	min1h := 50.0
	max5m := -20.0
	_, err := client.Pools.Filter(ctx, "ethereum", &PoolFilterOptions{
		SortBy:           "price_change_percentage_1h",
		PriceChange1hMin: &min1h,
		PriceChange5mMax: &max5m,
	})
	if err != nil {
		t.Fatalf("Filter returned error: %v", err)
	}

	for _, want := range []string{
		"order_by=price_change_percentage_1h",
		"price_change_percentage_1h_min=50",
		// Negative bounds are the whole point of a max on a price change, and
		// they are also where a naive %f formatter would produce -20.000000.
		"price_change_percentage_5m_max=-20",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("query %q is missing %q", got, want)
		}
	}
}

// TestSortFieldWindowsArePoolOnly guards the asymmetry between the two search
// endpoints. Verified against api.dexpaprika.com on 2026-08-07: /pools/search
// lists the three shorter windows in its 400 body, /tokens/search does not, and
// token rows carry no price_change_percentage_5m field at all. Sending one to
// tokens/search is a 400, so the fallback to volume_usd_24h is the correct
// behaviour rather than an oversight.
func TestSortFieldWindowsArePoolOnly(t *testing.T) {
	for _, field := range []string{
		"price_change_percentage_6h",
		"price_change_percentage_1h",
		"price_change_percentage_5m",
	} {
		if got := mapPoolSortField(field); got != field {
			t.Errorf("mapPoolSortField(%q) = %q, want it passed through", field, got)
		}
		if got := mapTokenSortField(field); got != "volume_usd_24h" {
			t.Errorf("mapTokenSortField(%q) = %q, want the volume_usd_24h fallback", field, got)
		}
	}
}

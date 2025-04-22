package dexpaprika

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSearch_Search(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a context with longer timeout (30 seconds instead of 10)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test searching for a common term
	query := "ethereum"
	results, err := client.Search.Search(ctx, query)
	if err != nil {
		t.Fatalf("Search.Search returned error: %v", err)
	}

	if results == nil {
		t.Fatal("Search.Search returned nil, expected search results")
	}

	// Check that we got some results back
	if len(results.Tokens) == 0 && len(results.Pools) == 0 && len(results.Dexes) == 0 {
		t.Error("Search.Search returned empty results for query 'ethereum'")
	}

	// Test specific query handling with individual timeouts
	queries := []struct {
		query   string
		timeout time.Duration
		desc    string
	}{
		{"uniswap", 10 * time.Second, "should find DEXes"},
		{"bitcoin", 10 * time.Second, "should find tokens"},
		{"eth", 15 * time.Second, "should find tokens"},              // Allow more time for this common term
		{"0xc02aaa39b2", 15 * time.Second, "partial address search"}, // Allow more time for address search
	}

	for _, q := range queries {
		t.Run("Query:"+q.query, func(t *testing.T) {
			// Create a separate context with appropriate timeout for each query
			queryCtx, queryCancel := context.WithTimeout(context.Background(), q.timeout)
			defer queryCancel()

			results, err := client.Search.Search(queryCtx, q.query)
			if err != nil {
				t.Errorf("Search.Search(%q) returned error: %v", q.query, err)
				return
			}

			if results == nil {
				t.Errorf("Search.Search(%q) returned nil results", q.query)
				return
			}

			t.Logf("Query %q (%s) returned: %d tokens, %d pools, %d dexes",
				q.query, q.desc, len(results.Tokens), len(results.Pools), len(results.Dexes))
		})
	}
}

func TestSearch_InvalidQuery(t *testing.T) {
	client := NewClient()

	// Test with empty query
	ctx := context.Background()
	_, err := client.Search.Search(ctx, "")

	if err == nil {
		t.Fatal("Expected error for empty query, got nil")
	}
}

// TestSearch_SearchWithMock tests the Search method using a mock server
func TestSearch_SearchWithMock(t *testing.T) {
	// Define test cases
	tests := []struct {
		name           string
		query          string
		response       string
		statusCode     int
		expectError    bool
		expectedTokens int
		expectedPools  int
		expectedDexes  int
	}{
		{
			name:           "valid query with results",
			query:          "uniswap",
			response:       `{"tokens": [{"name": "Uniswap", "symbol": "UNI"}], "pools": [{"name": "UNI-ETH"}], "dexes": [{"name": "Uniswap V2"}, {"name": "Uniswap V3"}]}`,
			statusCode:     http.StatusOK,
			expectError:    false,
			expectedTokens: 1,
			expectedPools:  1,
			expectedDexes:  2,
		},
		{
			name:           "valid query with no results",
			query:          "nonexistent",
			response:       `{"tokens": [], "pools": [], "dexes": []}`,
			statusCode:     http.StatusOK,
			expectError:    false,
			expectedTokens: 0,
			expectedPools:  0,
			expectedDexes:  0,
		},
		{
			name:           "empty query",
			query:          "",
			response:       `{"error": "Invalid query param"}`,
			statusCode:     http.StatusBadRequest,
			expectError:    true,
			expectedTokens: 0,
			expectedPools:  0,
			expectedDexes:  0,
		},
		{
			name:           "server error",
			query:          "error",
			response:       `{"error": "Internal server error"}`,
			statusCode:     http.StatusInternalServerError,
			expectError:    true,
			expectedTokens: 0,
			expectedPools:  0,
			expectedDexes:  0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check that the request is for the search endpoint
				if r.URL.Path != "/search" {
					t.Errorf("Expected request to '/search', got '%s'", r.URL.Path)
				}

				// Check that the query parameter is correct
				query := r.URL.Query().Get("query")
				if query != tc.query {
					t.Errorf("Expected query parameter to be '%s', got '%s'", tc.query, query)
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

			// Perform the search
			result, err := client.Search.Search(ctx, tc.query)

			// Check error
			if tc.expectError && err == nil {
				t.Error("Expected an error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// If we don't expect an error, check the results
			if !tc.expectError {
				if result == nil {
					t.Fatal("Expected non-nil result but got nil")
				}

				if len(result.Tokens) != tc.expectedTokens {
					t.Errorf("Expected %d tokens, got %d", tc.expectedTokens, len(result.Tokens))
				}

				if len(result.Pools) != tc.expectedPools {
					t.Errorf("Expected %d pools, got %d", tc.expectedPools, len(result.Pools))
				}

				if len(result.Dexes) != tc.expectedDexes {
					t.Errorf("Expected %d dexes, got %d", tc.expectedDexes, len(result.Dexes))
				}
			}
		})
	}
}

// TestSearch_CanceledContext tests behavior with a canceled context
func TestSearch_CanceledContext(t *testing.T) {
	// Setup mock server that never responds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep to simulate long-running request
		time.Sleep(500 * time.Millisecond)
		fmt.Fprintln(w, `{"tokens":[],"pools":[],"dexes":[]}`)
	}))
	defer server.Close()

	// Create client with mock server URL
	client := NewClient()
	client.SetBaseURL(server.URL)

	// Create canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Call method with canceled context
	_, err := client.Search.Search(ctx, "eth")

	// Assert context canceled error
	if err == nil {
		t.Fatal("Expected error due to canceled context, got nil")
	}
}

func TestSearch_Success(t *testing.T) {
	// Setup mock server
	mockResponse := `{
		"tokens": [
			{
				"id": "eth-ethereum",
				"name": "Ethereum",
				"symbol": "ETH",
				"chain": "ethereum"
			},
			{
				"id": "eth-ether",
				"name": "Ether",
				"symbol": "ETH",
				"chain": "ethereum"
			}
		],
		"pools": [],
		"dexes": []
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Errorf("Expected path to be /search, got %s", r.URL.Path)
		}

		query := r.URL.Query().Get("query")
		if query != "eth" {
			t.Errorf("Expected query parameter 'query' to be 'eth', got %s", query)
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, mockResponse)
	}))
	defer server.Close()

	// Create client with mock server URL
	client := NewClient()
	client.SetBaseURL(server.URL)

	// Call method
	ctx := context.Background()
	results, err := client.Search.Search(ctx, "eth")

	// Assert results
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results.Tokens) != 2 {
		t.Fatalf("Expected 2 token results, got %d", len(results.Tokens))
	}

	// Check first result
	if results.Tokens[0].Name != "Ethereum" {
		t.Errorf("Expected first result name to be 'Ethereum', got %s", results.Tokens[0].Name)
	}

	if results.Tokens[0].Symbol != "ETH" {
		t.Errorf("Expected first result symbol to be 'ETH', got %s", results.Tokens[0].Symbol)
	}

	if results.Tokens[0].ID != "eth-ethereum" {
		t.Errorf("Expected first result id to be 'eth-ethereum', got %s", results.Tokens[0].ID)
	}

	if results.Tokens[0].Chain != "ethereum" {
		t.Errorf("Expected first result chain to be 'ethereum', got %s", results.Tokens[0].Chain)
	}

	// Check second result
	if results.Tokens[1].Name != "Ether" {
		t.Errorf("Expected second result name to be 'Ether', got %s", results.Tokens[1].Name)
	}
}

func TestSearch_ServerErrors(t *testing.T) {
	tests := []struct {
		name         string
		serverStatus int
		serverBody   string
		expectError  bool
	}{
		{
			name:         "Server error 500",
			serverStatus: 500,
			serverBody:   `{"error":"Internal server error"}`,
			expectError:  true,
		},
		{
			name:         "Invalid JSON",
			serverStatus: 200,
			serverBody:   `{not valid json`,
			expectError:  true,
		},
		{
			name:         "Empty response",
			serverStatus: 200,
			serverBody:   `{"tokens":[],"pools":[],"dexes":[]}`,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
				fmt.Fprintln(w, tt.serverBody)
			}))
			defer server.Close()

			// Create client with mock server URL
			client := NewClient()
			client.SetBaseURL(server.URL)

			// Call method
			ctx := context.Background()
			_, err := client.Search.Search(ctx, "eth")

			// Assert error
			if (err != nil) != tt.expectError {
				t.Errorf("Expected error: %v, got error: %v", tt.expectError, err)
			}
		})
	}
}

func TestSearch_Timeout(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping timeout test in short mode")
	}

	// Setup mock server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep to force timeout
		time.Sleep(2 * time.Second)
		fmt.Fprintln(w, `{"tokens":[],"pools":[],"dexes":[]}`)
	}))
	defer server.Close()

	// Create client with mock server URL
	client := NewClient()
	client.SetBaseURL(server.URL)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Call method with timeout context
	_, err := client.Search.Search(ctx, "eth")

	// Assert timeout error
	if err == nil {
		t.Fatal("Expected error due to timeout, got nil")
	}
}

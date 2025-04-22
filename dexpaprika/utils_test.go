package dexpaprika

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestUtils_GetStats(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get stats
	stats, err := client.Utils.GetStats(ctx)
	if err != nil {
		t.Fatalf("Utils.GetStats returned error: %v", err)
	}

	if stats == nil {
		t.Fatal("Utils.GetStats returned nil, expected stats")
	}

	// Check that stats have reasonable values
	if stats.Chains <= 0 {
		t.Error("Utils.GetStats returned invalid chains count (expected > 0)")
	}

	if stats.Pools <= 0 {
		t.Error("Utils.GetStats returned invalid pools count (expected > 0)")
	}

	if stats.Tokens <= 0 {
		t.Error("Utils.GetStats returned invalid tokens count (expected > 0)")
	}

	if stats.Factories <= 0 {
		t.Error("Utils.GetStats returned invalid factories count (expected > 0)")
	}

	t.Logf("Stats: %d chains, %d factories, %d pools, %d tokens",
		stats.Chains, stats.Factories, stats.Pools, stats.Tokens)
}

func TestUtils_GetStatsWithCanceledContext(t *testing.T) {
	// Create a client with test settings
	client := NewClient(
		WithRetryConfig(1, 1*time.Second, 2*time.Second),
	)

	// Create a context and cancel it immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel context before the request

	// Get stats with canceled context
	stats, err := client.Utils.GetStats(ctx)

	// Expect an error due to canceled context
	if err == nil {
		t.Fatal("Utils.GetStats with canceled context returned no error, expected context canceled error")
	}

	// Stats should be nil
	if stats != nil {
		t.Error("Utils.GetStats with canceled context returned non-nil stats, expected nil")
	}
}

// TestUtils_GetStatsWithMock tests the utils service with a mock server
func TestUtils_GetStatsWithMock(t *testing.T) {
	// Define test cases with different server responses
	tests := []struct {
		name           string
		serverResponse string
		statusCode     int
		expectError    bool
		expectedStats  *Stats
	}{
		{
			name:           "successful response",
			serverResponse: `{"chains": 10, "factories": 20, "pools": 1000, "tokens": 2000}`,
			statusCode:     http.StatusOK,
			expectError:    false,
			expectedStats:  &Stats{Chains: 10, Factories: 20, Pools: 1000, Tokens: 2000},
		},
		{
			name:           "empty response",
			serverResponse: `{}`,
			statusCode:     http.StatusOK,
			expectError:    false,
			expectedStats:  &Stats{},
		},
		{
			name:           "invalid JSON response",
			serverResponse: `{"chains": "invalid"}`,
			statusCode:     http.StatusOK,
			expectError:    true,
			expectedStats:  nil,
		},
		{
			name:           "server error",
			serverResponse: `{"error": "Internal server error"}`,
			statusCode:     http.StatusInternalServerError,
			expectError:    true,
			expectedStats:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check that the request is for the stats endpoint
				if r.URL.Path != "/stats" {
					t.Errorf("Expected request to '/stats', got '%s'", r.URL.Path)
				}

				// Set response headers
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.statusCode)
				fmt.Fprintln(w, tc.serverResponse)
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

			// Get stats
			stats, err := client.Utils.GetStats(ctx)

			// Check error
			if tc.expectError && err == nil {
				t.Error("Expected an error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check stats
			if tc.expectedStats != nil {
				if stats == nil {
					t.Fatal("Expected non-nil stats but got nil")
				}
				if stats.Chains != tc.expectedStats.Chains {
					t.Errorf("Expected Chains=%d, got %d", tc.expectedStats.Chains, stats.Chains)
				}
				if stats.Factories != tc.expectedStats.Factories {
					t.Errorf("Expected Factories=%d, got %d", tc.expectedStats.Factories, stats.Factories)
				}
				if stats.Pools != tc.expectedStats.Pools {
					t.Errorf("Expected Pools=%d, got %d", tc.expectedStats.Pools, stats.Pools)
				}
				if stats.Tokens != tc.expectedStats.Tokens {
					t.Errorf("Expected Tokens=%d, got %d", tc.expectedStats.Tokens, stats.Tokens)
				}
			}
		})
	}
}

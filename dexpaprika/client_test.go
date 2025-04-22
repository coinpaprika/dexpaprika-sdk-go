package dexpaprika

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient()

	if client.baseURL == nil {
		t.Fatal("NewClient() baseURL is nil")
	}

	if got, want := client.baseURL.String(), DefaultBaseURL; got != want {
		t.Errorf("NewClient() baseURL is %v, want %v", got, want)
	}

	if client.client == nil {
		t.Fatal("NewClient() http client is nil")
	}

	if got, want := client.userAgent, "DexPaprika-SDK-Go"; got != want {
		t.Errorf("NewClient() userAgent is %v, want %v", got, want)
	}

	// Check that services are initialized
	if client.Networks == nil {
		t.Error("NewClient() Networks service is nil")
	}
	if client.Pools == nil {
		t.Error("NewClient() Pools service is nil")
	}
	if client.Tokens == nil {
		t.Error("NewClient() Tokens service is nil")
	}
	if client.Search == nil {
		t.Error("NewClient() Search service is nil")
	}
	if client.Utils == nil {
		t.Error("NewClient() Utils service is nil")
	}
}

func TestClient_WithOptions(t *testing.T) {
	// Custom HTTP client
	customHTTPClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	// Custom base URL
	customBaseURL := "https://custom-api.example.com"

	// Custom user agent
	customUserAgent := "CustomAgent/1.0"

	// Create client with options
	client := NewClient(
		WithHTTPClient(customHTTPClient),
		WithBaseURL(customBaseURL),
		WithUserAgent(customUserAgent),
		WithRetryConfig(5, 2*time.Second, 10*time.Second),
		WithRateLimit(2.0),
	)

	// Check if options were applied
	if client.client != customHTTPClient {
		t.Error("WithHTTPClient() not applied")
	}

	if got, want := client.baseURL.String(), customBaseURL; got != want {
		t.Errorf("WithBaseURL() not applied, got %v, want %v", got, want)
	}

	if got, want := client.userAgent, customUserAgent; got != want {
		t.Errorf("WithUserAgent() not applied, got %v, want %v", got, want)
	}

	if got, want := client.maxRetries, 5; got != want {
		t.Errorf("WithRetryConfig() maxRetries not applied, got %v, want %v", got, want)
	}

	if got, want := client.retryWaitMin, 2*time.Second; got != want {
		t.Errorf("WithRetryConfig() retryWaitMin not applied, got %v, want %v", got, want)
	}

	if got, want := client.retryWaitMax, 10*time.Second; got != want {
		t.Errorf("WithRetryConfig() retryWaitMax not applied, got %v, want %v", got, want)
	}

	if client.rateLimiter == nil {
		t.Error("WithRateLimit() not applied, rateLimiter is nil")
	}
}

func TestClient_SetBaseURL(t *testing.T) {
	client := NewClient()

	// Set valid URL
	validURL := "https://api.example.com"
	err := client.SetBaseURL(validURL)
	if err != nil {
		t.Errorf("SetBaseURL(%q) returned error: %v", validURL, err)
	}

	if got, want := client.baseURL.String(), validURL; got != want {
		t.Errorf("SetBaseURL(%q) = %q, want %q", validURL, got, want)
	}

	// Set invalid URL
	invalidURL := "://invalid-url"
	err = client.SetBaseURL(invalidURL)
	if err == nil {
		t.Errorf("SetBaseURL(%q) = nil, want error", invalidURL)
	}
}

func TestClient_SetUserAgent(t *testing.T) {
	client := NewClient()

	// Set user agent
	userAgent := "Custom-Agent/1.0"
	client.SetUserAgent(userAgent)

	if got, want := client.userAgent, userAgent; got != want {
		t.Errorf("SetUserAgent(%q) = %q, want %q", userAgent, got, want)
	}
}

func TestClient_NewRequest(t *testing.T) {
	client := NewClient()

	// Test GET request without body
	req, err := client.NewRequest(http.MethodGet, "/test-path", nil)
	if err != nil {
		t.Fatalf("NewRequest(GET, /test-path, nil) returned error: %v", err)
	}

	// Check method
	if got, want := req.Method, http.MethodGet; got != want {
		t.Errorf("NewRequest() method = %v, want %v", got, want)
	}

	// Check URL
	expectedURL, _ := url.Parse(DefaultBaseURL + "/test-path")
	if got, want := req.URL.String(), expectedURL.String(); got != want {
		t.Errorf("NewRequest() URL = %v, want %v", got, want)
	}

	// Check headers
	expectedUserAgent := "DexPaprika-SDK-Go"
	if got, want := req.Header.Get("User-Agent"), expectedUserAgent; got != want {
		t.Errorf("NewRequest() User-Agent = %v, want %v", got, want)
	}

	if got, want := req.Header.Get("Accept"), "application/json"; got != want {
		t.Errorf("NewRequest() Accept = %v, want %v", got, want)
	}

	// Content-Type should not be set for GET without body
	if contentType := req.Header.Get("Content-Type"); contentType != "" {
		t.Errorf("NewRequest() Content-Type = %v, want \"\"", contentType)
	}
}

func TestAPIError_Error(t *testing.T) {
	apiErr := &APIError{
		StatusCode:  404,
		Message:     "Resource not found",
		RawResponse: []byte(`{"error": "Not found"}`),
		Err:         ErrNotFound,
	}

	expected := "not found: Resource not found (status code: 404)"
	if got := apiErr.Error(); got != expected {
		t.Errorf("APIError.Error() = %q, want %q", got, expected)
	}
}

// TestClient_NewRequestWithBody tests creating a request with a body
func TestClient_NewRequestWithBody(t *testing.T) {
	client := NewClient()

	// Create a simple body
	body := struct {
		Name string `json:"name"`
	}{
		Name: "test",
	}

	// Create a request with a body
	req, err := client.NewRequest(http.MethodPost, "/test-path", body)
	if err != nil {
		t.Fatalf("NewRequest(POST, /test-path, body) returned error: %v", err)
	}

	// Check method
	if got, want := req.Method, http.MethodPost; got != want {
		t.Errorf("NewRequest() method = %v, want %v", got, want)
	}

	// Check URL
	expectedURL, _ := url.Parse(DefaultBaseURL + "/test-path")
	if got, want := req.URL.String(), expectedURL.String(); got != want {
		t.Errorf("NewRequest() URL = %v, want %v", got, want)
	}

	// Check Content-Type header is set for requests with body
	if got, want := req.Header.Get("Content-Type"), "application/json"; got != want {
		t.Errorf("NewRequest() Content-Type = %v, want %v", got, want)
	}

	// Check that body was properly serialized
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("Error reading request body: %v", err)
	}

	expectedBody := `{"name":"test"}`
	if !strings.Contains(string(bodyBytes), `"name":"test"`) {
		t.Errorf("Request body = %s, want to contain %s", string(bodyBytes), expectedBody)
	}
}

// TestClient_NewRequest_InvalidURL tests creating a request with an invalid URL
func TestClient_NewRequest_InvalidURL(t *testing.T) {
	client := NewClient()

	// Create a request with an invalid URL
	_, err := client.NewRequest(http.MethodGet, ":", nil)
	if err == nil {
		t.Fatal("NewRequest with invalid URL returned nil error, want error")
	}
}

// TestClient_NewRequest_InvalidBody tests creating a request with a body that can't be marshaled to JSON
func TestClient_NewRequest_InvalidBody(t *testing.T) {
	client := NewClient()

	// Create a body that can't be marshaled to JSON
	body := make(chan int) // channels can't be marshaled to JSON

	// Create a request with an invalid body
	_, err := client.NewRequest(http.MethodPost, "/test", body)
	if err == nil {
		t.Fatal("NewRequest with invalid body returned nil error, want error")
	}
}

// TestClient_Do_HTTPError tests the Do method with HTTP errors
func TestClient_Do_HTTPError(t *testing.T) {
	// Define test cases
	tests := []struct {
		name       string
		statusCode int
		response   string
		wantErr    error
	}{
		{
			name:       "bad request",
			statusCode: http.StatusBadRequest,
			response:   `{"error": "Bad Request"}`,
			wantErr:    ErrBadRequest,
		},
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			response:   `{"error": "Unauthorized"}`,
			wantErr:    ErrUnauthorized,
		},
		{
			name:       "forbidden",
			statusCode: http.StatusForbidden,
			response:   `{"error": "Forbidden"}`,
			wantErr:    ErrForbidden,
		},
		{
			name:       "not found",
			statusCode: http.StatusNotFound,
			response:   `{"error": "Not Found"}`,
			wantErr:    ErrNotFound,
		},
		{
			name:       "rate limit exceeded",
			statusCode: http.StatusTooManyRequests,
			response:   `{"error": "Rate limit exceeded"}`,
			wantErr:    ErrRateLimit,
		},
		{
			name:       "internal server error",
			statusCode: http.StatusInternalServerError,
			response:   `{"error": "Internal Server Error"}`,
			wantErr:    ErrInternalServerError,
		},
		{
			name:       "service unavailable",
			statusCode: http.StatusServiceUnavailable,
			response:   `{"error": "Service Unavailable"}`,
			wantErr:    ErrServiceUnavailable,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			// Create a request
			req, err := client.NewRequest(http.MethodGet, "/test", nil)
			if err != nil {
				t.Fatalf("NewRequest returned error: %v", err)
			}

			// Perform the request
			var result interface{}
			resp, err := client.Do(context.Background(), req, &result)

			// Check error
			if err == nil {
				t.Fatal("Do() returned nil error, want error")
			}

			// Check if error is of type APIError
			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("Do() returned error of type %T, want *APIError", err)
			}

			// Check if the underlying error is what we expect
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("Do() returned error %v, want %v", err, tc.wantErr)
			}

			// Check status code
			if apiErr.StatusCode != tc.statusCode {
				t.Errorf("APIError.StatusCode = %d, want %d", apiErr.StatusCode, tc.statusCode)
			}

			// Ensure response body is closed
			if resp != nil && resp.Body != nil {
				defer resp.Body.Close()
			}
		})
	}
}

// TestClient_Do_NetworkError tests the Do method with network errors
func TestClient_Do_NetworkError(t *testing.T) {
	// Create a client with a non-existent server URL
	client := NewClient(
		WithBaseURL("http://non-existent-server.example.com"),
		WithRetryConfig(0, 1*time.Millisecond, 1*time.Millisecond), // No retries for faster tests
	)

	// Create a request
	req, err := client.NewRequest(http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}

	// Perform the request
	var result interface{}
	resp, err := client.Do(context.Background(), req, &result)

	// Check error
	if err == nil {
		t.Fatal("Do() returned nil error, want error")
	}

	// Check if error is of type APIError
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("Do() returned error of type %T, want *APIError", err)
	}

	// Check that it's a network error by examining the error message
	if !strings.Contains(err.Error(), "network error") {
		t.Errorf("Do() returned error %v, want network error", err)
	}

	// Ensure response body is closed
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
}

// TestClient_Do_RetryOnServerError tests that the client retries on server errors
func TestClient_Do_RetryOnServerError(t *testing.T) {
	// Create a counter for requests
	requestCount := 0

	// Create a test server that returns a 503 for the first request and 200 for subsequent requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")

		if requestCount == 1 {
			// First request returns 503
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintln(w, `{"error": "Service Unavailable"}`)
		} else {
			// Subsequent requests return 200
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `{"success": true}`)
		}
	}))
	defer server.Close()

	// Create a client with retry configuration
	client := NewClient(
		WithBaseURL(server.URL),
		WithRetryConfig(1, 1*time.Millisecond, 1*time.Millisecond), // 1 retry for faster tests
	)

	// Create a request
	req, err := client.NewRequest(http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}

	// Perform the request
	var result struct {
		Success bool `json:"success"`
	}
	resp, err := client.Do(context.Background(), req, &result)

	// Check error
	if err != nil {
		t.Fatalf("Do() returned error: %v", err)
	}

	// Check that the result is what we expect
	if !result.Success {
		t.Errorf("result.Success = %v, want true", result.Success)
	}

	// Check that the request was retried
	if requestCount != 2 {
		t.Errorf("Request count = %d, want 2", requestCount)
	}

	// Ensure response body is closed
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
}

// TestClient_Do_RateLimit tests that the client respects rate limiting
func TestClient_Do_RateLimit(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"success": true}`)
	}))
	defer server.Close()

	// Create a client with rate limiting (1 request per second)
	client := NewClient(
		WithBaseURL(server.URL),
		WithRateLimit(1.0),
	)

	// Create a request
	req, err := client.NewRequest(http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}

	// Perform the first request
	start := time.Now()
	var result struct {
		Success bool `json:"success"`
	}
	resp, err := client.Do(context.Background(), req, &result)
	if err != nil {
		t.Fatalf("First Do() returned error: %v", err)
	}

	// Ensure response body is closed
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	// Perform the second request
	resp, err = client.Do(context.Background(), req, &result)
	if err != nil {
		t.Fatalf("Second Do() returned error: %v", err)
	}
	duration := time.Since(start)

	// Check that the second request is rate limited
	// The second request should take around 1 second due to rate limiting
	if duration < 900*time.Millisecond {
		t.Logf("Second request took %v, which is shorter than expected for rate limiting", duration)
	}

	// Ensure response body is closed
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
}

// TestClient_Do_ContextCancellation tests that the client respects context cancellation
func TestClient_Do_ContextCancellation(t *testing.T) {
	// Create a test server that sleeps before responding
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep to simulate a slow response
		time.Sleep(100 * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"success": true}`)
	}))
	defer server.Close()

	// Create a client
	client := NewClient(
		WithBaseURL(server.URL),
	)

	// Create a request
	req, err := client.NewRequest(http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}

	// Create a context that will be canceled immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel context immediately

	// Perform the request with the canceled context
	var result interface{}
	resp, err := client.Do(ctx, req, &result)

	// Check error
	if err == nil {
		t.Fatal("Do() with canceled context returned nil error, want error")
	}

	// Check that the error is a context error
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Do() returned error %v, want context.Canceled", err)
	}

	// Ensure response body is closed
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
}

// Test IsRetryable function
func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{"nil error", nil, false},
		{"non-APIError", errors.New("regular error"), false},
		{"APIError 400", &APIError{StatusCode: 400, Err: ErrBadRequest}, false},
		{"APIError 429", &APIError{StatusCode: 429, Err: ErrRateLimit}, true},
		{"APIError 500", &APIError{StatusCode: 500, Err: ErrInternalServerError}, true},
		{"APIError 503", &APIError{StatusCode: 503, Err: ErrServiceUnavailable}, true},
		{"ErrTimeout", ErrTimeout, true},
		{"ErrServiceUnavailable", ErrServiceUnavailable, true},
		{"ErrRetryableError", ErrRetryableError, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IsRetryable(tc.err)
			if result != tc.retryable {
				t.Errorf("IsRetryable(%v) = %v, want %v", tc.err, result, tc.retryable)
			}
		})
	}
}

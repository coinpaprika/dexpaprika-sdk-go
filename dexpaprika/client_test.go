package dexpaprika

import (
	"net/http"
	"net/url"
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

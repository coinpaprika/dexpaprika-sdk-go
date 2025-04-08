// Package dexpaprika provides a Go client for the DexPaprika API.
package dexpaprika

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	// DefaultBaseURL is the default DexPaprika API endpoint
	DefaultBaseURL = "https://api.dexpaprika.com"
	// DefaultTimeout is the default timeout for API requests
	DefaultTimeout = 30 * time.Second
	// DefaultMaxRetries is the default number of retry attempts
	DefaultMaxRetries = 3
	// DefaultRetryWaitMin is the minimum amount of time to wait between retries
	DefaultRetryWaitMin = 1 * time.Second
	// DefaultRetryWaitMax is the maximum amount of time to wait between retries
	DefaultRetryWaitMax = 5 * time.Second
)

// Client represents a DexPaprika API client
type Client struct {
	// HTTP client used to communicate with the API
	client *http.Client

	// Base URL for API requests
	baseURL *url.URL

	// User agent for client
	userAgent string

	// Retry configuration
	maxRetries   int
	retryWaitMin time.Duration
	retryWaitMax time.Duration

	// Rate limiting
	rateLimiter *time.Ticker

	// Services used for communicating with the API
	Networks *NetworksService
	Pools    *PoolsService
	Tokens   *TokensService
	Search   *SearchService
	Utils    *UtilsService
}

// ClientOption is a function that configures a Client
type ClientOption func(*Client)

// WithHTTPClient sets the HTTP client for the API client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		if httpClient != nil {
			c.client = httpClient
		}
	}
}

// WithBaseURL sets the base URL for the API client
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		url, err := url.Parse(baseURL)
		if err == nil {
			c.baseURL = url
		}
	}
}

// WithUserAgent sets the user agent for the API client
func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

// WithRetryConfig sets the retry configuration for the API client
func WithRetryConfig(maxRetries int, retryWaitMin, retryWaitMax time.Duration) ClientOption {
	return func(c *Client) {
		c.maxRetries = maxRetries
		c.retryWaitMin = retryWaitMin
		c.retryWaitMax = retryWaitMax
	}
}

// WithRateLimit sets rate limiting for the API client (requests per second)
func WithRateLimit(requestsPerSecond float64) ClientOption {
	return func(c *Client) {
		if requestsPerSecond > 0 {
			interval := time.Duration(1e9 / requestsPerSecond)
			c.rateLimiter = time.NewTicker(interval)
		}
	}
}

// NewClient returns a new DexPaprika API client with the given options
func NewClient(options ...ClientOption) *Client {
	baseURL, _ := url.Parse(DefaultBaseURL)

	c := &Client{
		client: &http.Client{
			Timeout: DefaultTimeout,
		},
		baseURL:      baseURL,
		userAgent:    "DexPaprika-SDK-Go",
		maxRetries:   DefaultMaxRetries,
		retryWaitMin: DefaultRetryWaitMin,
		retryWaitMax: DefaultRetryWaitMax,
	}

	// Apply options
	for _, option := range options {
		option(c)
	}

	// Initialize services
	c.Networks = &NetworksService{client: c}
	c.Pools = &PoolsService{client: c}
	c.Tokens = &TokensService{client: c}
	c.Search = &SearchService{client: c}
	c.Utils = &UtilsService{client: c}

	return c
}

// SetBaseURL sets a custom base URL for the client
func (c *Client) SetBaseURL(urlStr string) error {
	baseURL, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	c.baseURL = baseURL
	return nil
}

// SetUserAgent sets a custom user agent string for the client
func (c *Client) SetUserAgent(userAgent string) {
	c.userAgent = userAgent
}

// NewRequest creates an API request
func (c *Client) NewRequest(method, path string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	u := c.baseURL.ResolveReference(rel)

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err = json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	return req, nil
}

// Error types
var (
	ErrBadRequest          = errors.New("bad request")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrNotFound            = errors.New("not found")
	ErrRateLimit           = errors.New("rate limit exceeded")
	ErrInternalServerError = errors.New("internal server error")
	ErrServiceUnavailable  = errors.New("service unavailable")
	ErrTimeout             = errors.New("request timeout")
	ErrRetryableError      = errors.New("retryable error")
)

// APIError represents a structured API error
type APIError struct {
	StatusCode  int
	Message     string
	RawResponse []byte
	Err         error
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s (status code: %d)", e.Err, e.Message, e.StatusCode)
	}
	return fmt.Sprintf("%s (status code: %d)", e.Err, e.StatusCode)
}

func (e *APIError) Unwrap() error {
	return e.Err
}

// IsRetryable returns whether the error is potentially retryable
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		// 5xx errors are potentially retryable
		if apiErr.StatusCode >= 500 && apiErr.StatusCode < 600 {
			return true
		}
		// 429 Too Many Requests is retryable
		if apiErr.StatusCode == 429 {
			return true
		}
	}

	// Check for network or timeout errors
	if errors.Is(err, ErrRetryableError) || errors.Is(err, ErrTimeout) || errors.Is(err, ErrServiceUnavailable) {
		return true
	}

	return false
}

// Do sends an API request and returns the API response
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	var resp *http.Response
	var err error
	var respBody []byte

	// Apply rate limiting if configured
	if c.rateLimiter != nil {
		select {
		case <-c.rateLimiter.C:
			// Rate limit wait completed
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Retry logic
	for i := 0; i <= c.maxRetries; i++ {
		if i > 0 {
			// Calculate backoff duration
			backoff := c.retryWaitMin * time.Duration(1<<uint(i-1))
			if backoff > c.retryWaitMax {
				backoff = c.retryWaitMax
			}

			// Wait with backoff
			timer := time.NewTimer(backoff)
			select {
			case <-timer.C:
				// Backoff completed
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			}
		}

		// Clone the request to ensure we can retry with a fresh request
		reqClone := req.Clone(ctx)
		resp, err = c.client.Do(reqClone)

		// Check for context cancellation
		select {
		case <-ctx.Done():
			if resp != nil {
				resp.Body.Close()
			}
			return nil, ctx.Err()
		default:
		}

		// If there was a network error, try again
		if err != nil {
			if i == c.maxRetries {
				return nil, &APIError{
					StatusCode: 0,
					Err:        fmt.Errorf("network error after %d retries: %w", c.maxRetries, err),
				}
			}
			continue
		}

		// Read the body
		respBody, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			if i == c.maxRetries {
				return nil, &APIError{
					StatusCode:  resp.StatusCode,
					Err:         fmt.Errorf("error reading response body after %d retries: %w", c.maxRetries, err),
					RawResponse: respBody,
				}
			}
			continue
		}

		// Check the response code
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			apiErr := createAPIError(resp, respBody)

			// If it's a retryable error and we haven't hit max retries, try again
			if IsRetryable(apiErr) && i < c.maxRetries {
				continue
			}

			return resp, apiErr
		}

		// Reconstruct the response body for reading
		resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

		// Decode the response if a target was specified
		if v != nil {
			if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
				return resp, &APIError{
					StatusCode:  resp.StatusCode,
					Err:         fmt.Errorf("error decoding response body: %w", err),
					RawResponse: respBody,
				}
			}
		}

		// Success, break out of retry loop
		break
	}

	return resp, nil
}

// createAPIError creates an appropriate APIError based on the HTTP status code
func createAPIError(resp *http.Response, body []byte) *APIError {
	var errMsg string
	var err error

	// Try to extract error message from body
	var errorResp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error != "" {
		errMsg = errorResp.Error
	}

	// Map status codes to appropriate errors
	switch resp.StatusCode {
	case 400:
		err = ErrBadRequest
	case 401:
		err = ErrUnauthorized
	case 403:
		err = ErrForbidden
	case 404:
		err = ErrNotFound
	case 429:
		err = ErrRateLimit
	case 500:
		err = ErrInternalServerError
	case 503:
		err = ErrServiceUnavailable
	default:
		if resp.StatusCode >= 500 {
			err = ErrRetryableError
		} else {
			err = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
	}

	return &APIError{
		StatusCode:  resp.StatusCode,
		Message:     errMsg,
		RawResponse: body,
		Err:         err,
	}
}

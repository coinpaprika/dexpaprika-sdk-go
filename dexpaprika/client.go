// Package dexpaprika provides a Go client for the DexPaprika API.
// This SDK is generated based on the DexPaprika OpenAPI 3.1.0 specification.
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
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the default DexPaprika API endpoint.
	//
	// It serves keyless callers and registered free keys alike. Only Pro moves to
	// api-pro.dexpaprika.com, selected with WithBaseURL. The host is never
	// inferred from the presence of a key: sending a free key to the Pro host
	// returns 403, so guessing would break the people who just registered.
	DefaultBaseURL = "https://api.dexpaprika.com"

	// Version of this SDK, reported in the User-Agent. Keep in step with the git
	// tag; the module proxy serves tags, not this constant.
	Version = "1.8.0"

	// APIKeyEnvVar is consulted when no key is passed to NewClient.
	//
	// This is the NAME of an environment variable, not a credential. gosec's
	// G101 heuristic fires on the identifier containing "APIKey", so it is
	// suppressed here rather than by renaming the constant into something less
	// clear at the call site.
	APIKeyEnvVar = "DEXPAPRIKA_API_KEY" //nolint:gosec // G101: env var name, not a secret
	// DefaultTimeout is the default timeout for API requests
	DefaultTimeout = 30 * time.Second
	// DefaultMaxRetries is the default number of retry attempts
	DefaultMaxRetries = 3
	// MaxServerRequestedWait caps how long we will honour a server-supplied
	// retry delay. The API has asked for 32 seconds on a free key, which is
	// worth waiting out, but an unbounded value from the wire should never be
	// able to park a caller's goroutine indefinitely.
	MaxServerRequestedWait = 60 * time.Second
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

	// Optional API key. Empty means keyless, which is the default and works.
	apiKey string

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

// WithAPIKey sets the API key sent with every request.
//
// Optional. Without one the client is keyless, which works and needs no signup.
// An explicit key here beats the DEXPAPRIKA_API_KEY environment variable.
//
// The key is sent as the entire Authorization value. There is no "Bearer" prefix
// and no other scheme word: the API checksums the raw header, so a scheme word
// returns 401. This is the most common reason a working key looks broken.
func WithAPIKey(apiKey string) ClientOption {
	return func(c *Client) {
		c.apiKey = sanitizeAPIKey(apiKey)
	}
}

// sanitizeAPIKey trims a key and rejects anything that could break out of a
// header. A key carrying CR, LF or NUL is dropped rather than mangled: a mangled
// key authenticates as nobody, and because the data endpoints ignore an
// unreadable key instead of rejecting it, the caller would never find out.
func sanitizeAPIKey(key string) string {
	key = strings.TrimSpace(key)
	if strings.ContainsAny(key, "\r\n\x00") {
		return ""
	}
	return key
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
		baseURL: baseURL,
		// Was the bare string "DexPaprika-SDK-Go", which said the SDK was in use
		// but never which version.
		userAgent: "DexPaprika-SDK-Go/" + Version,
		// Env fallback applied before options, so WithAPIKey overrides it.
		apiKey:       sanitizeAPIKey(os.Getenv(APIKeyEnvVar)),
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
	if c.apiKey != "" {
		// The whole value, with no scheme word in front of it.
		req.Header.Set("Authorization", c.apiKey)
	}

	return req, nil
}

// Error types
var (
	ErrBadRequest          = errors.New("bad request")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrNotFound            = errors.New("not found")
	ErrGone                = errors.New("endpoint has been removed")
	ErrRateLimit           = errors.New("rate limit exceeded")
	ErrInternalServerError = errors.New("internal server error")
	ErrServiceUnavailable  = errors.New("service unavailable")
	ErrTimeout             = errors.New("request timeout")
	ErrRetryableError      = errors.New("retryable error")
)

// APIError represents a structured API error
type APIError struct {
	StatusCode int
	Message    string
	// Replacement is the endpoint the API suggests using instead, parsed from
	// the "replacement" field of the error body when present. It is populated
	// for any error status (not only 410), so self-documenting deprecations are
	// discoverable by callers via err.Replacement. Empty when the API did not
	// advertise a replacement.
	Replacement string
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
	// serverDelay carries a Retry-After the server sent on the previous attempt.
	// Without it the client backs off on its own schedule, which loses every
	// race it was told how to win: the default ladder tops out at retryWaitMax,
	// and a 429 routinely asks for longer than that.
	var serverDelay time.Duration
	for i := 0; i <= c.maxRetries; i++ {
		if i > 0 {
			// Calculate backoff duration
			backoff := c.retryWaitMin * time.Duration(1<<uint(i-1))
			if backoff > c.retryWaitMax {
				backoff = c.retryWaitMax
			}
			// The server's own number wins when it is longer. Retrying earlier
			// than asked just spends another request on the same 429.
			if serverDelay > backoff {
				backoff = serverDelay
				if backoff > MaxServerRequestedWait {
					backoff = MaxServerRequestedWait
				}
			}
			serverDelay = 0

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
		resp, err = c.client.Do(reqClone) //nolint:gosec // URL is constructed from trusted baseURL

		// Check for context cancellation
		select {
		case <-ctx.Done():
			if resp != nil {
				_ = resp.Body.Close()
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
		_ = resp.Body.Close()
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

			// If it's a retryable error, and we haven't hit max retries, try again
			if IsRetryable(apiErr) && i < c.maxRetries {
				serverDelay = retryAfter(resp, respBody)
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
// retryAfter reports how long the server asked us to wait before trying again,
// or zero when it did not say.
//
// Two shapes, because DexPaprika uses both. A request carrying an API key gets
// a standard Retry-After header (observed: "retry-after: 32"). A keyless
// request gets a 429 with no such header and the delay only in the JSON body
// (observed: {"error":"rate_limited","tier":"keyless","retry_after":3}). Both
// were captured off api.dexpaprika.com on 2026-08-07.
//
// Retry-After also permits an HTTP-date, so that form is parsed too rather than
// silently treated as zero.
func retryAfter(resp *http.Response, body []byte) time.Duration {
	if resp != nil {
		if raw := resp.Header.Get("Retry-After"); raw != "" {
			if secs, err := strconv.Atoi(raw); err == nil && secs > 0 {
				return time.Duration(secs) * time.Second
			}
			if when, err := http.ParseTime(raw); err == nil {
				if d := time.Until(when); d > 0 {
					return d
				}
			}
		}
	}

	var payload struct {
		RetryAfter float64 `json:"retry_after"`
	}
	if err := json.Unmarshal(body, &payload); err == nil && payload.RetryAfter > 0 {
		return time.Duration(payload.RetryAfter * float64(time.Second))
	}
	return 0
}

func createAPIError(resp *http.Response, body []byte) *APIError {
	var errMsg string
	var replacement string
	var err error

	// Try to extract an error message from the body. This parse is defensive:
	// the body may not be JSON, or may omit any of these fields. In that case
	// the fields stay empty and we fall back to the status-based defaults below.
	var errorResp struct {
		Error       string `json:"error"`
		Message     string `json:"message"`
		Replacement string `json:"replacement"`
	}
	if jsonErr := json.Unmarshal(body, &errorResp); jsonErr == nil {
		switch {
		case errorResp.Error != "" && errorResp.Message != "":
			errMsg = errorResp.Error + ": " + errorResp.Message
		case errorResp.Error != "":
			errMsg = errorResp.Error
		case errorResp.Message != "":
			// Some responses (notably the 410 deprecation body) carry only
			// "message" with no "error" key. Use it so the API's own text is
			// not silently dropped.
			errMsg = errorResp.Message
		}
		replacement = errorResp.Replacement
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
	case 410:
		err = ErrGone
		// Provide a helpful migration message for deprecated endpoints only
		// when the API did not supply one of its own.
		if errMsg == "" {
			errMsg = "This endpoint has been deprecated. Please use network-specific endpoints instead.\n\nExamples:\n- client.Pools.ListByNetwork('ethereum', opts)\n- client.Pools.ListByNetwork('solana', opts)\n- client.Pools.ListByNetwork('fantom', opts)\n\nFor more information, visit: https://docs.dexpaprika.com/changelog/changelog"
		}
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

	// If the API advertised a replacement endpoint, surface it in the message
	// for ANY error status. This keys on the presence of the "replacement"
	// field rather than a specific status or endpoint, so future deprecations
	// self-document without further SDK changes.
	if replacement != "" {
		if errMsg != "" {
			errMsg = errMsg + " Use " + replacement + " instead."
		} else {
			errMsg = "Use " + replacement + " instead."
		}
	}

	return &APIError{
		StatusCode:  resp.StatusCode,
		Message:     errMsg,
		Replacement: replacement,
		RawResponse: body,
		Err:         err,
	}
}

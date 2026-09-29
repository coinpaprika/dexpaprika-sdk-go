package dexpaprika

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// tokenOHLCVRequest calls Tokens.GetOHLCV against a throwaway origin and
// returns the path and query string that actually reached the wire.
func tokenOHLCVRequest(t *testing.T, opts *TokenOHLCVOptions) (string, url.Values) {
	t.Helper()

	var gotPath string
	var gotQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	if _, err := client.Tokens.GetOHLCV(context.Background(), "ethereum", "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2", opts); err != nil {
		t.Fatalf("Tokens.GetOHLCV: %v", err)
	}
	return gotPath, gotQuery
}

func TestTokens_GetOHLCV_PathAndQuery(t *testing.T) {
	path, q := tokenOHLCVRequest(t, &TokenOHLCVOptions{
		Start:    "-24h",
		End:      "-1h",
		Interval: "1h",
		Limit:    24,
	})

	wantPath := "/networks/ethereum/tokens/0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2/ohlcv"
	if path != wantPath {
		t.Errorf("path = %q, want %q", path, wantPath)
	}
	if got := q.Get("start"); got != "-24h" {
		t.Errorf("start sent as %q, want -24h", got)
	}
	if got := q.Get("end"); got != "-1h" {
		t.Errorf("end sent as %q, want -1h", got)
	}
	if got := q.Get("interval"); got != "1h" {
		t.Errorf("interval sent as %q, want 1h", got)
	}
	if got := q.Get("limit"); got != "24" {
		t.Errorf("limit sent as %q, want 24", got)
	}
}

func TestTokens_GetOHLCV_NeverSendsInversed(t *testing.T) {
	// TokenOHLCVOptions has no Inversed field at all, but guard the wire
	// contract directly: this endpoint has no such parameter, and a future
	// edit that reused OHLCVOptions here must not start sending one.
	_, q := tokenOHLCVRequest(t, &TokenOHLCVOptions{Start: "-24h"})
	if _, present := q["inversed"]; present {
		t.Errorf("query carries \"inversed\" = %q, this endpoint has no such parameter", q.Get("inversed"))
	}
}

func TestTokens_GetOHLCV_LimitCappedAt1000(t *testing.T) {
	if got := tokenOHLCVRequestLimit(t, 1000); got != "1000" {
		t.Errorf("limit 1000 sent as %q", got)
	}
	if got := tokenOHLCVRequestLimit(t, 5000); got != "1000" {
		t.Errorf("limit 5000 sent as %q, want it capped at 1000", got)
	}
	if got := tokenOHLCVRequestLimit(t, 0); got != "" {
		t.Errorf("limit 0 sent as %q, want it omitted so the server default applies", got)
	}
}

func tokenOHLCVRequestLimit(t *testing.T, limit int) string {
	t.Helper()
	_, q := tokenOHLCVRequest(t, &TokenOHLCVOptions{Start: "-24h", Limit: limit})
	return q.Get("limit")
}

func TestTokens_GetOHLCV_Decoding(t *testing.T) {
	body := `[{"time_open":"2026-09-29T00:00:00Z","time_close":"2026-09-29T01:00:00Z","open":1.1,"high":1.3,"low":1.0,"close":1.2,"volume":123456}]`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	records, err := client.Tokens.GetOHLCV(context.Background(), "ethereum", "0xtoken", &TokenOHLCVOptions{Start: "-24h"})
	if err != nil {
		t.Fatalf("Tokens.GetOHLCV: %v", err)
	}

	var want []OHLCVRecord
	if err := json.Unmarshal([]byte(body), &want); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("got %d records, want 1", len(records))
	}
	if records[0] != want[0] {
		t.Errorf("record = %+v, want %+v", records[0], want[0])
	}
}

func TestTokens_GetOHLCV_ValidatesArguments(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	if _, err := client.Tokens.GetOHLCV(ctx, "", "0xtoken", nil); err == nil || err.Error() != "network ID is required" {
		t.Errorf("empty network ID: got %v, want \"network ID is required\"", err)
	}
	if _, err := client.Tokens.GetOHLCV(ctx, "ethereum", "", nil); err == nil || err.Error() != "token address is required" {
		t.Errorf("empty token address: got %v, want \"token address is required\"", err)
	}
}

// TestTokens_GetOHLCV_403SurfacesAPIMessage covers the access gate: a
// keyless or free-key call gets a 403 whose body names the plan that lifts
// it, and the SDK must hand that message to the caller unchanged rather than
// substituting a generic "forbidden" string.
func TestTokens_GetOHLCV_403SurfacesAPIMessage(t *testing.T) {
	const apiMessage = "this endpoint requires a Dev or Pro plan"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"` + apiMessage + `"}`))
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithRetryConfig(0, time.Millisecond, time.Millisecond),
	)

	_, err := client.Tokens.GetOHLCV(context.Background(), "ethereum", "0xtoken", &TokenOHLCVOptions{Start: "-24h"})
	if err == nil {
		t.Fatal("GetOHLCV returned nil error, want a 403")
	}
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("error = %v, want it to wrap ErrForbidden", err)
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusForbidden {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusForbidden)
	}
	if apiErr.Message != apiMessage {
		t.Errorf("Message = %q, want the API's own text %q", apiErr.Message, apiMessage)
	}
}

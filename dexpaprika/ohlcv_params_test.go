package dexpaprika

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// ohlcvQuery calls GetOHLCV against a throwaway origin and returns the query
// string that actually reached the wire.
func ohlcvQuery(t *testing.T, opts *OHLCVOptions) url.Values {
	t.Helper()

	var got url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	if _, err := client.Pools.GetOHLCV(context.Background(), "ethereum", "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640", opts); err != nil {
		t.Fatalf("GetOHLCV: %v", err)
	}
	return got
}

func TestGetOHLCV_LimitUpTo1000(t *testing.T) {
	// The API returns up to 1000 candles; the SDK used to cut every request to 366.
	if got := ohlcvQuery(t, &OHLCVOptions{Start: "-24h", Limit: 1000}).Get("limit"); got != "1000" {
		t.Errorf("limit 1000 sent as %q", got)
	}
	if got := ohlcvQuery(t, &OHLCVOptions{Start: "-24h", Limit: 5000}).Get("limit"); got != "1000" {
		t.Errorf("limit 5000 sent as %q, want it capped at 1000", got)
	}
}

func TestGetOHLCV_RelativeStartPassesThrough(t *testing.T) {
	q := ohlcvQuery(t, &OHLCVOptions{Start: "-24h", End: "-1h"})
	if q.Get("start") != "-24h" || q.Get("end") != "-1h" {
		t.Errorf("start/end sent as %q/%q, want -24h/-1h unchanged", q.Get("start"), q.Get("end"))
	}
}

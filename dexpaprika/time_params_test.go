package dexpaprika

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// queryFor runs call against a throwaway origin that answers body, and returns
// the query string that actually reached the wire.
func queryFor(t *testing.T, body string, call func(c *Client) error) url.Values {
	t.Helper()

	var got url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	if err := call(NewClient(WithBaseURL(server.URL))); err != nil {
		t.Fatalf("request failed: %v", err)
	}
	return got
}

const testPool = "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640"

func TestGetTransactions_RelativeFromTo(t *testing.T) {
	q := queryFor(t, `{"transactions":[],"page_info":{}}`, func(c *Client) error {
		_, err := c.Pools.GetTransactions(context.Background(), "ethereum", testPool, 0, 10, "", WithFrom("-1h"), WithTo("-5m"))
		return err
	})
	if got := q.Get("from"); got != "-1h" {
		t.Errorf("from sent as %q, want -1h", got)
	}
	if got := q.Get("to"); got != "-5m" {
		t.Errorf("to sent as %q, want -5m", got)
	}
}

func TestGetTransactions_WithFromReplacesTimestamp(t *testing.T) {
	q := queryFor(t, `{"transactions":[],"page_info":{}}`, func(c *Client) error {
		_, err := c.Pools.GetTransactions(context.Background(), "ethereum", testPool, 0, 10, "", WithFromTimestamp(1790000000), WithFrom("-24h"))
		return err
	})
	if got := q["from"]; len(got) != 1 || got[0] != "-24h" {
		t.Errorf("from sent as %v, want exactly [-24h]", got)
	}
}

func TestGetTransactions_UnixTimestampUnchanged(t *testing.T) {
	q := queryFor(t, `{"transactions":[],"page_info":{}}`, func(c *Client) error {
		_, err := c.Pools.GetTransactions(context.Background(), "ethereum", testPool, 0, 10, "", WithFromTimestamp(1790000000))
		return err
	})
	if got := q.Get("from"); got != "1790000000" {
		t.Errorf("from sent as %q, want 1790000000", got)
	}
}

func TestFilter_RelativeCreatedAfterPassesThrough(t *testing.T) {
	q := queryFor(t, `{"results":[],"has_next_page":false}`, func(c *Client) error {
		_, err := c.Pools.Filter(context.Background(), "ethereum", &PoolFilterOptions{CreatedAfter: "-24h", CreatedBefore: "-1h"})
		return err
	})
	if got := q.Get("created_after"); got != "-24h" {
		t.Errorf("pools created_after sent as %q, want -24h", got)
	}
	if got := q.Get("created_before"); got != "-1h" {
		t.Errorf("pools created_before sent as %q, want -1h", got)
	}

	q = queryFor(t, `{"results":[],"has_next_page":false}`, func(c *Client) error {
		_, err := c.Tokens.Filter(context.Background(), "ethereum", &TokenFilterOptions{CreatedAfter: "-7d"})
		return err
	})
	if got := q.Get("created_after"); got != "-7d" {
		t.Errorf("tokens created_after sent as %q, want -7d", got)
	}
}

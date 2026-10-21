package dexpaprika

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
)

// recordingClient points a client at a throwaway origin that records the path
// of every request reaching it.
func recordingClient(t *testing.T) (*Client, func() []string) {
	t.Helper()
	var mu sync.Mutex
	var seen []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		seen = append(seen, r.URL.EscapedPath())
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(server.Close)
	client := NewClient(WithBaseURL(server.URL), WithRetryConfig(0, 0, 0))
	return client, func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), seen...)
	}
}

const (
	pathTestPool  = "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640"
	pathTestToken = "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"
)

type pathCall func(c *Client, network, address string) error

// Every method that builds a path from caller input. networkCalls put only the
// network id in the path; addressCalls put a pool or token address there too.
var networkCalls = map[string]pathCall{
	"Pools.ListByNetwork": func(c *Client, n, _ string) error {
		_, err := c.Pools.ListByNetwork(context.Background(), n, nil)
		return err
	},
	"Pools.ListByDex": func(c *Client, n, _ string) error {
		_, err := c.Pools.ListByDex(context.Background(), n, "uniswap_v3", nil)
		return err
	},
	"Pools.Filter": func(c *Client, n, _ string) error { _, err := c.Pools.Filter(context.Background(), n, nil); return err },
	"Tokens.GetPools": func(c *Client, n, _ string) error {
		_, err := c.Tokens.GetPools(context.Background(), n, pathTestToken, nil)
		return err
	},
	"Tokens.GetTop": func(c *Client, n, _ string) error {
		_, err := c.Tokens.GetTop(context.Background(), n, nil)
		return err
	},
	"Tokens.Filter": func(c *Client, n, _ string) error {
		_, err := c.Tokens.Filter(context.Background(), n, nil)
		return err
	},
	"Tokens.GetMultiPrices": func(c *Client, n, _ string) error {
		_, err := c.Tokens.GetMultiPrices(context.Background(), n, []string{pathTestToken})
		return err
	},
	"Networks.ListDexes": func(c *Client, n, _ string) error {
		_, err := c.Networks.ListDexes(context.Background(), n, 0, 0)
		return err
	},
}

var addressCalls = map[string]pathCall{
	"Pools.GetDetails": func(c *Client, n, a string) error {
		_, err := c.Pools.GetDetails(context.Background(), n, a, false)
		return err
	},
	"Pools.GetOHLCV": func(c *Client, n, a string) error {
		_, err := c.Pools.GetOHLCV(context.Background(), n, a, &OHLCVOptions{Start: "-24h", Interval: "1h"})
		return err
	},
	"Pools.GetTransactions": func(c *Client, n, a string) error {
		_, err := c.Pools.GetTransactions(context.Background(), n, a, 0, 0, "")
		return err
	},
	"Pools.GetIndicator": func(c *Client, n, a string) error {
		_, err := c.Pools.GetIndicator(context.Background(), n, a, &IndicatorOptions{Indicator: "rsi"})
		return err
	},
	"Tokens.GetDetails": func(c *Client, n, a string) error {
		_, err := c.Tokens.GetDetails(context.Background(), n, a)
		return err
	},
	"Tokens.GetOHLCV": func(c *Client, n, a string) error {
		_, err := c.Tokens.GetOHLCV(context.Background(), n, a, &TokenOHLCVOptions{Start: "-24h", Interval: "1h"})
		return err
	},
	"Tokens.GetIndicator": func(c *Client, n, a string) error {
		_, err := c.Tokens.GetIndicator(context.Background(), n, a, &IndicatorOptions{Indicator: "rsi"})
		return err
	},
}

var badSegments = []string{
	".", "..", "../../usage", "../../../stats", "eth/../x", `a\b`, "x?limit=1000", "x#y",
	"%2e%2e", "..%2Fusage", "a b", "a\nb", "a\x00b",
}

func TestNetworkIDThatLeavesItsSegmentIsRefusedBeforeSending(t *testing.T) {
	all := map[string]pathCall{}
	for k, v := range networkCalls {
		all[k] = v
	}
	for k, v := range addressCalls {
		all[k] = v
	}
	for name, call := range all {
		for _, bad := range badSegments {
			client, seen := recordingClient(t)
			if err := call(client, bad, pathTestPool); err == nil {
				t.Errorf("%s(network=%q): no error", name, bad)
			}
			if got := seen(); len(got) != 0 {
				t.Errorf("%s(network=%q): sent %v", name, bad, got)
			}
		}
	}
}

func TestAddressThatLeavesItsSegmentIsRefusedBeforeSending(t *testing.T) {
	for name, call := range addressCalls {
		for _, bad := range badSegments {
			client, seen := recordingClient(t)
			if err := call(client, "ethereum", bad); err == nil {
				t.Errorf("%s(address=%q): no error", name, bad)
			}
			if got := seen(); len(got) != 0 {
				t.Errorf("%s(address=%q): sent %v", name, bad, got)
			}
		}
	}
}

func TestRealIDsReachThePathUnchanged(t *testing.T) {
	cases := [][2]string{
		{"ethereum", pathTestPool},
		{"bob_network", pathTestToken},
		{"solana", "So11111111111111111111111111111111111111112"},
		{"sui", "0x0000000000000000000000000000000000000000000000000000000000000002::sui::SUI"},
		{"ton", "EQA-X_yo3fzzbDbJ_0bzFWKqtRuZFIRa1sJsveZJ1YpViO3r"},
	}
	for _, tc := range cases {
		network, address := tc[0], tc[1]
		client, seen := recordingClient(t)
		_, _ = client.Tokens.GetDetails(context.Background(), network, address)
		_, _ = client.Pools.GetDetails(context.Background(), network, address, false)
		want := []string{
			"/networks/" + network + "/tokens/" + address,
			"/networks/" + network + "/pools/" + address,
		}
		if got := seen(); !reflect.DeepEqual(got, want) {
			t.Errorf("%s %s: sent %v, want %v", network, address, got, want)
		}
	}
}

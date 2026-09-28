// Package livethrottle spaces the requests a test suite sends to the live
// DexPaprika API so that the suite stays under the per-minute limit.
//
// The limit is counted per IP (see
// https://docs.dexpaprika.com/knowledge-base/rate-limits). Before this package
// existed the live tests sent their requests back to back, the burst ran into
// 429, and whichever test happened to be running when the retries or its
// context deadline ran out failed. It was a different test on every CI run.
//
// Only test code imports this package.
package livethrottle

import (
	"context"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	// Keyless is the spacing without a key: 15 requests a minute, plus a margin.
	Keyless = 4200 * time.Millisecond
	// Keyed is the spacing with DEXPAPRIKA_API_KEY set: a free key allows 30
	// requests a minute, plus a margin.
	Keyed = 2100 * time.Millisecond
)

// Transport starts requests to Host at least Interval apart. Requests to any
// other host, such as an httptest server, pass straight through.
//
// The wait happens inside RoundTrip and therefore counts against the caller's
// context and http.Client timeout. Tests that call the live API need a
// deadline that allows for it.
type Transport struct {
	Base     http.RoundTripper
	Host     string
	Interval time.Duration

	mu       sync.Mutex
	next     time.Time
	requests int
}

// RoundTrip implements http.RoundTripper.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Hostname() == t.Host {
		if err := t.wait(req.Context()); err != nil {
			return nil, err
		}
	}
	return t.Base.RoundTrip(req)
}

// wait reserves the next free slot, then sleeps until it comes up. A caller
// whose context ends first gives its slot up unused, which only makes the
// spacing more conservative.
func (t *Transport) wait(ctx context.Context) error {
	t.mu.Lock()
	start := time.Now()
	if t.next.After(start) {
		start = t.next
	}
	t.next = start.Add(t.Interval)
	t.requests++
	t.mu.Unlock()

	d := time.Until(start)
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Requests reports how many requests to Host have gone through.
func (t *Transport) Requests() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.requests
}

// Install wraps http.DefaultTransport, which every client built without a
// Transport of its own ends up using, and returns the wrapper. Call it from
// TestMain before m.Run. The spacing follows DEXPAPRIKA_API_KEY: set, the
// keyed pace; unset or empty, the keyless one.
//
// The limit is per IP, so two test binaries throttled separately still add
// up. Run the packages one at a time (go test -p 1).
func Install(host string) *Transport {
	interval := Keyless
	if strings.TrimSpace(os.Getenv("DEXPAPRIKA_API_KEY")) != "" {
		interval = Keyed
	}
	t := &Transport{Base: http.DefaultTransport, Host: host, Interval: interval}
	http.DefaultTransport = t
	return t
}

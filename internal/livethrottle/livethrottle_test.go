package livethrottle

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"
)

// startTimes records when each request reached the origin behind the throttle.
func startTimes(t *testing.T, throttledHost string, interval time.Duration, n int) []time.Time {
	t.Helper()

	var (
		mu  sync.Mutex
		got []time.Time
	)
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		got = append(got, time.Now())
		mu.Unlock()
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if throttledHost == "" {
		throttledHost = u.Hostname()
	}
	client := &http.Client{Transport: &Transport{Base: http.DefaultTransport, Host: throttledHost, Interval: interval}}
	for i := 0; i < n; i++ {
		resp, err := client.Get(server.URL)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
	}
	mu.Lock()
	defer mu.Unlock()
	return got
}

func TestTransport_SpacesRequestsToHost(t *testing.T) {
	const interval = 60 * time.Millisecond
	got := startTimes(t, "", interval, 4)
	for i := 1; i < len(got); i++ {
		if gap := got[i].Sub(got[i-1]); gap < interval-5*time.Millisecond {
			t.Errorf("request %d started %v after the previous one, want at least %v", i, gap, interval)
		}
	}
}

func TestTransport_LeavesOtherHostsAlone(t *testing.T) {
	start := time.Now()
	startTimes(t, "api.example.invalid", time.Second, 4)
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Errorf("4 requests to an unthrottled host took %v, want no added delay", elapsed)
	}
}

func TestTransport_StopsWaitingWhenContextEnds(t *testing.T) {
	tr := &Transport{Base: http.DefaultTransport, Host: "api.example.invalid", Interval: time.Hour}
	tr.next = time.Now().Add(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.example.invalid/networks", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := tr.RoundTrip(req)
	if resp != nil {
		_ = resp.Body.Close()
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("RoundTrip error is %v, want context.DeadlineExceeded", err)
	}
}

func TestInstall_PaceFollowsAPIKey(t *testing.T) {
	orig := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = orig })

	t.Setenv("DEXPAPRIKA_API_KEY", "")
	if got := Install("api.dexpaprika.com").Interval; got != Keyless {
		t.Errorf("without a key the interval is %v, want %v", got, Keyless)
	}
	http.DefaultTransport = orig

	t.Setenv("DEXPAPRIKA_API_KEY", "api_test")
	if got := Install("api.dexpaprika.com").Interval; got != Keyed {
		t.Errorf("with a key the interval is %v, want %v", got, Keyed)
	}
}

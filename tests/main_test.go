package tests

import (
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/coinpaprika/dexpaprika-sdk-go/dexpaprika"
	"github.com/coinpaprika/dexpaprika-sdk-go/internal/livethrottle"
)

// TestMain spaces the suite's requests to the live API. See
// internal/livethrottle for why.
func TestMain(m *testing.M) {
	u, err := url.Parse(dexpaprika.DefaultBaseURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	throttle := livethrottle.Install(u.Hostname())
	code := m.Run()
	fmt.Fprintf(os.Stderr, "live API: %d requests to %s, %v apart\n", throttle.Requests(), u.Hostname(), throttle.Interval)
	os.Exit(code)
}

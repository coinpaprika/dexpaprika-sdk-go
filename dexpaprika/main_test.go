package dexpaprika

import (
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/coinpaprika/dexpaprika-sdk-go/internal/livethrottle"
)

// liveTimeout bounds a test that calls the live API. TestMain spaces those
// calls a few seconds apart and a 429 can ask for more, so the 5 to 30 second
// budgets the tests used to carry ran out in CI, a different test each run.
const liveTimeout = 3 * time.Minute

func TestMain(m *testing.M) {
	u, err := url.Parse(DefaultBaseURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	throttle := livethrottle.Install(u.Hostname())
	code := m.Run()
	fmt.Fprintf(os.Stderr, "live API: %d requests to %s, %v apart\n", throttle.Requests(), u.Hostname(), throttle.Interval)
	os.Exit(code)
}

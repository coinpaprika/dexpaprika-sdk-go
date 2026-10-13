package dexpaprika

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// Indicator values accepted by IndicatorOptions.Indicator.
const (
	IndicatorRSI    = "rsi"
	IndicatorEMA    = "ema"
	IndicatorMACD   = "macd"
	IndicatorBBands = "bbands"
	IndicatorVWAP   = "vwap"
)

// IndicatorOptions contains options for a technical indicator request.
//
// Indicator is required; Pools.GetIndicator and Tokens.GetIndicator return an
// error before sending a request if it is empty, the same check the API makes
// server-side with 400 missing_param. Every other field is sent only when
// set: the API refuses a parameter that belongs to a different indicator (for
// example fast_period with Indicator: "rsi" returns 400 invalid_param), so do
// not fill in a field just because it has a default on the wire.
//
// Leave Interval unset to get the server default rather than hard-coding one:
// it depends on the caller's plan and on whether the call is for a pool or a
// token. See https://docs.dexpaprika.com/knowledge-base/rate-limits and
// https://dexpaprika.com/api/pricing for current limits.
type IndicatorOptions struct {
	// Indicator selects the calculation: "rsi", "ema", "macd", "bbands" or
	// "vwap". Required.
	Indicator string
	// Interval: 1m, 5m, 10m, 15m, 30m, 1h, 6h, 12h, 24h or 1d (1d and 24h are
	// the same series on the wire). Leave empty for the server default.
	Interval string
	// Limit is the number of data points returned, 1 to 1000. The server
	// default is 100.
	Limit int
	// Start is a relative offset from now ("-24h", "-7d"), Unix seconds, RFC
	// 3339 or YYYY-MM-DD. Do not combine with Cursor.
	Start string
	// End is the same formats as Start. Exclusive, defaults to now.
	End string
	// Cursor pages through results; pass a previous response's NextCursor.
	// Do not combine with Start.
	Cursor string
	// Period is the lookback window. Accepted by rsi (default 14), ema
	// (default 20), bbands (default 20) and vwap (default 14). macd does not
	// accept it; sending it there returns 400 invalid_param.
	Period int
	// FastPeriod is the macd fast length, 1 to 200, default 12. macd only.
	FastPeriod int
	// SlowPeriod is the macd slow length, 2 to 200, default 26, and must be
	// greater than FastPeriod. macd only.
	SlowPeriod int
	// SignalPeriod is the macd signal length, 1 to 200, default 9. macd only.
	SignalPeriod int
	// StdDev is the bbands band width in standard deviations, greater than 0
	// and up to 10, default 2. bbands only.
	StdDev float64
}

// validateIndicatorOptions requires Indicator to be set before any request is
// sent. A missing indicator always fails server-side with 400 missing_param,
// so failing here saves the round trip.
func validateIndicatorOptions(opts *IndicatorOptions) error {
	if opts == nil || opts.Indicator == "" {
		return fmt.Errorf("indicator is required")
	}
	return nil
}

// addIndicatorQuery adds opts to q, sending only the fields the caller set.
func addIndicatorQuery(q url.Values, opts *IndicatorOptions) {
	q.Add("indicator", opts.Indicator)
	if opts.Interval != "" {
		q.Add("interval", opts.Interval)
	}
	if opts.Limit > 0 {
		limit := opts.Limit
		if limit > 1000 {
			limit = 1000
		}
		q.Add("limit", fmt.Sprintf("%d", limit))
	}
	if opts.Start != "" {
		q.Add("start", opts.Start)
	}
	if opts.End != "" {
		q.Add("end", opts.End)
	}
	if opts.Cursor != "" {
		q.Add("cursor", opts.Cursor)
	}
	if opts.Period > 0 {
		q.Add("period", fmt.Sprintf("%d", opts.Period))
	}
	if opts.FastPeriod > 0 {
		q.Add("fast_period", fmt.Sprintf("%d", opts.FastPeriod))
	}
	if opts.SlowPeriod > 0 {
		q.Add("slow_period", fmt.Sprintf("%d", opts.SlowPeriod))
	}
	if opts.SignalPeriod > 0 {
		q.Add("signal_period", fmt.Sprintf("%d", opts.SignalPeriod))
	}
	if opts.StdDev > 0 {
		q.Add("std_dev", strconv.FormatFloat(opts.StdDev, 'f', -1, 64))
	}
}

// IndicatorDataPoint is one point of a technical indicator series.
//
// Which fields are populated depends on the indicator requested: rsi, ema and
// vwap set Value; macd sets MACD, Signal and Histogram; bbands sets Upper,
// Middle and Lower. The fields an indicator does not report stay nil.
type IndicatorDataPoint struct {
	// Timestamp is the candle open time, RFC 3339.
	Timestamp string   `json:"timestamp"`
	Value     *float64 `json:"value,omitempty"`
	MACD      *float64 `json:"macd,omitempty"`
	Signal    *float64 `json:"signal,omitempty"`
	Histogram *float64 `json:"histogram,omitempty"`
	Upper     *float64 `json:"upper,omitempty"`
	Middle    *float64 `json:"middle,omitempty"`
	Lower     *float64 `json:"lower,omitempty"`
}

// IndicatorResponse is the response from a pool or token indicator request.
//
// PoolAddress is set on a response from Pools.GetIndicator, TokenAddress on a
// response from Tokens.GetIndicator; the other is always empty. Parameters
// holds the parameters the server actually used to compute Data, including
// any default the caller left unset. Data is oldest first, one point per
// completed candle.
type IndicatorResponse struct {
	Network      string               `json:"network"`
	PoolAddress  string               `json:"pool_address,omitempty"`
	TokenAddress string               `json:"token_address,omitempty"`
	Indicator    string               `json:"indicator"`
	Interval     string               `json:"interval"`
	Parameters   map[string]float64   `json:"parameters"`
	Data         []IndicatorDataPoint `json:"data"`
	HasNextPage  bool                 `json:"has_next_page"`
	NextCursor   string               `json:"next_cursor"`
}

// GetIndicator returns a technical indicator series for a pool.
//
// Implements the getPoolIndicator operation. opts.Indicator is required and
// selects rsi, ema, macd, bbands or vwap; only the fields the caller set on
// opts are sent, because the API refuses a parameter that does not belong to
// the chosen indicator.
//
// All five indicators are available on every plan; interval and history
// depth follow the same per-plan limits as Pools.GetOHLCV. See
// https://docs.dexpaprika.com/knowledge-base/rate-limits and
// https://dexpaprika.com/api/pricing for current limits, and
// https://docs.dexpaprika.com/api-reference/pools/get-a-technical-indicator-for-a-pool
// for the reference.
func (s *PoolsService) GetIndicator(ctx context.Context, networkID, poolAddress string, opts *IndicatorOptions) (*IndicatorResponse, error) {
	if err := validateNetworkID(networkID); err != nil {
		return nil, err
	}
	if err := validatePoolAddress(poolAddress); err != nil {
		return nil, err
	}
	if err := validateIndicatorOptions(opts); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/networks/%s/pools/%s/indicators", networkID, poolAddress)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	addIndicatorQuery(q, opts)
	req.URL.RawQuery = q.Encode()

	var response IndicatorResponse
	r, err := s.client.Do(ctx, req, &response)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return &response, nil
}

// GetIndicator returns a technical indicator series for a token, computed
// across every pool the token trades in on that network.
//
// Implements the getTokenIndicator operation. opts.Indicator is required, the
// same as Pools.GetIndicator, and only the fields the caller set on opts are
// sent.
//
// This endpoint requires a Dev, Pro or Enterprise plan. A keyless or
// free-key call gets a 403 whose *APIError matches errors.Is(err,
// ErrForbidden); its Message is the API's own text and its RequiredTier
// names the plan that lifts it ("dev"). See
// https://docs.dexpaprika.com/api-reference/tokens/get-a-technical-indicator-for-a-token
// and https://dexpaprika.com/api/pricing for current plan availability.
func (s *TokensService) GetIndicator(ctx context.Context, networkID, tokenAddress string, opts *IndicatorOptions) (*IndicatorResponse, error) {
	if err := validateNetworkID(networkID); err != nil {
		return nil, err
	}
	if tokenAddress == "" {
		return nil, fmt.Errorf("token address is required")
	}
	if err := validateIndicatorOptions(opts); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/networks/%s/tokens/%s/indicators", networkID, tokenAddress)

	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	addIndicatorQuery(q, opts)
	req.URL.RawQuery = q.Encode()

	var response IndicatorResponse
	r, err := s.client.Do(ctx, req, &response)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return &response, nil
}

package dexpaprika

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// poolIndicatorRequest calls Pools.GetIndicator against a throwaway origin
// and returns the path and query string that actually reached the wire.
func poolIndicatorRequest(t *testing.T, opts *IndicatorOptions) (string, url.Values) {
	t.Helper()

	var gotPath string
	var gotQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"network":"ethereum","pool_address":"0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640","indicator":"rsi","interval":"1h","parameters":{"period":14},"data":[],"has_next_page":false,"next_cursor":""}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	if _, err := client.Pools.GetIndicator(context.Background(), "ethereum", "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640", opts); err != nil {
		t.Fatalf("Pools.GetIndicator: %v", err)
	}
	return gotPath, gotQuery
}

// tokenIndicatorRequest is the token equivalent of poolIndicatorRequest.
func tokenIndicatorRequest(t *testing.T, opts *IndicatorOptions) (string, url.Values) {
	t.Helper()

	var gotPath string
	var gotQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"network":"ethereum","token_address":"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2","indicator":"rsi","interval":"1h","parameters":{"period":14},"data":[],"has_next_page":false,"next_cursor":""}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	if _, err := client.Tokens.GetIndicator(context.Background(), "ethereum", "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2", opts); err != nil {
		t.Fatalf("Tokens.GetIndicator: %v", err)
	}
	return gotPath, gotQuery
}

func TestPools_GetIndicator_PathAndQuery(t *testing.T) {
	path, q := poolIndicatorRequest(t, &IndicatorOptions{
		Indicator: IndicatorRSI,
		Interval:  "1h",
		Start:     "-6h",
		Limit:     50,
		Period:    14,
	})

	wantPath := "/networks/ethereum/pools/0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640/indicators"
	if path != wantPath {
		t.Errorf("path = %q, want %q", path, wantPath)
	}
	if got := q.Get("indicator"); got != "rsi" {
		t.Errorf("indicator sent as %q, want rsi", got)
	}
	if got := q.Get("interval"); got != "1h" {
		t.Errorf("interval sent as %q, want 1h", got)
	}
	if got := q.Get("start"); got != "-6h" {
		t.Errorf("start sent as %q, want -6h", got)
	}
	if got := q.Get("limit"); got != "50" {
		t.Errorf("limit sent as %q, want 50", got)
	}
	if got := q.Get("period"); got != "14" {
		t.Errorf("period sent as %q, want 14", got)
	}
}

func TestTokens_GetIndicator_PathAndQuery(t *testing.T) {
	path, q := tokenIndicatorRequest(t, &IndicatorOptions{
		Indicator: IndicatorEMA,
		Interval:  "1h",
	})

	wantPath := "/networks/ethereum/tokens/0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2/indicators"
	if path != wantPath {
		t.Errorf("path = %q, want %q", path, wantPath)
	}
	if got := q.Get("indicator"); got != "ema" {
		t.Errorf("indicator sent as %q, want ema", got)
	}
}

// TestGetIndicator_OnlySetFieldsAreSent guards the rule the API enforces
// server-side: a parameter that belongs to another indicator is refused
// (indicator=rsi&fast_period=3 -> 400 invalid_param), so the SDK must never
// send a zero-value field the caller did not set.
func TestGetIndicator_OnlySetFieldsAreSent(t *testing.T) {
	_, q := poolIndicatorRequest(t, &IndicatorOptions{Indicator: IndicatorRSI})

	for _, param := range []string{
		"interval", "limit", "start", "end", "cursor", "period",
		"fast_period", "slow_period", "signal_period", "std_dev",
	} {
		if _, present := q[param]; present {
			t.Errorf("query carries %q = %q with nothing set on IndicatorOptions, want it omitted", param, q.Get(param))
		}
	}
	if len(q) != 1 || q.Get("indicator") != "rsi" {
		t.Errorf("query = %v, want only indicator=rsi", q)
	}
}

func TestGetIndicator_MACDParams(t *testing.T) {
	_, q := poolIndicatorRequest(t, &IndicatorOptions{
		Indicator:    IndicatorMACD,
		FastPeriod:   12,
		SlowPeriod:   26,
		SignalPeriod: 9,
	})
	if got := q.Get("fast_period"); got != "12" {
		t.Errorf("fast_period sent as %q, want 12", got)
	}
	if got := q.Get("slow_period"); got != "26" {
		t.Errorf("slow_period sent as %q, want 26", got)
	}
	if got := q.Get("signal_period"); got != "9" {
		t.Errorf("signal_period sent as %q, want 9", got)
	}
	if _, present := q["period"]; present {
		t.Errorf("query carries %q, macd does not accept period", q.Get("period"))
	}
}

func TestGetIndicator_BBandsStdDev(t *testing.T) {
	_, q := poolIndicatorRequest(t, &IndicatorOptions{
		Indicator: IndicatorBBands,
		Period:    20,
		StdDev:    2.5,
	})
	if got := q.Get("std_dev"); got != "2.5" {
		t.Errorf("std_dev sent as %q, want 2.5", got)
	}
	if got := q.Get("period"); got != "20" {
		t.Errorf("period sent as %q, want 20", got)
	}
}

func TestGetIndicator_Cursor(t *testing.T) {
	_, q := poolIndicatorRequest(t, &IndicatorOptions{
		Indicator: IndicatorRSI,
		Cursor:    "1791882000",
	})
	if got := q.Get("cursor"); got != "1791882000" {
		t.Errorf("cursor sent as %q, want 1791882000", got)
	}
}

func TestGetIndicator_LimitCappedAt1000(t *testing.T) {
	if _, q := poolIndicatorRequest(t, &IndicatorOptions{Indicator: IndicatorRSI, Limit: 1000}); q.Get("limit") != "1000" {
		t.Errorf("limit 1000 sent as %q", q.Get("limit"))
	}
	if _, q := poolIndicatorRequest(t, &IndicatorOptions{Indicator: IndicatorRSI, Limit: 5000}); q.Get("limit") != "1000" {
		t.Errorf("limit 5000 sent as %q, want it capped at 1000", q.Get("limit"))
	}
}

// TestGetIndicator_RequiresIndicator covers client-side validation: a missing
// Indicator always fails server-side with 400 missing_param, so the SDK
// fails before sending a request.
func TestGetIndicator_RequiresIndicator(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	if _, err := client.Pools.GetIndicator(ctx, "ethereum", "0xpool", nil); err == nil || err.Error() != "indicator is required" {
		t.Errorf("Pools.GetIndicator nil opts: got %v, want \"indicator is required\"", err)
	}
	if _, err := client.Pools.GetIndicator(ctx, "ethereum", "0xpool", &IndicatorOptions{}); err == nil || err.Error() != "indicator is required" {
		t.Errorf("Pools.GetIndicator empty Indicator: got %v, want \"indicator is required\"", err)
	}
	if _, err := client.Tokens.GetIndicator(ctx, "ethereum", "0xtoken", &IndicatorOptions{}); err == nil || err.Error() != "indicator is required" {
		t.Errorf("Tokens.GetIndicator empty Indicator: got %v, want \"indicator is required\"", err)
	}
	if _, err := client.Pools.GetIndicator(ctx, "", "0xpool", &IndicatorOptions{Indicator: IndicatorRSI}); err == nil || err.Error() != "network ID is required" {
		t.Errorf("Pools.GetIndicator empty network: got %v, want \"network ID is required\"", err)
	}
	if _, err := client.Pools.GetIndicator(ctx, "ethereum", "", &IndicatorOptions{Indicator: IndicatorRSI}); err == nil || err.Error() != "pool address is required" {
		t.Errorf("Pools.GetIndicator empty pool address: got %v, want \"pool address is required\"", err)
	}
	if _, err := client.Tokens.GetIndicator(ctx, "ethereum", "", &IndicatorOptions{Indicator: IndicatorRSI}); err == nil || err.Error() != "token address is required" {
		t.Errorf("Tokens.GetIndicator empty token address: got %v, want \"token address is required\"", err)
	}
}

// TestPools_GetIndicator_DecodesRSI covers the rsi/ema/vwap shape: a flat
// Value per point.
func TestPools_GetIndicator_DecodesRSI(t *testing.T) {
	body := `{"network":"ethereum","pool_address":"0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640","indicator":"rsi","interval":"1h","parameters":{"period":14},"data":[{"timestamp":"2026-10-13T06:00:00Z","value":54.38666301724629},{"timestamp":"2026-10-13T07:00:00Z","value":53.184343874602874}],"has_next_page":true,"next_cursor":"1791882000"}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	resp, err := client.Pools.GetIndicator(context.Background(), "ethereum", "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640", &IndicatorOptions{Indicator: IndicatorRSI})
	if err != nil {
		t.Fatalf("Pools.GetIndicator: %v", err)
	}

	if resp.Network != "ethereum" {
		t.Errorf("Network = %q, want ethereum", resp.Network)
	}
	if resp.PoolAddress != "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640" {
		t.Errorf("PoolAddress = %q, want the pool address", resp.PoolAddress)
	}
	if resp.TokenAddress != "" {
		t.Errorf("TokenAddress = %q, want empty on a pool response", resp.TokenAddress)
	}
	if resp.Indicator != "rsi" {
		t.Errorf("Indicator = %q, want rsi", resp.Indicator)
	}
	if got := resp.Parameters["period"]; got != 14 {
		t.Errorf("Parameters[period] = %v, want 14", got)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("got %d data points, want 2", len(resp.Data))
	}
	if resp.Data[0].Value == nil || *resp.Data[0].Value != 54.38666301724629 {
		t.Errorf("Data[0].Value = %v, want 54.38666301724629", resp.Data[0].Value)
	}
	if resp.Data[0].Timestamp != "2026-10-13T06:00:00Z" {
		t.Errorf("Data[0].Timestamp = %q, want 2026-10-13T06:00:00Z", resp.Data[0].Timestamp)
	}
	if resp.Data[0].MACD != nil || resp.Data[0].Upper != nil {
		t.Errorf("Data[0] carries a macd or bbands field for an rsi response: %+v", resp.Data[0])
	}
	if !resp.HasNextPage {
		t.Error("HasNextPage = false, want true")
	}
	if resp.NextCursor != "1791882000" {
		t.Errorf("NextCursor = %q, want 1791882000", resp.NextCursor)
	}
}

// TestPools_GetIndicator_DecodesMACD covers the macd shape: MACD, Signal and
// Histogram per point, with Value left nil.
func TestPools_GetIndicator_DecodesMACD(t *testing.T) {
	body := `{"network":"ethereum","pool_address":"0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640","indicator":"macd","interval":"1h","parameters":{"fast_period":12,"signal_period":9,"slow_period":26},"data":[{"histogram":11.220674833348312,"macd":-11.534946379135818,"signal":-22.75562121248413,"timestamp":"2026-10-13T06:00:00Z"}],"has_next_page":false,"next_cursor":""}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	resp, err := client.Pools.GetIndicator(context.Background(), "ethereum", "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640", &IndicatorOptions{
		Indicator:    IndicatorMACD,
		FastPeriod:   12,
		SlowPeriod:   26,
		SignalPeriod: 9,
	})
	if err != nil {
		t.Fatalf("Pools.GetIndicator: %v", err)
	}

	if len(resp.Data) != 1 {
		t.Fatalf("got %d data points, want 1", len(resp.Data))
	}
	point := resp.Data[0]
	if point.MACD == nil || *point.MACD != -11.534946379135818 {
		t.Errorf("MACD = %v, want -11.534946379135818", point.MACD)
	}
	if point.Signal == nil || *point.Signal != -22.75562121248413 {
		t.Errorf("Signal = %v, want -22.75562121248413", point.Signal)
	}
	if point.Histogram == nil || *point.Histogram != 11.220674833348312 {
		t.Errorf("Histogram = %v, want 11.220674833348312", point.Histogram)
	}
	if point.Value != nil {
		t.Errorf("Value = %v, want nil on a macd response", point.Value)
	}
	if resp.Parameters["fast_period"] != 12 || resp.Parameters["slow_period"] != 26 || resp.Parameters["signal_period"] != 9 {
		t.Errorf("Parameters = %v, want fast_period 12, slow_period 26, signal_period 9", resp.Parameters)
	}
}

// TestTokens_GetIndicator_Forbidden403SurfacesPlan covers the token
// indicator access gate: a keyless or free-key call gets a 403 plan_required
// whose required_tier names the plan that lifts it, and the SDK must hand
// that through on the error cleanly rather than losing it.
func TestTokens_GetIndicator_Forbidden403SurfacesPlan(t *testing.T) {
	const apiMessage = "this endpoint requires a Dev plan or higher"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"plan_required","message":"` + apiMessage + `","required_tier":"dev"}`))
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithRetryConfig(0, time.Millisecond, time.Millisecond),
	)

	_, err := client.Tokens.GetIndicator(context.Background(), "ethereum", "0xtoken", &IndicatorOptions{Indicator: IndicatorRSI})
	if err == nil {
		t.Fatal("Tokens.GetIndicator returned nil error, want a 403")
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
	if apiErr.Message != "plan_required: "+apiMessage {
		t.Errorf("Message = %q, want the API's own text", apiErr.Message)
	}
	if apiErr.RequiredTier != "dev" {
		t.Errorf("RequiredTier = %q, want dev", apiErr.RequiredTier)
	}
}

// TestGetIndicator_400SurfacesParam covers the 400 shapes the API sends for
// this endpoint: missing_param on no indicator, and invalid_param on a
// parameter that does not belong to the chosen indicator. Both name the
// offending field in "param", which the SDK must expose on APIError.Param.
func TestGetIndicator_400SurfacesParam(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_param","message":"invalid query parameters: fast_period (unsupported parameter for rsi)","param":"fast_period"}`))
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithRetryConfig(0, time.Millisecond, time.Millisecond),
	)

	_, err := client.Pools.GetIndicator(context.Background(), "ethereum", "0xpool", &IndicatorOptions{Indicator: IndicatorRSI, FastPeriod: 3})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is %T, want *APIError", err)
	}
	if !errors.Is(err, ErrBadRequest) {
		t.Errorf("error = %v, want it to wrap ErrBadRequest", err)
	}
	if apiErr.Param != "fast_period" {
		t.Errorf("Param = %q, want fast_period", apiErr.Param)
	}
}

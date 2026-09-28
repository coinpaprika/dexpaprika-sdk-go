package dexpaprika

import (
	"context"
	"encoding/json"
	"testing"
)

// The body is what GET /networks/ethereum/multi/prices returned on 2026-09-28.
const multiPricesBody = `[{"chain":"ethereum","id":"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2","price_usd":2683.9242610156934,"last_updated":"2026-09-28T13:06:30Z"},{"chain":"ethereum","id":"0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48","price_usd":0.9999998572249761,"last_updated":"2026-09-28T13:06:30Z"}]`

// TestGetMultiPrices_KeepsLastUpdated decodes the live response shape and
// encodes it again. A field missing from TokenPrice is dropped on the way in,
// so it is absent on the way out, which is how last_updated was lost before
// 1.10.1.
func TestGetMultiPrices_KeepsLastUpdated(t *testing.T) {
	var prices []TokenPrice
	queryFor(t, multiPricesBody, func(c *Client) error {
		var err error
		prices, err = c.Tokens.GetMultiPrices(context.Background(), "ethereum", []string{
			"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
			"0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
		})
		return err
	})
	if len(prices) != 2 {
		t.Fatalf("got %d prices, want 2", len(prices))
	}

	out, err := json.Marshal(prices)
	if err != nil {
		t.Fatalf("re-encoding prices: %v", err)
	}
	var rows []map[string]any
	if err := json.Unmarshal(out, &rows); err != nil {
		t.Fatalf("decoding re-encoded prices: %v", err)
	}
	for i, row := range rows {
		if got := row["last_updated"]; got != "2026-09-28T13:06:30Z" {
			t.Errorf("price %d: last_updated is %v after decoding, want 2026-09-28T13:06:30Z", i, got)
		}
	}
}

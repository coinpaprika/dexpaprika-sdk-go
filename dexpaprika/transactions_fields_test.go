package dexpaprika

import (
	"context"
	"encoding/json"
	"sort"
	"testing"
)

// The body is what GET /networks/ethereum/pools/0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640/transactions?limit=1
// returned on 2026-09-28. Solana rows carry the same keys.
const transactionsBody = `{"transactions":[{"id":"0xba5e6abbf0db7b9e3287314c027c9d641e778ecba8947379ff57702e94b4337e","log_index":18,"transaction_index":3,"factory_id":"0x1f98431c8ad98523631ae4a59f267346ea31f984","pool_id":"0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640","chain":"ethereum","sender":"0x6487986a78b126538746937fbd25516971129b3a","recipient":"0x6487986a78b126538746937fbd25516971129b3a","token_0":"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2","token_0_symbol":"WETH","token_1":"0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48","token_1_symbol":"USDC","amount_0":1860551367961434,"amount_1":-5000000,"volume_0":0.001860551367961434,"volume_1":5,"price_0":2686.0318100209583,"price_1":0.00037229640999381807,"price_0_usd":2687.4460054447513,"price_1_usd":0.9993393685258273,"created_at_block_number":26076286,"created_at_block_hash":"0x777edec312fa456179e7ca9669cb2a17cd79e55e09027643142d2694e0762584","created_at":"2026-09-28T13:34:47Z","canonical_chain":true}],"page_info":{"limit":1,"page":1,"total_items":48864,"total_pages":48864}}`

func keysOf(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// TestGetTransactions_KeepsEveryField decodes the live row, encodes it again
// and compares the keys. A field missing from Transaction is dropped on the
// way in and absent on the way out, which is how created_at, the symbols,
// volumes and prices were lost before 1.10.1.
func TestGetTransactions_KeepsEveryField(t *testing.T) {
	var resp *TransactionsResponse
	queryFor(t, transactionsBody, func(c *Client) error {
		var err error
		resp, err = c.Pools.GetTransactions(context.Background(), "ethereum", "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640", 0, 1, "")
		return err
	})
	if len(resp.Transactions) != 1 {
		t.Fatalf("got %d transactions, want 1", len(resp.Transactions))
	}

	var wire struct {
		Transactions []map[string]any `json:"transactions"`
	}
	if err := json.Unmarshal([]byte(transactionsBody), &wire); err != nil {
		t.Fatalf("decoding fixture: %v", err)
	}
	out, err := json.Marshal(resp.Transactions[0])
	if err != nil {
		t.Fatalf("re-encoding transaction: %v", err)
	}
	var row map[string]any
	if err := json.Unmarshal(out, &row); err != nil {
		t.Fatalf("decoding re-encoded transaction: %v", err)
	}

	want, got := keysOf(wire.Transactions[0]), keysOf(row)
	if len(got) != len(want) {
		t.Fatalf("after decoding the row has %d keys, the API sent %d\n got: %v\nwant: %v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("key mismatch after decoding\n got: %v\nwant: %v", got, want)
		}
	}

	// Values, read back from the re-encoded row so this also compiles against
	// the old struct and fails there on the key comparison above.
	for key, want := range map[string]any{
		"created_at":      "2026-09-28T13:34:47Z",
		"token_0_symbol":  "WETH",
		"token_1_symbol":  "USDC",
		"price_0_usd":     2687.4460054447513,
		"volume_1":        5.0,
		"canonical_chain": true,
	} {
		if row[key] != want {
			t.Errorf("%s = %v after decoding, want %v", key, row[key], want)
		}
	}
}

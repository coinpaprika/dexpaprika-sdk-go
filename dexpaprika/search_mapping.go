package dexpaprika

// This file holds the small, pure helpers that translate the SDK's historical
// sort-field names into the canonical field names accepted by the unified
// /pools/search and /tokens/search endpoints.
//
// The search endpoints reject legacy values with HTTP 400, so every sort field
// the caller supplies has to be normalized before it goes on the wire. An empty
// or unrecognized value falls back to the API default, "volume_usd_24h".

// defaultSearchSortField is returned for empty or unknown sort fields.
const defaultSearchSortField = "volume_usd_24h"

// mapPoolSortField maps a pools sort field to the canonical value accepted by
// /networks/{network}/pools/search. Canonical values pass through unchanged.
func mapPoolSortField(field string) string {
	switch field {
	// Legacy -> canonical
	case "volume_usd", "volume_24h":
		return "volume_usd_24h"
	case "transactions":
		return "txns_24h"
	case "last_price_change_usd_24h":
		return "price_change_percentage_24h"
	case "volume_7d":
		return "volume_usd_7d"
	case "volume_30d":
		return "volume_usd_30d"
	case "liquidity":
		return "liquidity_usd"
	// Canonical pass-through
	case "volume_usd_24h",
		"volume_usd_7d",
		"volume_usd_30d",
		"liquidity_usd",
		"txns_24h",
		"created_at",
		"price_usd",
		"price_change_percentage_24h",
		// Shorter windows, pools only. tokens/search does not carry them and
		// mapTokenSortField deliberately does not list them.
		"price_change_percentage_6h",
		"price_change_percentage_1h",
		"price_change_percentage_5m":
		return field
	default:
		return defaultSearchSortField
	}
}

// mapTokenSortField maps a tokens sort field to the canonical value accepted by
// /networks/{network}/tokens/search. Canonical values pass through unchanged.
// Note: tokens/search rejects ordering by price, so "price_usd" is remapped to
// the default rather than passed through.
func mapTokenSortField(field string) string {
	switch field {
	// Legacy -> canonical
	case "volume_24h":
		return "volume_usd_24h"
	case "txns":
		return "txns_24h"
	case "price_change":
		return "price_change_percentage_24h"
	case "fdv":
		return "fdv_usd"
	case "volume_7d":
		return "volume_usd_7d"
	case "volume_30d":
		return "volume_usd_30d"
	case "price_usd":
		// tokens/search returns 400 when ordering by price.
		return "volume_usd_24h"
	// Canonical pass-through
	case "volume_usd_24h",
		"volume_usd_7d",
		"volume_usd_30d",
		"liquidity_usd",
		"txns_24h",
		"fdv_usd",
		"created_at",
		"price_change_percentage_24h":
		return field
	default:
		return defaultSearchSortField
	}
}

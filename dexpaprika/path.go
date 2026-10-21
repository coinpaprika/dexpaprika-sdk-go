package dexpaprika

import (
	"fmt"
	"strings"
)

// unsafeSegmentChars are separators and anything a URL parser or the server
// could decode into one: slash, backslash, query and fragment marks and
// percent escapes. Whitespace and control characters are refused separately.
// Ids the API returns contain none of these; Sui and Aptos ids with "::" and
// TON addresses with "-" and "_" are unaffected.
const unsafeSegmentChars = `/\?#%`

// checkSegment returns an error unless value is safe to place in a request
// path as one segment. A value such as "../../usage" or "x?limit=1000" would
// otherwise send the request, and the client's API key, to a different path
// or query than the method was written for.
func checkSegment(what, value string) error {
	bad := value == "." || value == ".." || strings.ContainsAny(value, unsafeSegmentChars)
	for _, r := range value {
		if r <= 0x20 || r == 0x7f {
			bad = true
		}
	}
	if bad {
		if len(value) > 100 {
			value = value[:100]
		}
		return fmt.Errorf("%s must be a single id or address, got %q", what, value)
	}
	return nil
}

// validateTokenAddress validates that a token address is not empty and is
// safe to place in a request path.
func validateTokenAddress(tokenAddress string) error {
	if tokenAddress == "" {
		return fmt.Errorf("token address is required")
	}
	return checkSegment("token address", tokenAddress)
}

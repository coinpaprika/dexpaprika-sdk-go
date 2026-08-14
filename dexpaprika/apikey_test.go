package dexpaprika

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// headersFor spins up a throwaway origin, points a client at it and returns the
// headers that actually reached the wire. Asserting on the request rather than
// on the struct is the only level that proves the rules are applied.
func headersFor(t *testing.T, options ...ClientOption) http.Header {
	t.Helper()

	var got http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	options = append([]ClientOption{WithBaseURL(server.URL)}, options...)
	client := NewClient(options...)

	req, err := client.NewRequest(http.MethodGet, "/networks", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	resp, err := client.client.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer resp.Body.Close()

	return got
}

// ── The Bearer rule ────────────────────────────────────────────────────────
// Authorization: Bearer api_... returns 401 because the API checksums the raw
// header value. The mistake has resurfaced three times in four months, so pin it
// against every scheme word somebody might reach for.

func TestKeyIsTheEntireAuthorizationValue(t *testing.T) {
	t.Setenv(APIKeyEnvVar, "")
	if got := headersFor(t, WithAPIKey("api_abc123")).Get("Authorization"); got != "api_abc123" {
		t.Fatalf("Authorization = %q, want %q", got, "api_abc123")
	}
}

func TestNoSchemeWordIsEverPrepended(t *testing.T) {
	t.Setenv(APIKeyEnvVar, "")
	value := headersFor(t, WithAPIKey("api_abc123")).Get("Authorization")
	for _, scheme := range []string{"bearer", "token", "apikey", "basic", "key"} {
		if strings.HasPrefix(strings.ToLower(value), scheme) {
			t.Fatalf("Authorization %q starts with scheme word %q", value, scheme)
		}
	}
}

// ── Keyless stays the default ──────────────────────────────────────────────

func TestNoKeySendsNoAuthorizationHeader(t *testing.T) {
	t.Setenv(APIKeyEnvVar, "")
	if _, present := headersFor(t)["Authorization"]; present {
		t.Fatal("keyless client sent an Authorization header")
	}
}

func TestBlankKeyIsKeylessNotAnEmptyHeader(t *testing.T) {
	t.Setenv(APIKeyEnvVar, "")
	for _, value := range []string{"", "   ", "\t"} {
		if _, present := headersFor(t, WithAPIKey(value))["Authorization"]; present {
			t.Fatalf("blank key %q produced an Authorization header", value)
		}
	}
}

// ── Precedence ─────────────────────────────────────────────────────────────

func TestEnvironmentVariableIsUsedWhenNoOptionIsGiven(t *testing.T) {
	t.Setenv(APIKeyEnvVar, "api_from_env")
	if got := headersFor(t).Get("Authorization"); got != "api_from_env" {
		t.Fatalf("Authorization = %q, want the value from the environment", got)
	}
}

func TestExplicitOptionBeatsTheEnvironment(t *testing.T) {
	t.Setenv(APIKeyEnvVar, "api_from_env")
	if got := headersFor(t, WithAPIKey("api_explicit")).Get("Authorization"); got != "api_explicit" {
		t.Fatalf("Authorization = %q, want the explicit option to win", got)
	}
}

func TestSurroundingWhitespaceIsTrimmed(t *testing.T) {
	t.Setenv(APIKeyEnvVar, "  api_padded\n")
	if got := headersFor(t).Get("Authorization"); got != "api_padded" {
		t.Fatalf("Authorization = %q, want the trimmed key", got)
	}
}

// ── Header injection ───────────────────────────────────────────────────────

func TestKeyWithControlCharactersIsDropped(t *testing.T) {
	for _, value := range []string{"api_a\r\nX-Evil: 1", "api_a\nb", "api_a\x00b"} {
		if got := sanitizeAPIKey(value); got != "" {
			t.Fatalf("sanitizeAPIKey(%q) = %q, want it dropped", value, got)
		}
	}
}

// ── Identification ─────────────────────────────────────────────────────────

func TestUserAgentCarriesTheVersion(t *testing.T) {
	t.Setenv(APIKeyEnvVar, "")
	got := headersFor(t).Get("User-Agent")
	if want := "DexPaprika-SDK-Go/" + Version; got != want {
		t.Fatalf("User-Agent = %q, want %q", got, want)
	}
	// Was the bare name, which said the SDK was in use but never which version.
	if got == "DexPaprika-SDK-Go" {
		t.Fatal("User-Agent regressed to the unversioned string")
	}
}

func TestCallerSuppliedUserAgentStillWins(t *testing.T) {
	t.Setenv(APIKeyEnvVar, "")
	if got := headersFor(t, WithUserAgent("my-app/1.0")).Get("User-Agent"); got != "my-app/1.0" {
		t.Fatalf("User-Agent = %q, want the caller's", got)
	}
}

// ── Host rules ─────────────────────────────────────────────────────────────

func TestAKeyAloneNeverChangesTheHost(t *testing.T) {
	t.Setenv(APIKeyEnvVar, "")
	client := NewClient(WithAPIKey("api_abc123"))
	if got := client.baseURL.String(); got != DefaultBaseURL {
		t.Fatalf("baseURL = %q, want %q. Free keys are served from the default host; only Pro moves.", got, DefaultBaseURL)
	}
}

func TestProCustomersSetTheHostExplicitly(t *testing.T) {
	t.Setenv(APIKeyEnvVar, "")
	client := NewClient(WithAPIKey("api_pro"), WithBaseURL("https://api-pro.dexpaprika.com"))
	if got := client.baseURL.String(); got != "https://api-pro.dexpaprika.com" {
		t.Fatalf("baseURL = %q, want the Pro host", got)
	}
}

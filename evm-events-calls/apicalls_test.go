package evm_events_calls

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildFullAPIURL(t *testing.T) {
	// Get mainnet chain config
	mainnetChain := ChainConfigByID["mainnet"]
	require.NotNil(t, mainnetChain, "mainnet chain config should exist")

	// Test address
	address := "0x1f98431c8ad98523631ae4a59f267346ea31f984"

	// Test parameters as specified in the request
	params := url.Values{
		"module":  {"contract"},
		"action":  {"getsourcecode"},
		"address": {address},
	}

	// Test with API key
	apiKey := "test-api-key"
	result := buildFullAPIURL(mainnetChain, params, apiKey)
	t.Logf("URL with API key: %s", result)

	// URL parameters are sorted alphabetically by url.Values.Encode()
	expectedWithKey := "https://api.etherscan.io/v2/api?action=getsourcecode&address=0x1f98431c8ad98523631ae4a59f267346ea31f984&apiKey=test-api-key&chainid=1&module=contract"
	require.Equal(t, expectedWithKey, result)

	// Test without API key
	resultNoKey := buildFullAPIURL(mainnetChain, params, "")
	t.Logf("URL without API key: %s", resultNoKey)

	expectedWithoutKey := "https://api.etherscan.io/v2/api?action=getsourcecode&address=0x1f98431c8ad98523631ae4a59f267346ea31f984&chainid=1&module=contract"
	require.Equal(t, expectedWithoutKey, resultNoKey)
}

func TestBuildFullAPIURLWithLegacyChain(t *testing.T) {
	// Test with a chain that uses ApiEndpoint instead of ApiBaseURL
	legacyChain := &ChainConfig{
		ID:             "test-legacy",
		ApiEndpoint:    "https://api.test.com",
		ApiBaseURL:     "", // Empty to trigger legacy behavior
		ApiQueryParams: url.Values{"chainid": {"999"}},
	}

	params := url.Values{
		"module":  {"contract"},
		"action":  {"getabi"},
		"address": {"0x123"},
	}

	apiKey := "legacy-key"
	result := buildFullAPIURL(legacyChain, params, apiKey)
	t.Logf("Legacy chain URL: %s", result)

	// URL parameters are sorted alphabetically by url.Values.Encode()
	expected := "https://api.test.com/api?action=getabi&address=0x123&apiKey=legacy-key&chainid=999&module=contract"
	require.Equal(t, expected, result)
}

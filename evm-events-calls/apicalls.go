package evm_events_calls

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/streamingfast/dhttp"
	"github.com/streamingfast/eth-go"
	"github.com/tidwall/gjson"
)

var httpClient = http.Client{
	Transport: dhttp.NewLoggingRoundTripper(zlog, tracer, http.DefaultTransport),
	Timeout:   30 * time.Second,
}

// buildAPIURL constructs the API URL based on the chain configuration
func buildAPIURL(chain *ChainConfig) string {
	if chain.ApiBaseURL != "" {
		// Use Etherscan V2 structure - ApiBaseURL already includes /api path
		return chain.ApiBaseURL
	}
	// Fall back to Etherscan V1 ApiEndpoint
	return chain.ApiEndpoint
}

// buildFullAPIURL constructs the complete API URL with module, action, and parameters
func buildFullAPIURL(chain *ChainConfig, params url.Values, apiKey string) string {
	baseURL := buildAPIURL(chain)

	allParams := url.Values{}

	for key, values := range chain.ApiQueryParams {
		allParams[key] = values
	}

	for key, values := range params {
		allParams[key] = values
	}

	if apiKey != "" {
		allParams.Set("apiKey", apiKey)
	}

	if chain.ApiBaseURL != "" {
		// Etherscan V2: baseURL already includes /api path
		baseURL += "?" + allParams.Encode()
	} else {
		// Etherscan V1: need to add /api path
		baseURL += "/api?" + allParams.Encode()
	}

	return baseURL
}

// buildDirectAPIURL constructs the direct API URL for contract ABI fetching
func buildDirectAPIURL(chain *ChainConfig, address string) string {
	if chain.ApiBaseURL != "" {
		url := chain.ApiBaseURL + "/" + address
		if len(chain.ApiQueryParams) > 0 {
			url += "?" + chain.ApiQueryParams.Encode()
		}
		return url
	}
	return fmt.Sprintf("%s/%s", chain.ApiEndpoint, address)
}

func getContractABIFollowingProxy(ctx context.Context, contractAddress string, chain *ChainConfig) (*ABI, error) {

	if cachedABI := chain.abiCache[contractAddress]; cachedABI != nil {
		return cachedABI, nil
	}

	if chain.ApiEndpointDirect {
		abi, abiContent, err := getContractABIDirect(ctx, contractAddress, chain)
		if err != nil {
			return nil, err
		}
		return &ABI{abi, abiContent}, nil
	}
	abi, abiContent, wait, err := getContractABI(ctx, contractAddress, chain, os.Getenv(chain.APIKeyEnvVar))
	if err != nil {
		return nil, err
	}

	<-wait.C
	implementationAddress, wait, err := getProxyContractImplementation(ctx, contractAddress, chain, os.Getenv(chain.APIKeyEnvVar))
	if err != nil {
		return nil, err
	}
	<-wait.C

	if implementationAddress != "" {
		implementationABI, implementationABIContent, wait, err := getContractABI(ctx, implementationAddress, chain, os.Getenv(chain.APIKeyEnvVar))
		if err != nil {
			return nil, err
		}

		for k, v := range implementationABI.LogEventsMap {
			abi.LogEventsMap[k] = append(abi.LogEventsMap[k], v...)
		}

		for k, v := range implementationABI.LogEventsByNameMap {
			abi.LogEventsByNameMap[k] = append(abi.LogEventsByNameMap[k], v...)
		}

		abiAsArray := []map[string]interface{}{}
		if err := json.Unmarshal([]byte(abiContent), &abiAsArray); err != nil {
			return nil, fmt.Errorf("unmarshalling abiContent as array: %w", err)
		}

		implementationABIAsArray := []map[string]interface{}{}
		if err := json.Unmarshal([]byte(implementationABIContent), &implementationABIAsArray); err != nil {
			return nil, fmt.Errorf("unmarshalling implementationABIContent as array: %w", err)
		}

		abiAsArray = append(abiAsArray, implementationABIAsArray...)

		content, err := json.Marshal(abiAsArray)
		if err != nil {
			return nil, fmt.Errorf("re-marshalling ABI")
		}
		abiContent = string(content)

		fmt.Printf("Fetched contract ABI for Implementation %s of Proxy %s\n", implementationAddress, contractAddress)
		<-wait.C
	}

	return &ABI{abi, abiContent}, nil
}

func getContractABIDirect(ctx context.Context, address string, chain *ChainConfig) (*eth.ABI, string, error) {
	url := buildDirectAPIURL(chain, address)
	fmt.Println("getting from url", url)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("new request: %w", err)
	}

	req.Header.Add("user-agent", "substreams-codegen/1.0.0")

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("getting contract Abi: %w", err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	abiContent := gjson.GetBytes(data, "Abi").String()
	if abiContent == "" {
		abiContent = gjson.GetBytes(data, "abi").String()
	}
	if abiContent == "" {
		abiContent = string(data)
	}

	ethABI, err := eth.ParseABIFromBytes([]byte(abiContent))
	if err != nil {
		return nil, "", fmt.Errorf("parsing Abi %q: %w", abiContent, err)
	}
	return ethABI, abiContent, err

}

func getContractABI(ctx context.Context, address string, chain *ChainConfig, apiKey string) (*eth.ABI, string, *time.Timer, error) {
	params := url.Values{
		"module":  {"contract"},
		"action":  {"getabi"},
		"address": {address},
	}
	url := buildFullAPIURL(chain, params, apiKey)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", nil, fmt.Errorf("new request: %w", err)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, "", nil, fmt.Errorf("getting contract Abi: %w", err)
	}
	defer res.Body.Close()

	type Response struct {
		Message string      `json:"message"` // ex: `OK-Missing/Invalid API Key, rate limit of 1/5sec applied`
		Result  interface{} `json:"result"`
	}

	var response Response
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, "", nil, fmt.Errorf("unmarshaling: %w", err)
	}

	timer := timerUntilNextCall(response.Message)

	abiContent, ok := response.Result.(string)
	if !ok {
		return nil, "", timer, fmt.Errorf(`invalid response "Result" field type, expected "string" got "%T"`, response.Result)
	}

	// Check for common API error messages that appear in the Result field
	if strings.HasPrefix(abiContent, "Contract source code not verified") {
		return nil, "", timer, fmt.Errorf("contract source code is not verified on the block explorer - you'll need to provide the ABI manually")
	}
	if strings.HasPrefix(abiContent, "Invalid Address") || strings.HasPrefix(abiContent, "Invalid address") {
		return nil, "", timer, fmt.Errorf("invalid contract address or contract does not exist at this address")
	}
	if strings.HasPrefix(abiContent, "Max rate limit reached") || strings.Contains(response.Message, "rate limit") {
		return nil, "", timer, fmt.Errorf("API rate limit reached - please wait a moment or provide an API key via %s environment variable", chain.APIKeyEnvVar)
	}

	ethABI, err := eth.ParseABIFromBytes([]byte(abiContent))
	if err != nil {
		return nil, "", timer, fmt.Errorf("failed to parse ABI from API response: %w", err)
	}
	return ethABI, abiContent, timer, err
}

// getProxyContractImplementation returns the implementation address and a timer to wait before next call
func getProxyContractImplementation(ctx context.Context, address string, chain *ChainConfig, apiKey string) (string, *time.Timer, error) {
	params := url.Values{
		"module":  {"contract"},
		"action":  {"getsourcecode"},
		"address": {address},
	}
	url := buildFullAPIURL(chain, params, apiKey)
	// check for proxy contract's implementation
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

	if err != nil {
		return "", nil, fmt.Errorf("new request: %w", err)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("getting contract Abi from etherscan: %w", err)
	}
	defer res.Body.Close()

	type Response struct {
		Message string `json:"message"` // ex: `OK-Missing/Invalid API Key, rate limit of 1/5sec applied`
		Result  []struct {
			Implementation string `json:"Implementation"`
			// ContractName string `json:"ContractName"`
		} `json:"result"`
	}

	var response Response

	bod, err := io.ReadAll(res.Body)
	if err != nil {
		return "", nil, err
	}
	if err := json.NewDecoder(bytes.NewReader(bod)).Decode(&response); err != nil {
		return "", nil, fmt.Errorf("unmarshaling %s: %w", string(bod), err)
	}

	timer := timerUntilNextCall(response.Message)

	if len(response.Result) == 0 {
		return "", timer, nil
	}

	if len(response.Result[0].Implementation) != 42 {
		return "", timer, nil
	}

	return response.Result[0].Implementation, timer, nil
}

func timerUntilNextCall(msg string) *time.Timer {
	// etherscan-specific
	if strings.HasPrefix(msg, "OK-Missing/Invalid API Key") {
		return time.NewTimer(time.Second * 5)
	}
	return time.NewTimer(time.Millisecond * 400)
}

// // Deprecated: use getContractABIFollowingProxy at the right place instead.
// func getAndSetContractABIs(ctx context.Context, contracts []*Contract, chain *ChainConfig) ([]*Contract, error) {
// 	for _, contract := range contracts {
// 		Abi, abiContent, Err := getContractABIFollowingProxy(ctx, contract.Address, chain)
// 		if Err != nil {
// 			return nil, fmt.Errorf("getting contract ABI for %s: %w", contract.Address, Err)
// 		}

// 		//fmt.Println("this is the complete abiContent after merge", abiContent)
// 		contract.abiContent = abiContent
// 		contract.Abi = Abi

// 		fmt.Printf("Fetched contract ABI for %s\n", contract.Address)
// 	}

// 	return contracts, nil
// }

// This is the NEW version, used by the new convo model.
func getContractInitialBlock(ctx context.Context, chain *ChainConfig, contractAddress string) (uint64, error) {
	if initBlock, found := chain.initialBlockCache[contractAddress]; found {
		// For testing purposes, when populating on-disk ABIs with setTestInitialBlock()
		return initBlock, nil
	}

	apiKey := os.Getenv(chain.APIKeyEnvVar)
	params := url.Values{
		"module":  {"account"},
		"action":  {"txlist"},
		"address": {contractAddress},
		"page":    {"1"},
		"offset":  {"1"},
		"sort":    {"asc"},
	}
	url := buildFullAPIURL(chain, params, apiKey)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return chain.FirstStreamableBlock, fmt.Errorf("new request: %w", err)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return chain.FirstStreamableBlock, fmt.Errorf("failed request to etherscan: %w", err)
	}
	defer res.Body.Close()

	type Response struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  []struct {
			BlockNumber string `json:"blockNumber"`
		} `json:"result"`
	}

	var response Response
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return chain.FirstStreamableBlock, fmt.Errorf("unmarshaling: %w", err)
	}

	if len(response.Result) == 0 {
		return chain.FirstStreamableBlock, fmt.Errorf("empty result from response %v", response)
	}

	blockNum, err := strconv.ParseUint(response.Result[0].BlockNumber, 10, 64)
	if err != nil {
		return chain.FirstStreamableBlock, fmt.Errorf("parsing block number: %w", err)
	}

	return blockNum, nil
}

// // Deprecated: use `getContractStartBlock` in the new convo instead.
// func getContractCreationBlock(ctx context.Context, contracts []*Contract, chain *ChainConfig) (uint64, error) {
// 	// TURN this into a SINGLE contract request, and return the start block
// 	var lowestStartBlock uint64 = math.MaxUint64
// 	for _, contract := range contracts {
// 		req, Err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api?module=account&action=txlist&address=%s&page=1&offset=1&sort=asc&apikey=%s", chain.ApiEndpoint, contract.Address, etherscanAPIKey), nil)
// 		if Err != nil {
// 			return 0, fmt.Errorf("new request: %w", Err)
// 		}

// 		res, Err := httpClient.Do(req)
// 		if Err != nil {
// 			return 0, fmt.Errorf("failed request to etherscan: %w", Err)
// 		}
// 		defer res.Body.Close()

// 		type Response struct {
// 			Status  string `json:"status"`
// 			Message string `json:"message"`
// 			Result  []struct {
// 				BlockNumber string `json:"blockNumber"`
// 			} `json:"result"`
// 		}

// 		var response Response
// 		if Err := json.NewDecoder(res.Body).Decode(&response); Err != nil {
// 			return 0, fmt.Errorf("unmarshaling: %w", Err)
// 		}

// 		if len(response.Result) == 0 {
// 			return 0, fmt.Errorf("empty result from response %v", response)
// 		}

// 		<-timerUntilNextCall(response.Message).C

// 		blockNum, Err := strconv.ParseUint(response.Result[0].BlockNumber, 10, 64)
// 		if Err != nil {
// 			return 0, fmt.Errorf("parsing block number: %w", Err)
// 		}

// 		if blockNum < lowestStartBlock {
// 			lowestStartBlock = blockNum
// 		}

// 		fmt.Printf("Fetched initial block %d for %s (lowest %d)\n", blockNum, contract.GetAddress(), lowestStartBlock)
// 	}
// 	return lowestStartBlock, nil
// }

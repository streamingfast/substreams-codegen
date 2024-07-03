package ethfull

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/streamingfast/dhttp"
	"github.com/streamingfast/eth-go"
)

var etherscanAPIKey = os.Getenv("SUBDEV_ETHERSCAN_API_KEY")

var httpClient = http.Client{
	Transport: dhttp.NewLoggingRoundTripper(zlog, tracer, http.DefaultTransport),
	Timeout:   30 * time.Second,
}

func getContractABIFollowingProxy(ctx context.Context, contractAddress string, chain *ChainConfig) (*ABI, error) {
	if cachedABI := chain.abiCache[contractAddress]; cachedABI != nil {
		// For testing purposes, when populating on-disk ABIs with setTestABI()
		return cachedABI, nil
	}

	abi, abiContent, wait, err := getContractABI(ctx, contractAddress, chain.ApiEndpoint)
	if err != nil {
		return nil, err
	}

	<-wait.C
	implementationAddress, wait, err := getProxyContractImplementation(ctx, contractAddress, chain.ApiEndpoint)
	if err != nil {
		return nil, err
	}
	<-wait.C

	if implementationAddress != "" {
		implementationABI, implementationABIContent, wait, err := getContractABI(ctx, implementationAddress, chain.ApiEndpoint)
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

func getContractABI(ctx context.Context, address string, endpoint string) (*eth.ABI, string, *time.Timer, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api?module=contract&action=getabi&address=%s&apiKey=%s", endpoint, address, etherscanAPIKey), nil)
	if err != nil {
		return nil, "", nil, fmt.Errorf("new request: %w", err)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, "", nil, fmt.Errorf("getting contract abi: %w", err)
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

	ethABI, err := eth.ParseABIFromBytes([]byte(abiContent))
	if err != nil {
		return nil, "", timer, fmt.Errorf("parsing abi %q: %w", abiContent, err)
	}
	return ethABI, abiContent, timer, err
}

// getProxyContractImplementation returns the implementation address and a timer to wait before next call
func getProxyContractImplementation(ctx context.Context, address string, endpoint string) (string, *time.Timer, error) {
	// check for proxy contract's implementation
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api?module=contract&action=getsourcecode&address=%s&apiKey=%s", endpoint, address, etherscanAPIKey), nil)

	if err != nil {
		return "", nil, fmt.Errorf("new request: %w", err)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("getting contract abi from etherscan: %w", err)
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
// 		abi, abiContent, err := getContractABIFollowingProxy(ctx, contract.Address, chain)
// 		if err != nil {
// 			return nil, fmt.Errorf("getting contract ABI for %s: %w", contract.Address, err)
// 		}

// 		//fmt.Println("this is the complete abiContent after merge", abiContent)
// 		contract.abiContent = abiContent
// 		contract.abi = abi

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

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api?module=account&action=txlist&address=%s&page=1&offset=1&sort=asc&apikey=%s", chain.ApiEndpoint, contractAddress, etherscanAPIKey), nil)
	if err != nil {
		return 0, fmt.Errorf("new request: %w", err)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed request to etherscan: %w", err)
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
		return 0, fmt.Errorf("unmarshaling: %w", err)
	}

	if len(response.Result) == 0 {
		return 0, fmt.Errorf("empty result from response %v", response)
	}

	blockNum, err := strconv.ParseUint(response.Result[0].BlockNumber, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing block number: %w", err)
	}

	return blockNum, nil
}

// // Deprecated: use `getContractStartBlock` in the new convo instead.
// func getContractCreationBlock(ctx context.Context, contracts []*Contract, chain *ChainConfig) (uint64, error) {
// 	// TURN this into a SINGLE contract request, and return the start block
// 	var lowestStartBlock uint64 = math.MaxUint64
// 	for _, contract := range contracts {
// 		req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api?module=account&action=txlist&address=%s&page=1&offset=1&sort=asc&apikey=%s", chain.ApiEndpoint, contract.Address, etherscanAPIKey), nil)
// 		if err != nil {
// 			return 0, fmt.Errorf("new request: %w", err)
// 		}

// 		res, err := httpClient.Do(req)
// 		if err != nil {
// 			return 0, fmt.Errorf("failed request to etherscan: %w", err)
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
// 		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
// 			return 0, fmt.Errorf("unmarshaling: %w", err)
// 		}

// 		if len(response.Result) == 0 {
// 			return 0, fmt.Errorf("empty result from response %v", response)
// 		}

// 		<-timerUntilNextCall(response.Message).C

// 		blockNum, err := strconv.ParseUint(response.Result[0].BlockNumber, 10, 64)
// 		if err != nil {
// 			return 0, fmt.Errorf("parsing block number: %w", err)
// 		}

// 		if blockNum < lowestStartBlock {
// 			lowestStartBlock = blockNum
// 		}

// 		fmt.Printf("Fetched initial block %d for %s (lowest %d)\n", blockNum, contract.GetAddress(), lowestStartBlock)
// 	}
// 	return lowestStartBlock, nil
// }

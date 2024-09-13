package starknet_events

import (
	starknetABI "github.com/dipdup-io/starknet-go-api/pkg/abi"
	"github.com/streamingfast/substreams-codegen/loop"
)

type ABI struct {
	decodedAbi *starknetABI.Abi
	raw        string
}

type StarknetABI struct {
}

func CmdDecodeABI(contract *Contract) loop.Cmd {
	return func() loop.Msg {
		contractABI := starknetABI.Abi{}
		err := contractABI.UnmarshalJSON(contract.RawABI)
		if err != nil {
			panic("decoding contract abi")
		}

		return ReturnRunDecodeContractABI{Abi: &ABI{&contractABI, string(contract.RawABI)}, Err: err}
	}
}

package starknet_events

import (
	"encoding/json"

	"github.com/streamingfast/substreams-codegen/loop"
)

type ABI struct {
	decodedEvents StarknetEvents
	raw           string
}

type StarknetEvents []*StarknetEvent

type StarknetEvent struct {
	CommonAttribute

	Variants []CommonAttribute `json:"variants"`
}

type OtherItem struct {
	CommonAttribute
}
type CommonAttribute struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Kind string `json:"kind"`
}

const (
	EventType = "event"
)

func (s *StarknetEvents) ExtractEvents(data []byte) error {
	var Attributes []CommonAttribute
	if err := json.Unmarshal(data, &Attributes); err != nil {
		return err
	}

	items := make([]interface{}, 0)

	for _, attribute := range Attributes {
		switch attribute.Type {
		case EventType:
			items = append(items, &StarknetEvent{})
		default:
			items = append(items, &OtherItem{})
		}
	}

	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}

	for _, item := range items {
		switch i := item.(type) {
		case *StarknetEvent:
			*s = append(*s, i)
		default:
			continue
		}
	}

	return nil
}

func CmdDecodeABI(contract *Contract) loop.Cmd {
	return func() loop.Msg {
		events := StarknetEvents{}
		err := events.ExtractEvents(contract.RawABI)
		if err != nil {
			panic("decoding contract abi")
		}

		return ReturnRunDecodeContractABI{Abi: &ABI{events, string(contract.RawABI)}, Err: err}
	}
}

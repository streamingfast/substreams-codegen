package starknet_events

import (
	"encoding/json"
	"fmt"

	"github.com/streamingfast/substreams-codegen/loop"
)

type ABI struct {
	decodedEvents StarknetEvents
	// Those items are processed to set aliases when generating the abi in Rust
	otherItems StarknetOtherItems
	raw        string
}

type StarknetEvents []*StarknetEvent

type StarknetOtherItems []*OtherItem

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

func (a *ABI) ExtractItemsFromABI(data []byte) error {
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
			a.decodedEvents = append(a.decodedEvents, i)
		case *OtherItem:
			a.otherItems = append(a.otherItems, i)
		default:
			continue
		}
	}

	return nil
}

func CmdDecodeABI(contract *Contract) loop.Cmd {
	return func() loop.Msg {
		abi := &ABI{
			decodedEvents: StarknetEvents{},
			otherItems:    StarknetOtherItems{},
			raw:           string(contract.RawABI),
		}

		err := abi.ExtractItemsFromABI(contract.RawABI)
		if err != nil {
			return ReturnRunDecodeContractABI{Abi: abi, Err: fmt.Errorf("extract items from ABI: %w", err)}
		}

		return ReturnRunDecodeContractABI{Abi: abi}
	}
}

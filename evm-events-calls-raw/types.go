package evm_events_calls_raw

import (
	pbconvo "github.com/streamingfast/substreams-codegen/pb/sf/codegen/conversation/v1"
)

type MsgStart struct{ pbconvo.UserInput_Start }

type AskProjectName struct{}
type InputProjectName struct{ pbconvo.UserInput_TextInput }

type AskChainName struct{}
type MsgInvalidChainName struct{}
type InputChainName struct{ pbconvo.UserInput_Selection }

type StartFirstContract struct{} // Start asking for contract inputs

type MsgContractSwitch struct{}

type AskContractAddress struct{}
type MsgInvalidContractAddress struct {
	Err error
}
type InputContractAddress struct{ pbconvo.UserInput_TextInput }

type AskContractName struct{}
type MsgInvalidContractName struct {
	Err error
}
type InputContractName struct{ pbconvo.UserInput_TextInput }

type AskContractTrackWhat struct{}
type InputContractTrackWhat struct{ pbconvo.UserInput_Selection }

type AskContractInitialBlock struct{}
type InputContractInitialBlock struct{ pbconvo.UserInput_TextInput }
type FetchContractInitialBlock struct{}
type ReturnFetchContractInitialBlock struct {
	InitialBlock uint64
	Err          error
}

type SetContractInitialBlock struct{ InitialBlock uint64 }

type AskAddContract struct{}
type InputAddContract struct{ pbconvo.UserInput_Confirmation }

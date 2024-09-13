package starknet_events

import (
	pbconvo "github.com/streamingfast/substreams-codegen/pb/sf/codegen/conversation/v1"
)

type MsgStart struct{ pbconvo.UserInput_Start }

type AskProjectName struct{}
type InputProjectName struct{ pbconvo.UserInput_TextInput }

type AskChainName struct{}

type InputChainName struct{ pbconvo.UserInput_Selection }

// type StartFirstContract struct{} // Start asking for contract inputs
type MsgContractSwitch struct{}

type AskContractAddress struct{}
type AskEventAddress struct{}
type InputEventAddress struct{ pbconvo.UserInput_TextInput }

type MsgInvalidContractAddress struct {
	Err error
}

type MsgInvalidEventAddress struct {
	Err error
}

type InputContractAddress struct{ pbconvo.UserInput_TextInput }

type AskContractName struct{}
type MsgInvalidContractName struct {
	Err error
}
type InputContractName struct{ pbconvo.UserInput_TextInput }
type AskContractInitialBlock struct{}
type InputContractInitialBlock struct{ pbconvo.UserInput_TextInput }
type SetContractInitialBlock struct{ InitialBlock uint64 }

type FetchContractABI struct{}
type ReturnFetchContractABI struct {
	abi string
	err error
}

type StartFirstContract struct{} // Start asking for contract inputs
type AskContractABI struct{}
type InputContractABI struct{ pbconvo.UserInput_TextInput }

type RunDecodeContractABI struct{}
type ReturnRunDecodeContractABI struct {
	Abi *ABI
	Err error
}

type AskConfirmContractABI struct{}
type InputConfirmContractABI struct{ pbconvo.UserInput_Confirmation }

//type FetchContractInitialBlock struct{}
//type ReturnFetchContractInitialBlock struct {
//	InitialBlock uint64
//	Err          error
//}

type AskAddContract struct{}
type InputAddContract struct{ pbconvo.UserInput_Confirmation }

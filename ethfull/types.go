package ethfull

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

type AskDynamicContractAddress struct{}

type InputDynamicContractAddress struct{ pbconvo.UserInput_TextInput }

type AskContractName struct{}
type MsgInvalidContractName struct {
	Err error
}
type InputContractName struct{ pbconvo.UserInput_TextInput }

type AskDynamicContractName struct{}
type MsgInvalidDynamicContractName struct {
	Err error
}
type InputDynamicContractName struct{ pbconvo.UserInput_TextInput }

type AskContractTrackWhat struct{}
type InputContractTrackWhat struct{ pbconvo.UserInput_Selection }

type AskDynamicContractTrackWhat struct{}
type InputDynamicContractTrackWhat struct{ pbconvo.UserInput_Selection }

type FetchContractABI struct{}
type ReturnFetchContractABI struct {
	abi string
	err error
}

type FetchDynamicContractABI struct{}
type ReturnFetchDynamicContractABI struct {
	abi string
	err error
}

type AskContractABI struct{}
type InputContractABI struct{ pbconvo.UserInput_TextInput }

type AskDynamicContractABI struct{}
type InputDynamicContractABI struct{ pbconvo.UserInput_TextInput }

type RunDecodeContractABI struct{}
type ReturnRunDecodeContractABI struct {
	abi *ABI
	err error
}

type RunDecodeDynamicContractABI struct{}
type ReturnRunDecodeDynamicContractABI struct {
	abi *ABI
	err error
}

type AskConfirmContractABI struct{}
type InputConfirmContractABI struct{ pbconvo.UserInput_Confirmation }

type AskContractInitialBlock struct{}
type InputContractInitialBlock struct{ pbconvo.UserInput_TextInput }
type FetchContractInitialBlock struct{}
type ReturnFetchContractInitialBlock struct {
	InitialBlock uint64
	Err          error
}

type SetContractInitialBlock struct{ InitialBlock uint64 }

type AskContractIsFactory struct{}
type InputContractIsFactory struct{ pbconvo.UserInput_Confirmation }

type AskFactoryCreationEvent struct{}
type InputFactoryCreationEvent struct{ pbconvo.UserInput_Selection }

type AskFactoryCreationEventField struct{}
type InputFactoryCreationEventField struct{ pbconvo.UserInput_Selection }

type AskAddContract struct{}
type InputAddContract struct{ pbconvo.UserInput_Confirmation }

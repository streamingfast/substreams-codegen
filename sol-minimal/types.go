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

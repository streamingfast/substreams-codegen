package injective_events

import pbconvo "github.com/streamingfast/substreams-codegen/pb/sf/codegen/conversation/v1"

type AskInitialStartBlockType struct{}
type InputAskInitialStartBlockType struct{ pbconvo.UserInput_TextInput }

type AskDataType struct{}
type InputDataType struct{ pbconvo.UserInput_Selection }

type AskEventType struct{}
type InputEventType struct{ pbconvo.UserInput_TextInput }

type AskEventAttribute struct{}
type InputEventAttribute struct{ pbconvo.UserInput_TextInput }

type AskAnotherEventType struct{}
type InputAskAnotherEventType struct{ pbconvo.UserInput_Selection }

type MsgEventSwitch struct{}

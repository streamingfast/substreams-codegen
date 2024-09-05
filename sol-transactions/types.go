package soltransactions

import pbconvo "github.com/streamingfast/substreams-codegen/pb/sf/codegen/conversation/v1"

type AskDataType struct{}
type InputDataType struct{ pbconvo.UserInput_Selection }

type AskProgramId struct{}
type InputProgramId struct{ pbconvo.UserInput_TextInput }

type ShowInstructions struct{}
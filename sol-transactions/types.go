package soltransactions

import pbconvo "github.com/streamingfast/substreams-codegen/pb/sf/codegen/conversation/v1"

type AskFilter struct{}
type InputFilter struct{ pbconvo.UserInput_TextInput }
type ShowInstructions struct{}
package soltransactions

import pbconvo "github.com/streamingfast/substreams-codegen/pb/sf/codegen/conversation/v1"

type AskExtrinsicId struct{}
type InputExtrinsicId struct{ pbconvo.UserInput_TextInput }
type ShowInstructions struct{}

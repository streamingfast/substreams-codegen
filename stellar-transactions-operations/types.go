package stellartransactionsoperations

import pbconvo "github.com/streamingfast/substreams-codegen/pb/sf/codegen/conversation/v1"

type AskFilterType struct {}
type InputFilterType struct { pbconvo.UserInput_Selection }

type AskFilter struct {}
type InputFilter struct { pbconvo.UserInput_TextInput }

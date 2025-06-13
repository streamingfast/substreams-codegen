package trontransactions

import pbconvo "github.com/streamingfast/substreams-codegen/pb/sf/codegen/conversation/v1"

type AskFilter struct {}
type InvalidFilter struct { Err error }
type InputFilter struct { pbconvo.UserInput_TextInput }
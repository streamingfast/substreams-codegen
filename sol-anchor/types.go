package solanchor

import pbconvo "github.com/streamingfast/substreams-codegen/pb/sf/codegen/conversation/v1"

type AskIDLFormat struct{}
type InputIDLFormat struct{ pbconvo.UserInput_Selection }

type AskIDLJSON struct{}
type InputIDLJSON struct{ pbconvo.UserInput_TextInput }

type AskIDLFile struct{}
type InputIDLFile struct{ pbconvo.UserInput_LocalFile }

type AskProgramID struct{}
type InputProgramID struct{ pbconvo.UserInput_TextInput }
type AskConfirmIDL struct{}
type InputConfirmIDL struct{ pbconvo.UserInput_Confirmation }

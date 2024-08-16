package starknetsql

import pbconvo "github.com/streamingfast/substreams-codegen/pb/sf/codegen/conversation/v1"

type AskTransactionFilter struct{}
type InputTransactionFilter struct{ pbconvo.UserInput_TextInput }

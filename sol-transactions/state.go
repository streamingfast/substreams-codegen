package soltransactions

import (
	"strings"
	"time"
)

const TRANSACTIONS_TYPE = "transactions_type"
const INSTRUCTIONS_TYPE = "instructions_type"

type Project struct {
	Name            string `json:"name"`
	ChainName       string `json:"chainName"`
	Compile         bool   `json:"compile,omitempty"` // optional field to write in state and automatically compile with no confirmation.
	Download        bool   `json:"download,omitempty"`
	InitialBlock    uint64 `json:"initialBlock,omitempty"`
	InitialBlockSet bool   `json:"initialBlockSet,omitempty"`
	DataType        string `json:"dataType,omitempty"`
	ProgramId       string `json:"programId,omitempty"`

	// Remote build part removed for the moment
	// confirmDoCompile       bool
	// confirmDownloadOnly    bool

	generatedCodeCompleted bool
	compilingBuild         bool
	projectFiles           map[string][]byte

	buildStarted time.Time
}

func (p *Project) ModuleName() string {
	if p.DataType == TRANSACTIONS_TYPE {
		return "map_filtered_transactions"
	}

	return "map_filtered_instructions"
}
func (p *Project) KebabName() string { return strings.ReplaceAll(p.Name, "_", "-") }

func (p *Project) IsTransactionsDataType() bool { return p.DataType == TRANSACTIONS_TYPE }
func (p *Project) IsInstructionsDataType() bool { return p.DataType == INSTRUCTIONS_TYPE }

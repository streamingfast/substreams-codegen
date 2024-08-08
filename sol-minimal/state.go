package ethfull

import (
	"strings"
	"time"
)

type Project struct {
	Name      string `json:"name"`
	ChainName string `json:"chainName"`
	Compile   bool   `json:"compile,omitempty"` // optional field to write in state and automatically compile with no confirmation.
	Download  bool   `json:"download,omitempty"`

	currentContractIdx     int
	confirmEnoughContracts bool
	confirmDoCompile       bool
	confirmDownloadOnly    bool
	generatedCodeCompleted bool
	compilingBuild         bool
	projectFiles           map[string][]byte

	buildStarted time.Time

	// always set by the server
	// only for SQL projects

	SqlOutputFlavor      string `json:"sql_output_flavor,omitempty"`      // either "clickhouse" or "sql"
	SubgraphOutputFlavor string `json:"subgraph_output_flavor,omitempty"` // either "trigger" or "entity"
}

func (p *Project) ModuleName() string { return strings.ReplaceAll(p.Name, "-", "_") }
func (p *Project) KebabName() string  { return strings.ReplaceAll(p.Name, "_", "-") }
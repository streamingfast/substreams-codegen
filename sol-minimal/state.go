package solminimal

import (
	"strings"
	"time"
)

type Project struct {
	Name            string `json:"name"`
	ChainName       string `json:"chainName"`
	Compile         bool   `json:"compile,omitempty"` // optional field to write in state and automatically compile with no confirmation.
	Download        bool   `json:"download,omitempty"`
	InitialBlock    uint64 `json:"initialBlock,omitempty"`
	InitialBlockSet bool   `json:"initialBlockSet,omitempty"`

	// Remote build part removed for the moment
	// confirmDoCompile       bool
	// confirmDownloadOnly    bool

	generatedCodeCompleted bool
	compilingBuild         bool
	projectFiles           map[string][]byte

	buildStarted time.Time
}

func (p *Project) ModuleName() string { return strings.ReplaceAll(p.Name, "-", "_") }
func (p *Project) KebabName() string  { return strings.ReplaceAll(p.Name, "_", "-") }

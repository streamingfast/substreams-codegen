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

	confirmDoCompile       bool
	confirmDownloadOnly    bool
	generatedCodeCompleted bool
	compilingBuild         bool
	projectFiles           map[string][]byte

	buildStarted time.Time
}

func (p *Project) ChainConfig() *ChainConfig { return ChainConfigByID[p.ChainName] }
func (p *Project) ChainEndpoint() string     { return ChainConfigByID[p.ChainName].FirehoseEndpoint }

func (p *Project) ModuleName() string { return strings.ReplaceAll(p.Name, "-", "_") }
func (p *Project) KebabName() string  { return strings.ReplaceAll(p.Name, "_", "-") }

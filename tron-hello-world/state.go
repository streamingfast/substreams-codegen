package tronhelloworld

import (
	"strings"

	registry "github.com/pinax-network/graph-networks-libs/packages/golang/lib"
)

type Project struct {
	Name     string            `json:"name"`
	Compile  bool              `json:"compile,omitempty"` // optional field to write in state and automatically compile with no confirmation.
	Download bool              `json:"download,omitempty"`
	Chain    *registry.Network `json:"chain,omitempty"`
}

func (p *Project) ModuleName() string { return strings.ReplaceAll(p.Name, "-", "_") }
func (p *Project) KebabName() string  { return strings.ReplaceAll(p.Name, "_", "-") }

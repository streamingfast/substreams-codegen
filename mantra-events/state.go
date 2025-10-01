package mantra_events

import (
	"embed"
	"fmt"
	"sort"
	"strings"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/base"
)

//go:embed templates/*
var templatesFS embed.FS

const EVENTS_DATA_TYPE = "events"
const EVENT_GROUPS_DATA_TYPE = "event_groups"
const TRXS_DATA_TYPE = "transactions"

type Project struct {
	base.ConversationState
	InitialBlock    uint64       `json:"initialBlock,omitempty"`
	InitialBlockSet bool         `json:"initialBlockSet,omitempty"`
	DataType        string       `json:"dataType,omitempty"`
	EventDescs      []*eventDesc `json:"messageTypes,omitempty"`
	currentEventIdx int
	EventsComplete  bool `json:"eventsComplete,omitempty"`
}

type eventDesc struct {
	EventType  string            `json:"eventType"`
	Attributes map[string]string `json:"attributes"`
	Incomplete bool              `json:"incomplete,omitempty"`
}

func (p *Project) Generate() codegen.ReturnGenerate {
	return codegen.GenerateTemplateTree(p, templatesFS, map[string]string{
		".gitignore.gotmpl":             ".gitignore",
		"README.md.gotmpl":              "README.md",
		"substreams.yaml.gotmpl":        "substreams.yaml",
		"common-templates/buf.gen.yaml": "buf.gen.yaml",
		"Cargo.toml.gotmpl":             "Cargo.toml",
		"src/lib.rs.gotmpl":             "src/lib.rs",
		"proto/mydata.proto.gotmpl":     "proto/mydata.proto",
	})
}

func (e eventDesc) GetEventQuery() string {
	attributes := make([]string, 0, len(e.Attributes))
	for k, v := range e.Attributes {
		switch {
		case v == "":
			attributes = append(attributes, fmt.Sprintf("attr:%s", k))
		default:
			attributes = append(attributes, fmt.Sprintf("attr:%s:%s", k, v))
		}
	}
	if len(attributes) == 0 {
		return fmt.Sprintf("type:%s", e.EventType)
	}

	sort.Strings(attributes)

	return fmt.Sprintf("(type:%s && (%s))", e.EventType, strings.Join(attributes, " && "))
}
func (e eventDesc) GetEventIndexQuery() string {
	attributes := make([]string, 0, len(e.Attributes))
	for k := range e.Attributes {
		attributes = append(attributes, fmt.Sprintf("attr:%s", k))
	}
	if len(attributes) == 0 {
		return fmt.Sprintf("type:%s", e.EventType)
	}
	return fmt.Sprintf("(type:%s && (%s))", e.EventType, strings.Join(attributes, " && "))
}

func (p *Project) GetEventsQuery() string {
	outs := make([]string, 0, len(p.EventDescs))
	for _, desc := range p.EventDescs {
		outs = append(outs, desc.GetEventQuery())
	}
	return strings.Join(outs, " || ")
}

func (p *Project) GetEventsIndexQuery() string {
	outs := make([]string, 0, len(p.EventDescs))
	for _, desc := range p.EventDescs {
		outs = append(outs, desc.GetEventIndexQuery())
	}
	return strings.Join(outs, " || ")
}

func (p *Project) IsEvents() bool {
	return p.DataType == EVENTS_DATA_TYPE
}

func (p *Project) IsEventGroups() bool {
	return p.DataType == EVENT_GROUPS_DATA_TYPE
}

func (p *Project) IsTransactions() bool {
	return p.DataType == TRXS_DATA_TYPE
}

func (p *Project) HasAttributeValues() bool {
	for _, evt := range p.EventDescs {
		for _, val := range evt.Attributes {
			if val != "" {
				return true
			}
		}
	}
	return false
}

func (p *Project) ModuleName() string {
	return fmt.Sprintf("map_%s", p.DataType)
}

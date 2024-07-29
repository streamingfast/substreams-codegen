package starknet

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/golang-cz/textcase"
)

const EVENTS_DATA_TYPE = "events"
const EVENT_GROUPS_DATA_TYPE = "event_groups"
const TRXS_DATA_TYPE = "transactions"

type eventDesc struct {
	EventType  string            `json:"eventType"`
	Attributes map[string]string `json:"attributes"`
	Incomplete bool              `json:"incomplete,omitempty"`
}

type Project struct {
	Name              string       `json:"name"`
	ChainName         string       `json:"chainName"`
	InitialBlock      uint64       `json:"initialBlock,omitempty"`
	InitialBlockSet   bool         `json:"initialBlockSet,omitempty"`
	Compile           bool         `json:"compile,omitempty"` // optional field to write in state and automatically compile with no confirmation.
	Download          bool         `json:"download,omitempty"`
	DataType          string       `json:"dataType,omitempty"`
	TransactionFilter string       `json:"transactionFilter,omitempty"`
	EventDescs        []*eventDesc `json:"messageTypes,omitempty"`
	currentEventIdx   int
	eventsComplete    bool

	filterAsked bool

	confirmDoCompile       bool
	confirmDownloadOnly    bool
	generatedCodeCompleted bool

	compilingBuild bool
	projectZip     []byte
	sourceZip      []byte

	buildStarted time.Time

	// always set by the server
	outputType outputType

	SqlOutputFlavor string `json:"sql_output_flavor,omitempty"` // either "clickhouse" or "sql"
}

func (p *Project) ChainConfig() *ChainConfig { return ChainConfigByID[p.ChainName] }
func (p *Project) ChainEndpoint() string     { return ChainConfigByID[p.ChainName].FirehoseEndpoint }
func (p *Project) KebabName() string         { return strings.ReplaceAll(p.Name, "_", "-") }

func (p *Project) SubgraphProjectName() string         { return textcase.KebabCase(p.Name) }
func (p *Project) SQLImportVersion() string            { return "1.0.7" }
func (p *Project) GraphImportVersion() string          { return "0.1.0" }
func (p *Project) DatabaseChangeImportVersion() string { return "1.2.1" }
func (p *Project) EntityChangeImportVersion() string   { return "1.1.0" }
func (p *Project) StarknetFoundationalVersion() string { return "0.1.3" }

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

func validateIncomingState(p *Project) error {
	return nil
}
package codegen

import "sort"

var Registry = make(map[string]*ConversationHandler)

type ConversationHandler struct {
	ID          string
	Title       string
	Description string
	Weight      int
	Factory     ConversationFactory
}

func RegisterConversation(conversationID string, title, description string, newFunc ConversationFactory, weight int) {
	handler := ConversationHandler{
		ID:          conversationID,
		Title:       title,
		Description: description,
		Factory:     newFunc,
		Weight:      weight,
	}
	Registry[conversationID] = &handler
}

func ListConversationHandlers() []*ConversationHandler {
	var handlers []*ConversationHandler
	for _, handler := range Registry {
		handlers = append(handlers, handler)
	}
	sort.Slice(handlers, func(i, j int) bool {
		return handlers[i].Weight > handlers[j].Weight // heighest weight first
	})
	return handlers
}

var FileDescriptions = map[string]string{
	"contract.proto":      "File containing the contract proto definition",
	"build.rs":            "File containing the build script for the project",
	"Cargo.toml":          "Cargo manifest file, a configuration file which defines the project",
	"substreams.yaml":     "Substreams manifest, a configuration file which defines the different modules",
	"rust-toolchain.toml": "File containing the rust toolchain version",
	"lib.rs":              "Substreams modules definition code in Rust",
	".gitignore":          "File containing the gitignore rules",
	"mod.rs":              "Rust module definitions file",
	"contract.abi.json":   "File containing the contract ABI definition",
}

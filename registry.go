package codegen

import (
	"sort"
)

var Registry = make(map[string]*ConversationHandler)

type ConversationHandler struct {
	ID          string
	Title       string
	Description string

	// Weight is used to sort the list of conversations, higher weight first
	// EVM: 80+
	// Injective; 70+
	// Solana: 60+
	// Starknet: 50+
	// Substrate: 40+

	Weight int

	Factory ConversationFactory
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
	// Git files
	".gitignore": "File containing the gitignore rules which ignores the target/ directory and any spkg package produced",

	// Documentation files
	"README.md":       "File containing instructions on how to compile and run the project",
	"CONTRIBUTING.md": "File containing the project contributing guidelines",

	// Rust files
	"src/lib.rs":    "This is the main entrypoint file where your modules' code lives. Modify it and run `substreams build` to rebuild your package",
	"src/pb/mod.rs": "Rust module definitions file",
	"src/build.rs":  "This file contains any build step needed to compile the project, think abi generation",

	// Proto files
	"proto/contract.proto": "This file contains the data models used by your Substreams modules",
	"proto/mydata.proto":   "This file contains the data models used by your Substreams modules",

	// Toml files
	"Cargo.toml":          "Cargo manifest file, a configuration file which defines the project and it's dependencies",
	"rust-toolchain.toml": "File containing the rust toolchain version and what build target to use",

	// Substreams yaml files
	"substreams.yaml": "Substreams manifest, a configuration file which defines the different modules",

	// ABI JSON files
	"abi/contract.abi.json": "File containing the contract ABI definition",
}

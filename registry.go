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
	// do NOT document files like `.gitignore`, `README.md`, or things towards which it is not ESSENTIAL that we drag attention to.
	// do NOT document files like `COnfig.toml` or other Rust artifacts. A Rust developer will know. It's not our job HERE to introduce them to these concepts.

	"src/lib.rs": "Modify this file to reflect your needs. This is the main entrypoint.",

	"proto/contract.proto": "This file contains the data models used by your Substreams modules",
	"proto/mydata.proto":   "Modify this file to reflect your needs. It contains protobuf models.",

	"substreams.yaml": "Substreams manifest, a configuration file which defines the different modules",
}

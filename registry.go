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

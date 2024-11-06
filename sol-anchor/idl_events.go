package solanchor

// --- EVENTS

type Event struct {
	Name   string  `json:"name"`
	Fields []Field `json:"fields"`
}

func (e *Event) SnakeCaseName() string {
	return toSnakeCase(e.Name, true)
}

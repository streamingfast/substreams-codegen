package solanchor

import "fmt"

// --- EVENTS

type Event struct {
	Name          string  `json:"name"`
	Fields        []Field `json:"fields"`
	Discriminator []uint8 `json:"discriminator"`
}

func (e *Event) SnakeCaseName() string {
	return toSnakeCase(e.Name, true)
}

func (e *Event) PrintDiscriminator() string {
	numbersAsString := ""
	for i, n := range e.Discriminator {
		numbersAsString += fmt.Sprintf("%du8", n)

		if i < len(e.Discriminator)-1 {
			numbersAsString += ","
		}
	}
	return fmt.Sprintf("&[%s]", numbersAsString)
}

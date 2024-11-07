package solanchor

import (
	"github.com/golang-cz/textcase"
)

// --- INSTRUCTIONS
type Instruction struct {
	Name string  `json:"name"`
	Args []Field `json:"args"`
}

func (i *Instruction) PascalCaseName() string {
	return textcase.PascalCase(i.Name)
}

func (i *Instruction) SnakeCaseName() string {
	return toSnakeCase(i.Name, true)
}
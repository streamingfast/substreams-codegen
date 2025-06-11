package solanchor

import (
	"fmt"

	"github.com/golang-cz/textcase"
)

type InstructionAccount struct {
	Name     string `json:"name"`
	Writable bool   `json:"writable"`
	Signer   bool   `json:"signer"`
	Address  string `json:"address"`
}

func (a *InstructionAccount) SnakeCaseName() string {
	return toSnakeCase(a.Name, true)
}

// --- INSTRUCTIONS
type Instruction struct {
	Name          string               `json:"name"`
	Args          []Field              `json:"args"`
	Accounts      []InstructionAccount `json:"accounts"`
	Discriminator []uint8              `json:"discriminator"`
}

func (i *Instruction) PascalCaseName() string {
	return textcase.PascalCase(i.Name)
}

func (i *Instruction) SnakeCaseName() string {
	return toSnakeCase(i.Name, true)
}

func (e *Instruction) PrintDiscriminator() string {
	numbersAsString := ""
	for i, n := range e.Discriminator {
		numbersAsString += fmt.Sprintf("%du8", n)

		if i < len(e.Discriminator)-1 {
			numbersAsString += ","
		}
	}
	return fmt.Sprintf("&[%s]", numbersAsString)
}

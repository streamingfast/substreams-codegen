package solminimal

import (
	"testing"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
	"github.com/stretchr/testify/assert"
)

func TestConvoNextStep(t *testing.T) {
	p := &Project{}
	next := func() loop.Msg {
		return p.NextStep()()
	}

	assert.Equal(t, codegen.AskProjectName{}, next())
	p.Name = "my-proj"

	projectFiles, err := p.Generate()
	assert.NoError(t, err)
	assert.NotEmpty(t, projectFiles)
}

package ethminimal

import (
	"testing"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
	"github.com/stretchr/testify/assert"
)

func TestConvoNextStep(t *testing.T) {
	convo := New()
	next := func() loop.Msg {
		return convo.NextStep()()
	}
	p := convo.(*Convo).State

	assert.Equal(t, codegen.AskProjectName{}, next())
	p.Name = "my-proj"

	assert.Equal(t, codegen.AskChainName{}, next())
	p.ChainName = "arbitrum"

	res := p.Generate()
	assert.NoError(t, res.Err)
	assert.NotEmpty(t, res.ProjectFiles)
}

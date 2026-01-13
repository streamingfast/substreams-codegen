package ethhelloworld

import (
	"fmt"
	"testing"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
	"github.com/stretchr/testify/assert"
)

func TestConvoNextStep(t *testing.T) {
	convo := New()
	fmt.Println(convo.GetState().GetChainName())
	next := func() loop.Msg {
		return convo.NextStep()()
	}
	p := convo.(*Convo).State

	assert.Equal(t, codegen.AskProjectName{}, next())
	p.Name = "my-proj"

	assert.Equal(t, codegen.AskChainName{}, next())
	p.ChainName = "arbitrum"

	assert.Equal(t, codegen.AskInitialStartBlockType{}, next())
	p.InitialBlock = 0
	p.InitialBlockSet = true

	assert.Equal(t, codegen.AskSubstreamsConsumptionChoice{}, next())
	sinkChoice := codegen.SubstreamsSinkChoiceSourceOnly
	p.SubstreamsSinkChoice = &sinkChoice

	assert.Equal(t, codegen.RunGenerate{}, next())

	res := p.Generate()
	assert.NoError(t, res.Err)
	assert.NotEmpty(t, res.ProjectFiles)
}

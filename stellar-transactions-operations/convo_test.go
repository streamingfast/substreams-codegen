package stellartransactionsoperations

import (
	"testing"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	p.ChainName = "stellar"

	res := p.Generate()
	assert.NoError(t, res.Err)
	assert.NotEmpty(t, res.ProjectFiles)
}

func TestIsFilterCorrect(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"hello,world,test", true},       // Valid
		{"123,456,789", true},            // Valid
		{"abc,123,xyz_edsds", true},      // Valid
		{"hello, world", false},          // Invalid (space present)
		{"hello ,world", false},          // Invalid (space present)
		{"hello,,world", false},          // Invalid (double comma)
		{"hello,world,", false},          // Invalid (trailing comma)
		{",hello,world", false},          // Invalid (leading comma)
		{"singleword", true},             // Valid (single word)
		{"", false},                      // Invalid (empty string)
		{"payment,create_account", true}, // Valid
		{"GADLRTGF4GCU2CNAHYPKAEBGQSBX2M3UYIZZJODZAVC5A5QCAE7AT66C", true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			require.Equal(t, test.expected, isFilterCorrect(test.input), "input: %s", test.input)
		})
	}
}

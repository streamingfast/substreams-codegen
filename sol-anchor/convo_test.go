package solanchor

import (
	"testing"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
	pbconvo "github.com/streamingfast/substreams-codegen/pb/sf/codegen/conversation/v1"
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

	p.idl = &IDL{
		Address:      "deadbeef",
		Events:       nil,
		Instructions: nil,
		Metadata: Metadata{
			Address: "deadbeef",
			Name:    "my contract",
		},
	}
	res := p.Generate()
	assert.NoError(t, res.Err)
	assert.NotEmpty(t, res.ProjectFiles)
}

func TestIDLParseErrorRetry(t *testing.T) {
	conv := New()
	conv.SetFactory(&codegen.MsgWrapFactory{})
	p := conv.(*Convo).State

	// Setup state
	p.Name = "test_project"
	p.ChainName = "solana-mainnet-beta"
	p.IdlFormat = "string"

	// Provide invalid JSON
	next := conv.Update(InputIDLJSON{
		UserInput_TextInput: pbconvo.UserInput_TextInput{
			Value: "{ invalid json }",
		},
	})

	// Should return a sequence with error message and ask for IDL again
	seq, ok := next().(loop.SeqMsg)
	require.True(t, ok, "expected SeqMsg")
	require.Len(t, seq, 2, "expected 2 commands in sequence")

	// First should be error message
	msg := seq[0]().(*pbconvo.SystemOutput)
	assert.Contains(t, msg.GetMessage().Markdown, "not valid JSON")

	// Second should ask for IDL again
	assert.Equal(t, AskIDLJSON{}, seq[1]())
}

func TestIDLEmptyInputRetry(t *testing.T) {
	conv := New()
	conv.SetFactory(&codegen.MsgWrapFactory{})
	p := conv.(*Convo).State

	// Setup state
	p.Name = "test_project"
	p.ChainName = "solana-mainnet-beta"
	p.IdlFormat = "string"

	// Provide empty input
	next := conv.Update(InputIDLJSON{
		UserInput_TextInput: pbconvo.UserInput_TextInput{
			Value: "",
		},
	})

	// Should return a sequence with message and ask for IDL again
	seq, ok := next().(loop.SeqMsg)
	require.True(t, ok, "expected SeqMsg")
	require.Len(t, seq, 2, "expected 2 commands in sequence")

	// First should be message about empty content
	msg := seq[0]().(*pbconvo.SystemOutput)
	assert.Contains(t, msg.GetMessage().Markdown, "empty")

	// Second should ask for IDL again
	assert.Equal(t, AskIDLJSON{}, seq[1]())
}

func TestIDLFileReadErrorRetry(t *testing.T) {
	conv := New()
	conv.SetFactory(&codegen.MsgWrapFactory{})
	p := conv.(*Convo).State

	// Setup state
	p.Name = "test_project"
	p.ChainName = "solana-mainnet-beta"
	p.IdlFormat = "file"

	// Provide file with read error
	errorMsg := "file not found"
	next := conv.Update(InputIDLFile{
		UserInput_LocalFile: pbconvo.UserInput_LocalFile{
			Error: &errorMsg,
		},
	})

	// Should return a sequence with error message and ask for IDL file again
	seq, ok := next().(loop.SeqMsg)
	require.True(t, ok, "expected SeqMsg")
	require.Len(t, seq, 2, "expected 2 commands in sequence")

	// First should be error message
	msg := seq[0]().(*pbconvo.SystemOutput)
	assert.Contains(t, msg.GetMessage().Markdown, "file not found")

	// Second should ask for IDL file again
	assert.Equal(t, AskIDLFile{}, seq[1]())
}

func TestIDLFileInvalidJSONRetry(t *testing.T) {
	conv := New()
	conv.SetFactory(&codegen.MsgWrapFactory{})
	p := conv.(*Convo).State

	// Setup state
	p.Name = "test_project"
	p.ChainName = "solana-mainnet-beta"
	p.IdlFormat = "file"

	// Provide file with invalid JSON content
	next := conv.Update(InputIDLFile{
		UserInput_LocalFile: pbconvo.UserInput_LocalFile{
			Value: []byte("not json at all"),
		},
	})

	// Should return a sequence with error message and ask for IDL file again
	seq, ok := next().(loop.SeqMsg)
	require.True(t, ok, "expected SeqMsg")
	require.Len(t, seq, 2, "expected 2 commands in sequence")

	// First should be error message about JSON
	msg := seq[0]().(*pbconvo.SystemOutput)
	assert.Contains(t, msg.GetMessage().Markdown, "not valid JSON")

	// Second should ask for IDL file again
	assert.Equal(t, AskIDLFile{}, seq[1]())
}

package evm_events_calls

import (
	"fmt"
	"testing"

	"github.com/streamingfast/eth-go"
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

	assert.Equal(t, codegen.AskChainName{}, next())

	p.ChainName = "arbitrum"

	assert.Equal(t, StartFirstContract{}, next())

	p.Contracts = append(p.Contracts, &Contract{})

	assert.Equal(t, AskContractAddress{}, next())

	p.Contracts[0].Address = "0x1231231230123123123012312312301231231230"

	assert.Equal(t, FetchContractABI{}, next())
}
func TestConvoUpdate(t *testing.T) {
	conv := New()
	conv.SetFactory(&codegen.MsgWrapFactory{})
	p := conv.(*Convo).State

	next := conv.Update(codegen.InputProjectName{UserInput_TextInput: pbconvo.UserInput_TextInput{
		Value: "my-proj",
	}})
	assert.Equal(t, "my-proj", p.Name)

	assert.Equal(t, codegen.AskChainName{}, next())
	next = conv.Update(codegen.InputChainName{UserInput_Selection: pbconvo.UserInput_Selection{
		Value: "mainnet",
	}})
	assert.Equal(t, "mainnet", p.ChainName)

	seq := next().(loop.SeqMsg)
	assert.Contains(t, seq[0]().(*pbconvo.SystemOutput).Entry.(*pbconvo.SystemOutput_Message_).Message.String(), "Ethereum Mainnet")
	assert.Equal(t, StartFirstContract{}, seq[1]())

	assert.Len(t, p.Contracts, 0)
	next = conv.Update(StartFirstContract{})
	assert.Len(t, p.Contracts, 1)

	assert.Equal(t, AskContractAddress{}, next())

	next = conv.Update(InputContractAddress{UserInput_TextInput: pbconvo.UserInput_TextInput{Value: "0x1231231230123123123012312312301231231230"}})
	assert.Equal(t, FetchContractABI{}, next())

	next = conv.Update(FetchContractABI{})
	decode := next().(ReturnFetchContractABI)

	assert.NotNil(t, decode.err)

	next = conv.Update(decode)
	seq = next().(loop.SeqMsg)
	assert.Contains(t, seq[0]().(*pbconvo.SystemOutput).Entry.(*pbconvo.SystemOutput_Message_).Message.String(), "ABI")
	assert.Equal(t, AskContractABIType{}, seq[1]())

	next = conv.Update(InputContractABIType{UserInput_Selection: pbconvo.UserInput_Selection{Value: "file"}})
	// After selecting type, it now goes directly to asking for the file (not fetch again)
	assert.Equal(t, AskContractABIFile{}, next())

	abi := "[]"
	next = conv.Update(InputContractABIFile{UserInput_LocalFile: pbconvo.UserInput_LocalFile{Value: []byte(abi)}})
	assert.Equal(t, RunDecodeContractABI{}, next())

	next = conv.Update(RunDecodeContractABI{})
	msg, ok := next().(ReturnRunDecodeContractABI)
	require.True(t, ok)
	assert.Nil(t, msg.err)

	next = conv.Update(msg)
	// Should show ABI preview before continuing
	seq = next().(loop.SeqMsg)
	require.Len(t, seq, 2)
	// First command is the preview message
	previewMsg := seq[0]().(*pbconvo.SystemOutput)
	assert.Contains(t, previewMsg.GetMessage().Markdown, "ABI") // Verify it's showing ABI info
	// Second command asks for confirmation
	assert.Equal(t, AskConfirmContractABI{}, seq[1]())

	// Confirm the ABI
	next = conv.Update(InputConfirmContractABI{UserInput_Confirmation: pbconvo.UserInput_Confirmation{Affirmative: true}})
	assert.Equal(t, FetchContractInitialBlock{}, next())

	next = conv.Update(ReturnFetchContractInitialBlock{err: fmt.Errorf("failed")})
	assert.Contains(t, next().(*pbconvo.SystemOutput).Entry.(*pbconvo.SystemOutput_TextInput_).TextInput.String(), "Please enter the contract initial block number")

	next = conv.Update(InputContractInitialBlock{UserInput_TextInput: pbconvo.UserInput_TextInput{Value: "123"}})
	assert.Equal(t, AskContractName{}, next())

	next = conv.Update(InputContractName{UserInput_TextInput: pbconvo.UserInput_TextInput{Value: "my_contract"}})
	assert.Equal(t, AskContractTrackWhat{}, next())

	next = conv.Update(InputContractTrackWhat{UserInput_Selection: pbconvo.UserInput_Selection{Value: "calls"}})
	assert.Equal(t, AskContractIsFactory{}, next())

	next = conv.Update(InputContractIsFactory{UserInput_Confirmation: pbconvo.UserInput_Confirmation{Affirmative: true}})
	assert.Equal(t, AskFactoryCreationEvent{}, next())

	next = conv.Update(InputFactoryCreationEvent{UserInput_Selection: pbconvo.UserInput_Selection{Value: "Transfer()"}})
	assert.Equal(t, AskFactoryCreationEventField{}, next())

	next = conv.Update(InputFactoryCreationEventField{UserInput_Selection: pbconvo.UserInput_Selection{Value: "0"}})
	assert.Equal(t, AskDynamicContractName{}, next())

	next = conv.Update(InputDynamicContractName{UserInput_TextInput: pbconvo.UserInput_TextInput{Value: "dyncontract"}})
	assert.Equal(t, AskDynamicContractTrackWhat{}, next())

	next = conv.Update(InputDynamicContractTrackWhat{UserInput_Selection: pbconvo.UserInput_Selection{Value: "events"}})
	assert.Equal(t, AskDynamicContractAddress{}, next())

	next = conv.Update(InputDynamicContractAddress{UserInput_TextInput: pbconvo.UserInput_TextInput{Value: "0x1231231230123123123012312312301231231232"}})
	assert.Equal(t, FetchDynamicContractABI{}, next())

	next = conv.Update(ReturnFetchDynamicContractABI{err: fmt.Errorf("failed")})
	seq = next().(loop.SeqMsg)
	assert.Equal(t, AskDynamicContractABI{}, seq[1]())

	next = conv.Update(InputDynamicContractABI{UserInput_TextInput: pbconvo.UserInput_TextInput{Value: "[]"}})
	assert.Equal(t, RunDecodeDynamicContractABI{}, next())

	next = conv.Update(ReturnRunDecodeDynamicContractABI{abi: &ABI{
		abi: &eth.ABI{},
		raw: "[]",
	}, err: nil})

	// Dynamic contract ABI preview should now be shown (our fix restored this)
	// Unlike regular contracts, dynamic contracts show preview then go directly to next step (no confirmation)
	seq = next().(loop.SeqMsg)
	require.Len(t, seq, 2)
	previewMsg = seq[0]().(*pbconvo.SystemOutput)
	assert.Contains(t, previewMsg.GetMessage().Markdown, "ABI")
	assert.Equal(t, AskAddContract{}, seq[1]())

	next = conv.Update(InputAddContract{UserInput_Confirmation: pbconvo.UserInput_Confirmation{Affirmative: false}})
	require.Equal(t, codegen.AskSubstreamsConsumptionChoice{}, next())

	next = conv.Update(codegen.InputSubstreamsConsumptionChoice{UserInput_Selection: pbconvo.UserInput_Selection{
		Value: "postgres",
		Label: "Postgres",
	}})
	assert.Equal(t, codegen.RunGenerate{}, next())

	next = conv.Update(codegen.ReturnGenerate{ProjectFiles: nil})
	seq = next().(loop.SeqMsg)

	// First part of sequence should be the download files command
	downloadMsg := seq[0]().(*pbconvo.SystemOutput)
	downloadFiles := downloadMsg.GetDownloadFiles()
	assert.NotNil(t, downloadFiles)
	assert.Len(t, downloadFiles.Files, 1)

	// Second part should trigger InputSourceDownloaded which leads to the project ready message
	next = conv.Update(codegen.InputSourceDownloaded{UserInput_TextInput: pbconvo.UserInput_TextInput{Value: "{project folder}"}})
	seq = next().(loop.SeqMsg)

	// Now the project ready message should be in the first part of this new sequence
	msg1 := seq[0]().(*pbconvo.SystemOutput)
	assert.Contains(t, msg1.GetMessage().Markdown, "substreams build")
	assert.Contains(t, msg1.GetMessage().Markdown, "substreams auth")
	assert.Contains(t, msg1.GetMessage().Markdown, "substreams gui")
	assert.Contains(t, msg1.GetMessage().Markdown, "substreams registry login")
	assert.Contains(t, msg1.GetMessage().Markdown, "substreams registry publish")

	// Second part should be the quit message
	msg2 := seq[1]()
	assert.NotNil(t, msg2)
	assert.IsType(t, loop.QuitMsg{}, msg2)
}

func unpackSeq(t *testing.T, m loop.Msg, len int) (out []loop.Msg) {
	t.Helper()
	seq, ok := m.(loop.SeqMsg)
	assert.True(t, ok)
	assert.Len(t, seq, len)
	for _, s := range seq {
		out = append(out, s())
	}
	return out
}

func TestContractNameAlreadyExists(t *testing.T) {
	conv := New()
	p := conv.(*Convo).State
	p.currentContractIdx = 0
	p.Contracts = append(p.Contracts, &Contract{
		BaseContract: BaseContract{
			Name: "test",
		},
	})

	next := conv.Update(InputContractName{pbconvo.UserInput_TextInput{Value: "test"}})

	seq := unpackSeq(t, next(), 2)

	assert.Equal(t, MsgInvalidContractName{
		Err: fmt.Errorf("contract with name test already exists in the project"),
	}, seq[0])
	assert.IsType(t, AskContractName{}, seq[1])
}

func TestDynamicContractNameAlreadyExists(t *testing.T) {
	conv := New()
	p := conv.(*Convo).State
	p.currentContractIdx = 0
	p.Contracts = append(p.Contracts, &Contract{
		BaseContract: BaseContract{
			Name: "test",
		},
	})

	next := conv.Update(InputDynamicContractName{pbconvo.UserInput_TextInput{Value: "test"}})
	seq := unpackSeq(t, next(), 2)

	assert.Equal(t, MsgInvalidDynamicContractName{
		Err: fmt.Errorf("contract with name test already exists in the project"),
	}, seq[0])

	assert.IsType(t, AskDynamicContractName{}, seq[1])
}

func TestContractAddressAlreadyExists(t *testing.T) {
	conv := New()
	p := conv.(*Convo).State
	p.currentContractIdx = 0
	p.Contracts = append(p.Contracts, &Contract{
		Address: "0x1f98431c8ad98523631ae4a59f267346ea31f984",
	})

	next := conv.Update(InputContractAddress{pbconvo.UserInput_TextInput{Value: "0x1f98431c8ad98523631ae4a59f267346ea31f984"}})
	seq := unpackSeq(t, next(), 2)

	assert.Equal(t, MsgInvalidContractAddress{
		Err: fmt.Errorf("contract address 0x1f98431c8ad98523631ae4a59f267346ea31f984 already exists in the project"),
	}, seq[0])

	assert.IsType(t, AskContractAddress{}, seq[1])
}

func TestDynamicContractAddressAlreadyExists(t *testing.T) {
	conv := New()
	p := conv.(*Convo).State
	p.currentContractIdx = 0
	p.Contracts = append(p.Contracts, &Contract{
		Address: "0x1f98431c8ad98523631ae4a59f267346ea31f984",
	})

	next := conv.Update(InputDynamicContractAddress{pbconvo.UserInput_TextInput{Value: "0x1f98431c8ad98523631ae4a59f267346ea31f984"}})
	seq := unpackSeq(t, next(), 2)

	assert.Equal(t, MsgInvalidContractAddress{
		Err: fmt.Errorf("contract address 0x1f98431c8ad98523631ae4a59f267346ea31f984 already exists in the project"),
	}, seq[0])

	assert.IsType(t, AskDynamicContractAddress{}, seq[1])
}

func TestWrappedABIFormat(t *testing.T) {
	conv := New()
	conv.SetFactory(&codegen.MsgWrapFactory{})
	p := conv.(*Convo).State

	// Setup: skip to the point where we need ABI input
	p.Name = "test-project"
	p.ChainName = "mainnet"
	p.Contracts = append(p.Contracts, &Contract{
		Address: "0x1231231230123123123012312312301231231230",
	})
	p.currentContractIdx = 0

	// Test with wrapped ABI format (Hardhat style)
	wrappedABI := `{
		"contractName": "TestContract",
		"abi": [
			{
				"anonymous": false,
				"inputs": [
					{"indexed": true, "name": "from", "type": "address"},
					{"indexed": true, "name": "to", "type": "address"},
					{"indexed": false, "name": "value", "type": "uint256"}
				],
				"name": "Transfer",
				"type": "event"
			}
		],
		"bytecode": "0x608060..."
	}`

	// Simulate file input with wrapped ABI
	next := conv.Update(InputContractABIFile{
		UserInput_LocalFile: pbconvo.UserInput_LocalFile{
			Value: []byte(wrappedABI),
		},
	})

	// Should proceed to RunDecodeContractABI
	assert.Equal(t, RunDecodeContractABI{}, next())

	// Execute the decode
	next = conv.Update(RunDecodeContractABI{})
	msg, ok := next().(ReturnRunDecodeContractABI)
	require.True(t, ok, "expected ReturnRunDecodeContractABI")
	require.NoError(t, msg.err, "should successfully decode wrapped ABI")
	require.NotNil(t, msg.abi, "ABI should not be nil")

	// Set the ABI on the contract for further testing
	p.Contracts[0].abi = msg.abi

	// Verify the ABI was correctly extracted
	events := p.Contracts[0].EventModels()
	require.Len(t, events, 1, "should have one event")
	assert.Equal(t, "Transfer", events[0].Proto.MessageName, "event name should be Transfer")
}

func TestWrappedABIFormatWithStringInput(t *testing.T) {
	conv := New()
	conv.SetFactory(&codegen.MsgWrapFactory{})
	p := conv.(*Convo).State

	// Setup
	p.Name = "test-project"
	p.ChainName = "mainnet"
	p.Contracts = append(p.Contracts, &Contract{
		Address: "0x1231231230123123123012312312301231231230",
	})
	p.currentContractIdx = 0
	p.Contracts[0].abiType = "string"

	// Test with wrapped ABI format via string input
	wrappedABI := `{"abi": [{"type": "function", "name": "transfer", "inputs": [], "outputs": [], "stateMutability": "nonpayable"}]}`

	next := conv.Update(InputContractABIString{
		UserInput_TextInput: pbconvo.UserInput_TextInput{
			Value: wrappedABI,
		},
	})

	// Should proceed to next step (not error out)
	assert.Equal(t, RunDecodeContractABI{}, next())
}

func TestInvalidWrappedABIFormat(t *testing.T) {
	conv := New()
	conv.SetFactory(&codegen.MsgWrapFactory{})
	p := conv.(*Convo).State

	// Setup
	p.Name = "test-project"
	p.ChainName = "mainnet"
	p.Contracts = append(p.Contracts, &Contract{
		Address: "0x1231231230123123123012312312301231231230",
	})
	p.currentContractIdx = 0

	// Test with object that has no "abi" field
	invalidWrapped := `{"contractName": "Test", "bytecode": "0x..."}`

	next := conv.Update(InputContractABIFile{
		UserInput_LocalFile: pbconvo.UserInput_LocalFile{
			Value: []byte(invalidWrapped),
		},
	})

	// Should show error and ask for ABI again
	seq := next().(loop.SeqMsg)
	msg := seq[0]().(*pbconvo.SystemOutput)
	assert.Contains(t, msg.GetMessage().Markdown, "abi")
}

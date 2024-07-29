package ethfull

import (
	"fmt"
	"strings"
	"testing"

	"github.com/streamingfast/eth-go"
	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
	pbconvo "github.com/streamingfast/substreams-codegen/pb/sf/codegen/conversation/v1"
	pbbuild "github.com/streamingfast/substreams-codegen/pb/sf/codegen/remotebuild/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvoNextStep(t *testing.T) {
	p := &Project{}
	next := func() loop.Msg {
		return p.NextStep()()
	}

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
	f := &codegen.MsgWrapFactory{}
	conv := NewWithSQL(f).(*Convo)
	p := conv.state

	next := conv.Update(codegen.InputProjectName{pbconvo.UserInput_TextInput{
		Value: "my-proj",
	}})
	assert.Equal(t, "my-proj", p.Name)

	assert.Equal(t, codegen.AskChainName{}, next())
	next = conv.Update(codegen.InputChainName{pbconvo.UserInput_Selection{
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
	assert.Equal(t, AskContractABI{}, seq[1]())

	next = conv.Update(InputContractABI{UserInput_TextInput: pbconvo.UserInput_TextInput{Value: "[]"}})
	assert.Equal(t, RunDecodeContractABI{}, next())

	next = conv.Update(RunDecodeContractABI{})
	msg, ok := next().(ReturnRunDecodeContractABI)
	require.True(t, ok)
	assert.Nil(t, msg.err)

	next = conv.Update(msg)
	// TODO: test the output with the given abi methods in there...
	assert.Equal(t, FetchContractInitialBlock{}, next())

	next = conv.Update(ReturnFetchContractInitialBlock{Err: fmt.Errorf("failed")})
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

	assert.Equal(t, AskAddContract{}, next())

	next = conv.Update(InputAddContract{UserInput_Confirmation: pbconvo.UserInput_Confirmation{Affirmative: false}})
	assert.Equal(t, codegen.AskSqlOutputFlavor{}, next())

	next = conv.Update(codegen.InputSQLOutputFlavor{UserInput_Selection: pbconvo.UserInput_Selection{Value: "sql"}})
	assert.Equal(t, codegen.RunGenerate{}, next())

	next = conv.Update(codegen.ReturnGenerate{ProjectFiles: nil})
	seq = next().(loop.SeqMsg)

	//cmds := next()
	msg1 := seq[0]().(*pbconvo.SystemOutput)
	assert.Equal(t, msg1.GetMessage().Markdown, "Code generation complete!")
	msg2 := seq[1]().(*pbconvo.SystemOutput)
	assert.NotNil(t, msg2.GetDownloadFiles())

	next = conv.Update(codegen.InputSourceDownloaded{})
	assert.Equal(t, codegen.AskConfirmCompile{}, next())

	next = conv.Update(codegen.InputConfirmCompile{UserInput_Confirmation: pbconvo.UserInput_Confirmation{Affirmative: true}})
	assert.Equal(t, codegen.RunBuild{}, next())

	next = conv.Update(codegen.RunBuild{})
	assert.IsType(t, codegen.CompilingBuild{}, next())

	respCh := make(chan *codegen.RemoteBuildState)
	go func() {
		respCh <- &codegen.RemoteBuildState{
			Logs: []string{"building"},
		}
	}()

	next = conv.Update(codegen.CompilingBuild{RemoteBuildChan: respCh, FirstTime: true})

	seq = next().(loop.SeqMsg)
	msg1 = seq[0]().(*pbconvo.SystemOutput)
	assert.True(t, strings.HasPrefix(msg1.GetLoading().Label, "Compiling your Substreams"))
	assert.IsType(t, codegen.CompilingBuild{}, seq[1]())

	go func() {
		respCh <- &codegen.RemoteBuildState{
			Logs: []string{"Done"},
			Artifacts: []*pbbuild.BuildResponse_BuildArtifact{
				{
					Filename: "test.spkg",
					Content:  []byte("test"),
				},
			},
		}
	}()

	next = conv.Update(codegen.CompilingBuild{RemoteBuildChan: respCh, FirstTime: false})

	seq = next().(loop.SeqMsg)
	assert.Len(t, seq, 2)

	assert.IsType(t, &pbconvo.SystemOutput{}, seq[0]())
	assert.IsType(t, codegen.ReturnBuild{}, seq[1]())

	next = conv.Update(codegen.ReturnBuild{
		Err: fmt.Errorf("failed"),
	})

	seq = next().(loop.SeqMsg)
	message := seq[0]().(*pbconvo.SystemOutput).GetMessage().Markdown
	assert.True(t, strings.Contains(message, "failed"))

	next = conv.Update(codegen.PackageDownloaded{})
	assert.IsType(t, loop.QuitMsg{}, next())
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
	f := &codegen.MsgWrapFactory{}
	conv := NewWithSQL(f).(*Convo)
	p := conv.state
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
	f := &codegen.MsgWrapFactory{}
	conv := NewWithSQL(f).(*Convo)
	p := conv.state
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
	f := &codegen.MsgWrapFactory{}
	conv := NewWithSQL(f).(*Convo)
	p := conv.state
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
	f := &codegen.MsgWrapFactory{}
	conv := NewWithSQL(f).(*Convo)
	p := conv.state
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

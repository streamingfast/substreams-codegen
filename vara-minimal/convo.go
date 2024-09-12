package varaminimal

import (
	"encoding/json"
	"fmt"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
)

type Convo struct {
	*codegen.Conversation[*Project]
}

func New() codegen.Converser {
	return &Convo{&codegen.Conversation[*Project]{
		State: &Project{},
	}}
}
func init() {
	codegen.RegisterConversation(
		"vara-minimal",
		"Simplest Substreams to get you started on Vara Mainnet",
		"This creating the most simple substreams on Vara Mainnet",
		codegen.ConversationFactory(New),
		41,
	)
}

func (c *Convo) NextStep() loop.Cmd {
	p := c.State
	if p.Name == "" {
		return cmd(codegen.AskProjectName{})
	}

	if p.ChainName == "" {
		return cmd(codegen.AskChainName{})
	}

	if !p.IsValidChainName(p.ChainName) {
		return loop.Seq(cmd(codegen.MsgInvalidChainName{}), cmd(codegen.AskChainName{}))
	}

	return cmd(codegen.RunGenerate{})
}

func (c *Convo) Update(msg loop.Msg) loop.Cmd {
	switch msg := msg.(type) {
	case codegen.MsgStart:
		var msgCmd loop.Cmd
		if msg.Hydrate != nil {
			if err := json.Unmarshal([]byte(msg.Hydrate.SavedState), &c.State); err != nil {
				return loop.Quit(fmt.Errorf(`something went wrong, here's an error message to share with our devs (%s); we've notified them already`, err))
			}

			msgCmd = c.Msg().Message("Ok, I reloaded your state.").Cmd()
		} else {
			msgCmd = c.Msg().Message("Ok, let's start a new package.").Cmd()
		}
		return loop.Seq(msgCmd, c.NextStep())

	case codegen.AskProjectName:
		return c.CmdAskProjectName()

	case codegen.InputProjectName:
		c.State.Name = msg.Value
		return c.NextStep()

	case codegen.AskChainName:
		var labels, values []string
		for _, conf := range ChainConfigs {
			labels = append(labels, conf.DisplayName)
			values = append(values, conf.ID)
		}
		return c.Action(codegen.InputChainName{}).ListSelect("Please select the chain").
			Labels(labels...).
			Values(values...).
			Cmd()

	case codegen.MsgInvalidChainName:
		return c.Msg().
			Messagef(`Hmm, %q seems like an invalid chain name. Maybe it was supported and is not anymore?`, c.State.ChainName).
			Cmd()

	case codegen.InputChainName:
		c.State.ChainName = msg.Value
		if c.State.IsValidChainName(msg.Value) {
			return loop.Seq(
				c.Msg().Messagef("Got it, will be using chain %q", c.State.ChainConfig().DisplayName).Cmd(),
				c.NextStep(),
			)
		}
		return c.NextStep()

	case codegen.RunGenerate:
		return c.CmdGenerate(c.State.Generate)

	case codegen.ReturnGenerate:
		return c.CmdDownloadFiles(msg)
	}

	return loop.Quit(fmt.Errorf("invalid loop message: %T", msg))
}

var cmd = codegen.Cmd

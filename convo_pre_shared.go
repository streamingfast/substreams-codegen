package codegen

import (
	"encoding/json"
	"fmt"
	"strings"

	registry "github.com/pinax-network/graph-networks-libs/packages/golang/lib"
	networks "github.com/streamingfast/firehose-networks"
	"github.com/streamingfast/substreams-codegen/loop"
)

type SharedFlowConfig struct {
	// If ValidChains is non-empty, it indicates that asking for the chain name should be asked
	// to the user.
	ValidChains []*registry.Network
}

func (c *Conversation[X]) IsPreSharedFlowMsg(msg loop.Msg, config SharedFlowConfig) bool {
	switch msg.(type) {
	case MsgStart, AskProjectName, InputProjectName:
		return true
	case MsgInvalidChainName, AskChainName, InputChainName:
		if len(config.ValidChains) > 0 {
			return true
		}

		return false
	default:
		return false
	}
}

// IsPreSharedFlowDone returns true if all pre-shared flow steps have been completed for this conversation.
// In your NextStep implementation, you should always check this method.
//
//	func (c *Convo) NextStep() (out loop.Cmd) {
//		  if !c.IsPreSharedFlowDone() {
//		    return c.NextPreSharedFlowStep()
//		  }
//	     ...
//		  return c.NextPostSharedFlowStep()
//		   }
func (c *Conversation[X]) IsPreSharedFlowDone(config SharedFlowConfig) bool {
	if c.State.GetModuleName() == "" {
		return false
	}

	if len(config.ValidChains) > 0 {
		if !c.isValidChainInput(config) {
			return false
		}

		return c.State.GetChainName() != ""
	}

	return true
}

// To be called if [IsPreSharedFlowDone] returns true, this method will
// route to the appropriate post-shared flow step based on the message received.
//
//	func (c *Convo) NextStep() (out loop.Cmd) {
//	  if !c.IsPreSharedFlowDone() {
//		 return c.NextPreSharedFlowStep()
//	  }
//
//	  ...
//
//	  return c.NextPostSharedFlowStep()
//	 }
func (c *Conversation[X]) NextPreSharedFlowStep(config SharedFlowConfig) loop.Cmd {
	if c.State.GetModuleName() == "" {
		return Cmd(AskProjectName{})
	}

	if len(config.ValidChains) == 0 {
		panic(fmt.Errorf("pre-shared flow is done, no next step available, you shouldn't have call NextPreSharedFlowStep"))
	}

	if c.State.GetChainName() == "" {
		return Cmd(AskChainName{})
	}

	if !c.isValidChainInput(config) {
		return loop.Seq(Cmd(MsgInvalidChainName{}), Cmd(AskChainName{}))
	}

	panic(fmt.Errorf("pre-shared flow is done, no next step available, you shouldn't have call NextPreSharedFlowStep"))
}

func (c *Conversation[X]) UpdatePreSharedFlowMsg(msg loop.Msg, config SharedFlowConfig, nextStep func() loop.Cmd) loop.Cmd {
	switch msg := msg.(type) {
	case MsgStart:
		c.SetClientVersion(msg.Version)
		var msgCmd loop.Cmd
		if msg.Hydrate != nil {
			if err := json.Unmarshal([]byte(msg.Hydrate.SavedState), &c.State); err != nil {
				return loop.Quit(fmt.Errorf(`something went wrong, here's an error message to share with our devs (%s); we've notified them already`, err))
			}

			if v, ok := any(c.State).(ValidatingConversationState); ok {
				if err := v.Validate(); err != nil {
					return loop.Quit(fmt.Errorf(`something went wrong, the initial state is invalid: %w`, err))
				}
			}

			msgCmd = c.Msg().Message("Ok, I reloaded your state.").Cmd()
		} else {
			msgCmd = c.Msg().Message("Ok, let's start a new package.").Cmd()
		}
		return loop.Seq(msgCmd, nextStep())

	case AskProjectName:
		return c.CmdAskProjectName()

	case InputProjectName:
		c.State.SetProjectName(msg.Value)
		return nextStep()

	case AskChainName:
		if len(config.ValidChains) == 0 {
			panic("AskChainName called but no valid chain names provided in SharedFlowConfig, this shouldn't happen")
		}

		var labels, values []string
		for _, network := range config.ValidChains {
			labels = append(labels, network.FullName)
			values = append(values, network.ID)
		}
		return c.Action(InputChainName{}).ListSelect("Please select the chain", "chain").
			Labels(labels...).
			Values(values...).
			Cmd()

	case MsgInvalidChainName:
		actual := c.State.GetChainName()
		var validChainNames []string
		for _, network := range config.ValidChains {
			validChainNames = append(validChainNames, `'`+network.ID+`'`)
		}

		return c.Msg().
			Messagef(`Hmm, %q seems like an invalid chain name. Maybe it was supported and is not anymore, current supported chains are %s`, actual, strings.Join(validChainNames, ", ")).
			Cmd()

	case InputChainName:
		c.State.SetChainName(msg.Value)
		if c.isValidChainInput(config) {
			return loop.Seq(
				c.Msg().Messagef("Got it, will be using chain %q", c.State.GetChainDisplayName()).Cmd(),
				nextStep(),
			)
		}
		return nextStep()
	}

	panic(fmt.Errorf("pre-shared flow message not handled, have you validated it was a pre-shared flow message before hand?: %T", msg))
}

func (c *Conversation[X]) CmdAskProjectName() loop.Cmd {
	return c.Action(InputProjectName{}).
		TextInput("Please enter the project name", "Submit").
		Description("Identifier with only lowercase letters, numbers and underscores, up to 64 characters.").
		DefaultValue("my_project").
		Validation("^([a-z][a-z0-9_]{0,63})$", "The project name must be a valid identifier with only lowercase letters, numbers and underscores, up to 64 characters.").
		Cmd()
}

func (c *Conversation[X]) isValidChainInput(config SharedFlowConfig) bool {
	if len(config.ValidChains) == 0 {
		return true
	}

	registry := make(networks.NetworkRegistry)
	for _, network := range config.ValidChains {
		registry[network.ID] = network
	}

	return registry.Has(c.State.GetChainName())
}

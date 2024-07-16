package pbconvo

import (
	"fmt"
	"strings"
)

var nl = "\n    ┃ "

func wrapnl(s string) string {
	return strings.ReplaceAll(s, "\n", "\n    ┃ ")
}

func (msg *SystemOutput) Humanize(seconds int) string {
	time := fmt.Sprintf("%4d┃ ", seconds)

	switch {
	case msg.GetMessage() != nil:
		text := msg.GetMessage().GetMarkdown()
		return fmt.Sprintf("%s%s", time, text)
	case msg.GetImageWithText() != nil:
		text := msg.GetImageWithText().GetMarkdown()
		return fmt.Sprintf("%s[ image ] %s", time, text)
	case msg.GetListSelect() != nil:
		sel := msg.GetListSelect()
		vals := make([]string, 0, len(sel.Values))
		for i, v := range sel.Values {
			vals = append(vals, fmt.Sprintf("%s- %s (%s)", nl, sel.Labels[i], v))
		}
		return fmt.Sprintf("%s%s%s", time, wrapnl(sel.Instructions), strings.Join(vals, ""))
	case msg.GetTextInput() != nil:
		inp := msg.GetTextInput()
		return fmt.Sprintf("%s%s%s%s%s> %s", time, wrapnl(inp.Prompt), nl, wrapnl(inp.Description), nl, inp.DefaultValue)
	case msg.GetConfirm() != nil:
		conf := msg.GetConfirm()
		return fmt.Sprintf("%s%s%s%s%s[ %s / %s ]", time, wrapnl(conf.Prompt), nl, wrapnl(conf.Description), nl, conf.AcceptButtonLabel, conf.DeclineButtonLabel)
	case msg.GetDownloadFiles() != nil:
		filenames := make([]string, 0, len(msg.GetDownloadFiles().Files))
		for _, f := range msg.GetDownloadFiles().Files {
			filenames = append(filenames, f.Filename)
		}
		return fmt.Sprintf("%s[Downloading files]%s%s", time, nl+"- ", strings.Join(filenames, nl+"- "))
	case msg.GetLoading() != nil:
		loading := msg.GetLoading()
		return time + "Loading ..." + loading.Label
	}
	return fmt.Sprintf("%s, %v", msg.ActionId, msg.Entry)
}

func (i UserInput_Selection) Humanize(seconds int) string {
	time := fmt.Sprintf("%4d ", seconds)
	return fmt.Sprintf("%s[Selected] %s (%s)", time, i.Label, i.Value)
}

func (i UserInput_TextInput) Humanize(seconds int) string {
	time := fmt.Sprintf("%4d ", seconds)
	return fmt.Sprintf("%s%s", time, i.Value)
}

func (i UserInput_Confirmation) Humanize(seconds int) string {
	time := fmt.Sprintf("%4d ", seconds)
	return fmt.Sprintf("%s[Confirmed] %t", time, i.Affirmative)
}

func (i UserInput_DownloadedFiles) Humanize(seconds int) string {
	time := fmt.Sprintf("%4d ", seconds)
	return fmt.Sprintf("[Downloaded files] %s", time)
}

func (i UserInput_File) Humanize(seconds int) string {
	return fmt.Sprintf("< [Uploaded file]: %s", i.File.Filename)
}

// not used
//func (i UserInput_Start) Humanize() string {
//	return fmt.Sprintf("< [Start, hydrate: %t] %s", i.Hydrate != nil, i.GeneratorId)
//}

type Humanizable interface {
	Humanize(int) string
}

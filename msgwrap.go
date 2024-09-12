package codegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"reflect"

	"github.com/streamingfast/substreams-codegen/loop"
	pbconvo "github.com/streamingfast/substreams-codegen/pb/sf/codegen/conversation/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type MsgStart struct {
	pbconvo.UserInput_Start
}

type IncomingMessage struct {
	Msg any
}

func (m *IncomingMessage) Humanize(seconds int) string {
	if h, ok := m.Msg.(pbconvo.Humanizable); ok {
		return h.Humanize(seconds)
	}
	if m.Msg == nil {
		return ("---")
	}

	return fmt.Sprintf("%d | %T %v", seconds, m.Msg, m.Msg)
}

type MsgWrapFactory struct {
	sendFunc   SendFunc
	inputTypes map[string]reflect.Type
	lastType   reflect.Type

	loop.EventLoop
}

func NewMsgWrapFactory(sendFunc SendFunc) *MsgWrapFactory {
	f := &MsgWrapFactory{
		sendFunc:   sendFunc,
		inputTypes: make(map[string]reflect.Type),
	}
	return f
}

func (f *MsgWrapFactory) SetupLoop(updateFunc func(msg loop.Msg) loop.Cmd) {
	f.EventLoop = loop.NewEventLoop(updateFunc)
}

func (f *MsgWrapFactory) NewMsg(state any) *MsgWrap {
	w := &MsgWrap{}
	w.Msg = &pbconvo.SystemOutput{}
	if state != nil {
		cnt, err := json.Marshal(state)
		if err != nil {
			panic(err)
		}
		w.Msg.State = string(cnt)
	}
	return w
}

func (f *MsgWrapFactory) NewInput(inputMsg any, state any) *MsgWrap {
	msg := f.NewMsg(state)
	reflectType := reflect.TypeOf(inputMsg)
	f.lastType = reflectType
	el := reflect.New(reflectType).Interface()
	_, ok := el.(protoreflect.ProtoMessage)
	if !ok {
		panic("only use NewInput with messages that embed a return value of type pbconvo.UserInput_*")
	}
	return msg
}

func (f *MsgWrapFactory) LastInput() reflect.Type {
	return f.lastType
}

type MsgWrap struct {
	Msg *pbconvo.SystemOutput
	Err error
}

func (w *MsgWrap) Messagef(markdown string, args ...interface{}) *MsgWrap {
	w.Msg.Entry = &pbconvo.SystemOutput_Message_{
		Message: &pbconvo.SystemOutput_Message{Markdown: fmt.Sprintf(markdown, args...)},
	}
	return w
}

func (w *MsgWrap) Errorf(msg string, args ...any) *MsgWrap {
	w.Err = fmt.Errorf(msg, args...)
	return w
}

func (w *MsgWrap) Message(markdown string) *MsgWrap {
	w.Msg.Entry = &pbconvo.SystemOutput_Message_{
		Message: &pbconvo.SystemOutput_Message{Markdown: markdown},
	}
	return w
}

func (w *MsgWrap) Confirm(prompt string, acceptLabel, declineLabel string) *MsgWrap {
	// TODO: to a type assertion on the `lastType`, to make sure it matches what we're asking here..
	w.Msg.Entry = &pbconvo.SystemOutput_Confirm_{
		Confirm: &pbconvo.SystemOutput_Confirm{
			Prompt:             prompt,
			AcceptButtonLabel:  acceptLabel,
			DeclineButtonLabel: declineLabel,
		},
	}
	return w
}

func (w *MsgWrap) DownloadFiles() *MsgWrap {
	// TODO: to a type assertion on the `lastType`, to make sure it matches what we're asking here..
	w.Msg.Entry = &pbconvo.SystemOutput_DownloadFiles_{
		DownloadFiles: &pbconvo.SystemOutput_DownloadFiles{},
	}
	return w
}

func (w *MsgWrap) AddFile(filename string, cnt []byte, fileType string, description string) *MsgWrap {
	switch entry := w.Msg.Entry.(type) {
	case *pbconvo.SystemOutput_DownloadFiles_:
		input := entry.DownloadFiles

		input.Files = append(input.Files, &pbconvo.SystemOutput_DownloadFile{
			Filename:    filename,
			Content:     cnt,
			Type:        fileType,
			Description: description,
		})
	default:
		panic("unsupported message type for this method")
	}
	return w
}

func (w *MsgWrap) Loading(loading bool, label string) *MsgWrap {
	w.Msg.Entry = &pbconvo.SystemOutput_Loading_{
		Loading: &pbconvo.SystemOutput_Loading{
			Loading: loading,
			Label:   label,
		},
	}
	return w
}

func (w *MsgWrap) StopLoading() *MsgWrap {
	w.Msg.Entry = &pbconvo.SystemOutput_Loading_{
		Loading: &pbconvo.SystemOutput_Loading{
			Loading: false,
		},
	}
	return w
}

func (w *MsgWrap) Loadingf(loading bool, label string, args ...interface{}) *MsgWrap {
	w.Msg.Entry = &pbconvo.SystemOutput_Loading_{
		Loading: &pbconvo.SystemOutput_Loading{
			Loading: loading,
			Label:   fmt.Sprintf(label, args...),
		},
	}
	return w
}

func (w *MsgWrap) TextInput(prompt string, submitButtonLabel string) *MsgWrap {
	// TODO: to a type assertion on the `lastType`, to make sure it matches what we're asking here..
	w.Msg.Entry = &pbconvo.SystemOutput_TextInput_{
		TextInput: &pbconvo.SystemOutput_TextInput{
			Prompt:            prompt,
			SubmitButtonLabel: submitButtonLabel,
		},
	}
	return w
}

func (w *MsgWrap) DefaultValue(value string) *MsgWrap {
	switch entry := w.Msg.Entry.(type) {
	case *pbconvo.SystemOutput_TextInput_:
		entry.TextInput.DefaultValue = value
	default:
		panic("unsupported message type for this method")
	}
	return w
}
func (w *MsgWrap) Description(description string) *MsgWrap {
	switch entry := w.Msg.Entry.(type) {
	case *pbconvo.SystemOutput_TextInput_:
		entry.TextInput.Description = description
	case *pbconvo.SystemOutput_Confirm_:
		entry.Confirm.Description = description
	default:
		panic("unsupported message type for this method")
	}
	return w
}

func (w *MsgWrap) ListSelect(instructions string) *MsgWrap {
	// TODO: to a type assertion on the `lastType`, to make sure it matches what we're asking here..
	w.Msg.Entry = &pbconvo.SystemOutput_ListSelect_{
		ListSelect: &pbconvo.SystemOutput_ListSelect{
			Instructions: instructions,
		},
	}
	return w
}

func (w *MsgWrap) Labels(labels ...string) *MsgWrap {
	switch entry := w.Msg.Entry.(type) {
	case *pbconvo.SystemOutput_ListSelect_:
		entry.ListSelect.Labels = labels
		entry.ListSelect.Values = labels
	default:
		panic("unsupported message type for this method")
	}
	return w
}

func (w *MsgWrap) SelectButton(label string) *MsgWrap {
	switch entry := w.Msg.Entry.(type) {
	case *pbconvo.SystemOutput_ListSelect_:
		entry.ListSelect.SelectButtonLabel = label
	default:
		panic("unsupported message type for this method")
	}
	return w
}

func (w *MsgWrap) Values(values ...string) *MsgWrap {
	switch entry := w.Msg.Entry.(type) {
	case *pbconvo.SystemOutput_ListSelect_:
		entry.ListSelect.Values = values
	default:
		panic("unsupported message type for this method")
	}
	return w
}

func (w *MsgWrap) Placeholder(message string) *MsgWrap {
	switch entry := w.Msg.Entry.(type) {
	case *pbconvo.SystemOutput_TextInput_:
		entry.TextInput.Placeholder = message
	default:
		panic("unsupported message type for this method")
	}
	return w
}

func (w *MsgWrap) Multiline(val int) *MsgWrap {
	switch entry := w.Msg.Entry.(type) {
	case *pbconvo.SystemOutput_TextInput_:
		entry.TextInput.MultiLine = int32(val)
	default:
		panic("unsupported message type for this method")
	}
	return w
}

func (w *MsgWrap) Validation(regexp string, errorMessage string) *MsgWrap {
	switch entry := w.Msg.Entry.(type) {
	case *pbconvo.SystemOutput_TextInput_:
		entry.TextInput.ValidationRegexp = regexp
		entry.TextInput.ValidationErrorMessage = errorMessage
	default:
		panic("unsupported message type for this method")
	}
	return w
}

func tplMe(templateText string, data interface{}) string {
	tpl, err := template.New("tpl").Parse(templateText)
	if err != nil {
		panic(fmt.Errorf("error parsing template: %w", err))
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		panic(fmt.Errorf("error executing template: %w", err))
	}
	return buf.String()
}

func (w *MsgWrap) MessageTpl(templateText string, data interface{}) *MsgWrap {
	w.Msg.Entry = &pbconvo.SystemOutput_Message_{
		Message: &pbconvo.SystemOutput_Message{Markdown: tplMe(templateText, data)},
	}
	return w
}

func (w *MsgWrap) Style(style string) *MsgWrap {
	// for example, "error", "warning", etc..
	switch entry := w.Msg.Entry.(type) {
	case *pbconvo.SystemOutput_Message_:
		entry.Message.Style = style
	default:
		panic("unsupported message type for this method")
	}
	return w
}

// This will wait for an answer
func (w *MsgWrap) Cmd() loop.Cmd {
	// Make sure this is called only on those that EXPECT a return value
	return func() loop.Msg {
		return w.Msg
	}
}

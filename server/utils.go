package server

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/streamingfast/dstore"
)

func NewErrConversationTimeout(timeout time.Duration) error {
	return ErrConversationTimeout{Timeout: timeout}
}

type ErrConversationTimeout struct {
	Timeout time.Duration
}

func (e ErrConversationTimeout) Error() string {
	return fmt.Sprintf("conversation timed out after %s", e.Timeout)
}

type SessionLogger interface {
	SaveSession(codegen string, events []string, lastState string) error
}

type StoreSessionLogger struct {
	store dstore.Store
}

func (s StoreSessionLogger) SaveSession(codegen string, events []string, lastState string) error {
	filename := fmt.Sprintf("session-%s-%s.log", codegen, time.Now().Format(time.RFC3339))

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	r := bufio.NewReader(&buf)
	go func() {
		for _, event := range events {
			_, err := fmt.Fprintln(w, event)
			if err != nil {
				fmt.Println("error writing to buffer", err)
				return
			}
		}
		fmt.Fprintln(w, "\nJSON state:")
		fmt.Fprintln(w, lastState)
		w.Flush()
	}()
	return s.store.WriteObject(context.TODO(), filename, r)
}

type PrintSessionLogger struct{}

func (p PrintSessionLogger) SaveSession(codegen string, events []string, lastState string) error {
	_ = lastState
	fmt.Println("Session log for codegen", codegen)
	for _, event := range events {
		println(event)
	}
	return nil
}

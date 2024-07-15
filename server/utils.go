package server

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/streamingfast/dstore"
	"go.uber.org/zap"
)

func (s *server) writeErrorWithMsg(w http.ResponseWriter, statusCode int, message string, err error) {
	if err != nil {
		s.logger.Error(message, zap.Error(err))
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}

func (s *server) writeError(w http.ResponseWriter, statusCode int, err error) {
	if err != nil {
		s.logger.Error("error encountered while serving request", zap.Error(err))
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(err.Error()))
}

type SessionLogger interface {
	SaveSession(codegen string, events []string) error
}

type StoreSessionLogger struct {
	store dstore.Store
}

func (s StoreSessionLogger) SaveSession(codegen string, events []string) error {
	filename := fmt.Sprintf("session-%s-%d.log", codegen, time.Now().Unix())

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
		w.Flush()
	}()
	return s.store.WriteObject(context.TODO(), filename, r)
}

type PrintSessionLogger struct{}

func (p PrintSessionLogger) SaveSession(codegen string, events []string) error {
	fmt.Println("Session log for codegen", codegen)
	for _, event := range events {
		println(event)
	}
	return nil
}

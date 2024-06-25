package server

import (
	"go.uber.org/zap"
	"net/http"
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

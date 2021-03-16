package util

import (
	"fmt"
)

const (
	NoneLog  = "none"
	DebugLog = "debug"
	InfoLog  = "info"
)

// Logger interface
type Logger interface {
	Debug(string)
	Info(string)
	StartProgress(string)
	StopProgress(string)
}

// NewStdoutLogger create new logger
func NewStdoutLogger(level string) Logger {
	return &StdoutLogger{
		level: level,
	}
}

// StdoutLogger stdout logging implementation
type StdoutLogger struct {
	level   string
	spinner *Spinner
}

func (s *StdoutLogger) log(msg string, level string) {
	if s.level == NoneLog || s.level == DebugLog && s.level != level {
		return
	}

	fmt.Println(msg)
}

// Info log
func (s *StdoutLogger) Info(msg string) {
	s.log(msg, InfoLog)
}

// Debug log
func (s *StdoutLogger) Debug(msg string) {
	s.log(msg, DebugLog)
}

func (s *StdoutLogger) StartProgress(msg string) {
	if s.level == NoneLog {
		return
	}

	s.spinner = NewSpinner(msg, "")
	s.spinner.Start()
}

func (s *StdoutLogger) StopProgress(msg string) {
	if s.level == NoneLog {
		return
	}

	s.spinner.Stop(msg)
}

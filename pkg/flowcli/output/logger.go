/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package output

import (
	"fmt"

	"github.com/onflow/flow-cli/pkg/flowcli/util"
)

const (
	NoneLog  = 0
	ErrorLog = 1
	DebugLog = 2
	InfoLog  = 3
)

// Logger interface
type Logger interface {
	Debug(string)
	Info(string)
	Error(string)
	StartProgress(string)
	StopProgress(string)
}

// NewStdoutLogger create new logger
func NewStdoutLogger(level int) *StdoutLogger {
	return &StdoutLogger{
		level: level,
	}
}

// StdoutLogger stdout logging implementation
type StdoutLogger struct {
	level   int
	spinner *Spinner
}

func (s *StdoutLogger) log(msg string, level int) {
	if s.level < level {
		return
	}

	fmt.Printf("%s\n", msg)
}

// Info log
func (s *StdoutLogger) Info(msg string) {
	s.log(msg, InfoLog)
}

// Debug log
func (s *StdoutLogger) Debug(msg string) {
	s.log(msg, DebugLog)
}

// Error log
func (s *StdoutLogger) Error(msg string) {
	s.log(fmt.Sprintf("❌  %s", util.Red(msg)), ErrorLog)
}

func (s *StdoutLogger) StartProgress(msg string) {
	if s.level == NoneLog {
		return
	}

	if s.spinner != nil {
		s.spinner.Stop("")
	}

	s.spinner = NewSpinner(msg, "")
	s.spinner.Start()
}

func (s *StdoutLogger) StopProgress(msg string) {
	if s.level == NoneLog {
		return
	}

	if s.spinner != nil {
		s.spinner.Stop(msg)
		s.spinner = nil
	}
}

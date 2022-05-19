/*
 * Flow CLI
 *
 * Copyright 2019 Dapper Labs, Inc.
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
)

const (
	NoneLog  = 0
	ErrorLog = 1
	DebugLog = 2
	InfoLog  = 3
)

type Logger interface {
	Debug(string)
	Info(string)
	Error(string)
	StartProgress(string)
	StopProgress()
}

// NewStdoutLogger returns a new stdout logger.
func NewStdoutLogger(level int) *StdoutLogger {
	return &StdoutLogger{
		level: level,
	}
}

// StdoutLogger is a stdout logging implementation.
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

func (s *StdoutLogger) Info(msg string) {
	s.log(msg, InfoLog)
}

func (s *StdoutLogger) Debug(msg string) {
	s.log(msg, DebugLog)
}

func (s *StdoutLogger) Error(msg string) {
	s.log(fmt.Sprintf("%s %s", ErrorEmoji(), Red(msg)), ErrorLog)
}

func (s *StdoutLogger) StartProgress(msg string) {
	if s.level == NoneLog {
		return
	}

	if s.spinner != nil {
		s.spinner.Stop()
	}

	s.spinner = NewSpinner(msg, "")
	s.spinner.Start()
}

func (s *StdoutLogger) StopProgress() {
	if s.level == NoneLog {
		return
	}

	if s.spinner != nil {
		s.spinner.Stop()
		s.spinner = nil
	}
}

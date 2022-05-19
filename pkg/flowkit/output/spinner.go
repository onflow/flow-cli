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
	"time"

	"github.com/gosuri/uilive"
)

var spinnerCharset = []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}

type Spinner struct {
	prefix string
	suffix string
	done   chan string
}

func NewSpinner(prefix, suffix string) *Spinner {
	return &Spinner{
		prefix: prefix,
		suffix: suffix,
		done:   make(chan string),
	}
}

func (s *Spinner) Start() {
	go s.run()
}

func (s *Spinner) run() {
	writer := uilive.New()

	ticker := time.NewTicker(100 * time.Millisecond)

	i := 0

	for {
		select {
		case <-s.done:
			_, _ = fmt.Fprintf(writer, "\r")
			_ = writer.Flush()
			close(s.done)
			return
		case <-ticker.C:
			_, _ = fmt.Fprintf(
				writer,
				"%s%c%s\n",
				s.prefix,
				spinnerCharset[i%len(spinnerCharset)],
				s.suffix,
			)
			_ = writer.Flush()
			i++
		}
	}
}

func (s *Spinner) Stop() {
	s.done <- ""
	time.Sleep(50 * time.Millisecond)
}

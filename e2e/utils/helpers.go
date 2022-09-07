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

package e2e

import (
	"bytes"
	"os/exec"
	"testing"
)

const PROMPT_UI_BUFFER_SIZE = 4096
const ASCII_A byte = 97

// Appends a padded prompt response to a buffer that can be passed
// to stdin for the CLI to read
//
// Padding is required because promptui reads prompt inputs with
// 4096 length buffers so we pad the rest of the  bytes with
// some value after our input or else PromptUI will throw us
// "unexpected EOF encountered" error
func RespondToPrompt(b []byte, buf *bytes.Buffer) {
	buf.Write(b)
	siz := len(b)

	pu := make([]byte, PROMPT_UI_BUFFER_SIZE-siz)
	for i := 0; i < PROMPT_UI_BUFFER_SIZE-siz; i++ {
		pu[i] = ASCII_A // some arbitrary character for padding
	}
	buf.Write(pu)
}

// start emulator without waiting for command to finish
// returns a cleanup function to kill emulator
func StartEmulator(t *testing.T) func() {
	emulatorCmd := exec.Command("./flow", "emulator")
	if err := emulatorCmd.Start(); err != nil {
		t.Error("Failed to start emulator: ", err)
	}
	return func() {
		if err := emulatorCmd.Process.Kill(); err != nil {
			t.Error("Failed to kill emulator: ", err)
		}
	}
}

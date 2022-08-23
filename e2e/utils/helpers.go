package e2e

import (
	"bytes"
	"os/exec"
	"testing"
)

const PROMPT_UI_BUFFER_SIZE = 4096
const ASCII_A byte = 97

// PromptUI (which is what makes the cli interactive)reads prompt
// inputs with 4096 length buffers so we e2e.Pad the rest of the bytes
// with some value after our input or else PromptUI will throw
// us "unexpected EOF encountered" error
func RespondToPrompt(b []byte, buf *bytes.Buffer) {
	buf.Write(b)
	siz := len(b)

	pu := make([]byte, PROMPT_UI_BUFFER_SIZE-siz)
	for i := 0; i < PROMPT_UI_BUFFER_SIZE-siz; i++ {
		pu[i] = ASCII_A // some arbitrary character for padding, in this case ascii 'a'
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

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
	"net"
	"os"
	"regexp"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/icmd"

	e2e "github.com/onflow/flow-cli/e2e/utils"
)

const FLOW_BINARY = "./flow"

func Test_cliCommands(t *testing.T) {
	result := icmd.RunCmd(icmd.Command(FLOW_BINARY, "init"))
	result.Assert(t, icmd.Success)
	defer os.Remove("./flow.json")

	// start emulator
	stopEmulator := e2e.StartEmulator(t)
	defer stopEmulator()

	// wait 5 seconds for emulator to start
	timeout := 5 * time.Second
	emulatorNotStarted := true
	start := time.Now()
	var elapsed time.Duration = 0
	for emulatorNotStarted && elapsed.Seconds() <= 30 {
		conn, _ := net.DialTimeout("tcp", "127.0.0.1:8888", timeout)
		if conn != nil {
			emulatorNotStarted = false
		}
		cur := time.Now()
		elapsed = cur.Sub(start)
	}

	// // ### ACCOUNTS AND SCRIPTS ###
	cmd := icmd.Command(FLOW_BINARY, "accounts", "create")

	buf := bytes.Buffer{}

	accountName := "remy\n"
	e2e.RespondToPrompt([]byte(accountName), &buf) // first prompt: account name
	e2e.RespondToPrompt([]byte{10}, &buf)          // second prompt: enter key
	e2e.RespondToPrompt([]byte{121, 10}, &buf)     // third prompt: "y" for yes and enter key
	cmd.Stdin = &buf

	result = icmd.RunCmd(cmd)
	result.Assert(t, icmd.Success)

	result = icmd.RunCommand(FLOW_BINARY, "accounts", "add-contract", "HelloWorld", "./files/contract.cdc")
	result.Assert(t, icmd.Success)

	result = icmd.RunCommand(FLOW_BINARY, "scripts", "execute", "./files/script.cdc")
	result.Assert(t, icmd.Success)
	expected := "\nResult: \"Hello world!\"\n\n\n"
	assert.Equal(t, result.Stdout(), expected, "Outputs of updated contract should be the same")

	result = icmd.RunCommand(FLOW_BINARY, "accounts", "update-contract", "HelloWorld", "./files/contractUpdated.cdc")
	result.Assert(t, icmd.Success)

	result = icmd.RunCommand(FLOW_BINARY, "scripts", "execute", "./files/script.cdc")
	result.Assert(t, icmd.Success)
	expected = "\nResult: \"Bonjour world!\"\n\n\n"
	assert.Equal(t, result.Stdout(), expected, "Outputs of updated contract should be the same")

	// ### TRANSACTIONS ### //
	result = icmd.RunCmd(icmd.Command(FLOW_BINARY, "transactions", "send", "./files/tx.cdc"))
	result.Assert(t, icmd.Success)

	cmd = icmd.Command(FLOW_BINARY, "transactions", "build", "./files/tx.cdc", "--filter", "payload", "--save", "./built.rlp")
	defer icmd.RunCommand("rm", "built.rlp")
	txBuf := bytes.Buffer{}
	e2e.RespondToPrompt([]byte{14, 10}, &txBuf) // down arrow key and enter
	cmd.Stdin = &txBuf
	result = icmd.RunCmd(cmd)
	result.Assert(t, icmd.Success)

	cmd = icmd.Command(FLOW_BINARY, "transactions", "sign", "./built.rlp", "--filter", "payload", "--save", "./signed.rlp")
	defer icmd.RunCommand("rm", "signed.rlp")
	txBuf.Reset()
	e2e.RespondToPrompt([]byte{14, 10}, &txBuf)
	cmd.Stdin = &txBuf
	result = icmd.RunCmd(cmd)
	result.Assert(t, icmd.Success)

	cmd = icmd.Command(FLOW_BINARY, "transactions", "send-signed", "./signed.rlp")
	txBuf.Reset()
	e2e.RespondToPrompt([]byte{14, 10}, &txBuf)
	cmd.Stdin = &txBuf
	result = icmd.RunCmd(cmd)
	result.Assert(t, icmd.Success)
	re := regexp.MustCompile("\nID\t(.*?)\n")
	matches := re.FindStringSubmatch(result.Stdout())
	txId := matches[1]

	result = icmd.RunCmd(icmd.Command(FLOW_BINARY, "transactions", "get", txId))
	result.Assert(t, icmd.Success)

	result = icmd.RunCmd(icmd.Command(FLOW_BINARY, "transactions", "decode", "./signed.rlp"))
	result.Assert(t, icmd.Success)

	result = icmd.RunCommand(FLOW_BINARY, "accounts", "remove-contract", "HelloWorld")
	result.Assert(t, icmd.Success)

	// ### BLOCKS ###
	result = icmd.RunCommand(FLOW_BINARY, "blocks", "get", "latest")
	result.Assert(t, icmd.Success)
	re = regexp.MustCompile("\nBlock ID\t\t(.*?)\n")
	matches = re.FindStringSubmatch(result.Stdout())
	blockId := matches[1]

	result = icmd.RunCommand(FLOW_BINARY, "blocks", "get", blockId)
	result.Assert(t, icmd.Success)

	result = icmd.RunCommand(FLOW_BINARY, "blocks", "get", "2")
	result.Assert(t, icmd.Success)

	// ### NETWORK ###
	result = icmd.RunCommand(FLOW_BINARY, "status")
	result.Assert(t, icmd.Success)

	// ### KEYS ###
	result = icmd.RunCmd(icmd.Command(FLOW_BINARY, "keys", "generate"))
	result.Assert(t, icmd.Success)
	re = regexp.MustCompile("\nPrivate Key \t (.*?) \n")
	matches = re.FindStringSubmatch(result.Stdout())
	privateKey := matches[1]

	// ### CONFIG ###
	cmd = icmd.Command(FLOW_BINARY, "config", "add", "account")
	confBuf := bytes.Buffer{}
	e2e.RespondToPrompt([]byte("joe\n"), &confBuf)
	e2e.RespondToPrompt([]byte("01cf0e2f2f715450\n"), &confBuf)
	e2e.RespondToPrompt([]byte{10}, &confBuf)
	e2e.RespondToPrompt([]byte{10}, &confBuf)
	e2e.RespondToPrompt([]byte(privateKey+"\n"), &confBuf)
	e2e.RespondToPrompt([]byte("0\n"), &confBuf)
	cmd.Stdin = &confBuf
	result = icmd.RunCmd(cmd)
	result.Assert(t, icmd.Success)
}

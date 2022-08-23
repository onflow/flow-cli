package e2e

import (
	"bytes"
	"os"
	"testing"

	e2e "github.com/onflow/flow-cli/e2e/utils"
	"gotest.tools/assert"
	"gotest.tools/v3/icmd"
)

func Test_keys(t *testing.T) {
	result := icmd.RunCmd(icmd.Command("./flow", "init"))
	result.Assert(t, icmd.Success)
	defer os.Remove("./flow.json")

	// start emulator
	// note: IS A race condition
	stopEmulator := e2e.StartEmulator(t)
	defer stopEmulator()

	// ### KEYS ###
	result = icmd.RunCmd(icmd.Command("./flow", "keys", "generate"))
	result.Assert(t, icmd.Success)

	// TODO: decode

	// // ### ACCOUNTS AND SCRIPTS ###
	cmd := icmd.Command("./flow", "accounts", "create")

	buf := bytes.Buffer{}

	// account name input text
	accountName := "gamer"
	acctInput := accountName + "\n"
	buf.WriteString(acctInput)
	e2e.Pad(len(acctInput), &buf)

	// second input: enter key
	buf.Write([]byte{10})
	e2e.Pad(1, &buf)

	// third input "y" for yes and enter key
	buf.Write([]byte{121, 10})
	e2e.Pad(2, &buf)

	cmd.Stdin = &buf

	result = icmd.RunCmd(cmd)
	result.Assert(t, icmd.Success)

	result = icmd.RunCommand("./flow", "accounts", "add-contract", "HelloWorld", "./files/contract.cdc")
	result.Assert(t, icmd.Success)

	result = icmd.RunCommand("./flow", "scripts", "execute", "./files/script.cdc")
	result.Assert(t, icmd.Success)
	expected := "\nResult: \"Hello world!\"\n\n\n"
	assert.Equal(t, result.Stdout(), expected, "Outputs of updated contract should be the same")

	result = icmd.RunCommand("./flow", "accounts", "update-contract", "HelloWorld", "./files/contractUpdated.cdc")
	result.Assert(t, icmd.Success)

	result = icmd.RunCommand("./flow", "scripts", "execute", "./files/script.cdc")
	result.Assert(t, icmd.Success)
	expected = "\nResult: \"Bonjour world!\"\n\n\n"
	assert.Equal(t, result.Stdout(), expected, "Outputs of updated contract should be the same")

	result = icmd.RunCommand("./flow", "accounts", "remove-contract", "HelloWorld")
	result.Assert(t, icmd.Success)

	// ### TRANSACTIONS ###

	cmd = icmd.Command("./flow", "transactions", "build", "./files/tx.cdc", "--filter", "payload", "--save", "./built.rlp")
	defer icmd.RunCommand("rm", "built.rlp")

	txBuf := bytes.Buffer{}

	txBuf.Write([]byte{14, 10})
	e2e.Pad(2, &txBuf)

	cmd.Stdin = &txBuf

	result = icmd.RunCmd(cmd)
	result.Assert(t, icmd.Success)

	// sign
	cmd = icmd.Command("./flow", "transactions", "sign", "./built.rlp", "--filter", "payload", "--save", "./signed.rlp")
	defer icmd.RunCommand("rm", "signed.rlp")

	txBuf.Reset()

	txBuf.Write([]byte{14, 10})
	e2e.Pad(2, &txBuf)

	cmd.Stdin = &txBuf

	result = icmd.RunCmd(cmd)
	result.Assert(t, icmd.Success)

	// send
	cmd = icmd.Command("./flow", "transactions", "send-signed", "./signed.rlp")

	txBuf.Reset()

	txBuf.Write([]byte{14, 10})
	e2e.Pad(2, &txBuf)

	cmd.Stdin = &txBuf

	result = icmd.RunCmd(cmd)
	result.Assert(t, icmd.Success)

	// // // get
	// result = icmd.Command("./flow", "transactions", "get", "LKJALKJLKDJFLJFOIJSDFJLKDSJFLKIOEJF")
	// resp, err = cmd.CombinedOutput()
	// assert.Nil(t, err, "Failed to sign transaction: ", err, resp)
}

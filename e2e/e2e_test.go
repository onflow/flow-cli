package e2e

import (
	"bytes"
	"fmt"
	"os/exec"
	"testing"
)

var PROMPT_UI_BUFFER_SIZE = 4096
var ASCII_A byte = 97

// PromptUI reads prompt inputs with 4096 length buffers so we pad
// the rest of the bytes with some value after our input or else
// PromptUI will throw us an unexpected EOF encountered error
func pad(siz int, buf *bytes.Buffer) {
	pu := make([]byte, PROMPT_UI_BUFFER_SIZE-siz)
	for i := 0; i < PROMPT_UI_BUFFER_SIZE-siz; i++ {
		pu[i] = ASCII_A // some arbitrary character for padding, in this case ascii 'a'
	}
	buf.Write(pu)
}

func Test_keys(t *testing.T) {
	// defer icmd.RunCommand("rm", "flow")

	// result := icmd.RunCommand("./flow", "init")
	// TODO: just run in background for now, leave as todo for later
	// wait for port somehow, try script first then pm2, maybe run a script from go
	// emulatorResult := icmd.RunCommand("./flow", "emulator", "init", "&")

	// NOTE: how to check this output as this is random
	// sign something with the key
	// TODO: extract key and name of account for later use
	// ### KEYS ###
	cmd := exec.Command("./flow", "keys", "generate")
	res, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", res)

	// ### ACCOUNTS AND KEYS ###
	// TODO: defer method to remove this from flow.json or remove flow.json altogether
	cmd = exec.Command("./flow", "accounts", "create")

	buf := bytes.Buffer{}

	// first input text
	accountName := "gamer\r\n"
	buf.WriteString(accountName)
	pad(len(accountName), &buf)

	// second input: enter key
	buf.Write([]byte{10})
	pad(1, &buf)

	// third input "y" for yes and enter key
	buf.Write([]byte{121, 10})
	pad(2, &buf)

	cmd.Stdin = &buf

	res, err = cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", res)

	cmd = exec.Command("./flow", "accounts", "add-contract", "HelloWorld", "./files/HelloWorld.cdc")
	res, err = cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", res)

	cmd = exec.Command("./flow", "scripts", "execute", "./files/script.cdc")
	res, err = cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", res)

	cmd = exec.Command("./flow", "accounts", "update-contract", "HelloWorld", "./files/HelloWorldUpdated.cdc")
	res, err = cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", res)

	cmd = exec.Command("./flow", "scripts", "execute", "./files/script.cdc")
	res, err = cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", res)

	cmd = exec.Command("./flow", "accounts", "remove-contract", "HelloWorld")
	res, err = cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", res)

	// ### TRANSACTIONS ###
}

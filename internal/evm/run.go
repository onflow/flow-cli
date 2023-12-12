package evm

import (
	"context"
	_ "embed"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/onflow/cadence"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/arguments"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"
)

//go:embed run.cdc
var runCode []byte

type flagsRun struct {
	Signer string `default:"" flag:"signer" info:"Account name from configuration used to sign the transaction as proposer, payer and suthorizer"`
	ABI    string `default:"" flag:"ABI" info:"ABI specification for the deployed contract"`
}

var runFlags = flagsRun{}

var runCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "run <caller address> <contract address> <function name> <arg1, arg2...>",
		Short:   "Execute a contract function by its' name and with provided parameters",
		Args:    cobra.MinimumNArgs(2),
		Example: "flow evm run 522b3294e6d06aa25ad0f1b8891242e335d3b459 balance",
	},
	Flags: &runFlags,
	RunS:  run,
}

func run(
	args []string,
	g command.GlobalFlags,
	_ output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	callerAddress := args[0]
	contractAddress := args[1]
	funcName := args[2]
	funcArgs := args[3:]
	abiFilename := runFlags.ABI

	if abiFilename == "" {
		return nil, fmt.Errorf("must provide the ABI file location, using the ABI flag")
	}

	abiFile, err := os.Open(abiFilename)
	if err != nil {
		return nil, fmt.Errorf("can't open ABI file: %w", err)
	}

	ABI, err := abi.JSON(abiFile)
	if err != nil {
		return nil, fmt.Errorf("can't deserialize ABI file: %w", err)
	}

	var data []byte
	if len(funcArgs) == 0 {
		data, err = ABI.Pack(funcName)
	} else {
		data, err = ABI.Pack(funcName, funcArgs)
	}
	if err != nil {
		return nil, fmt.Errorf("can't prepare arguments: %w", err)
	}

	decodedAddress, err := hex.DecodeString(contractAddress)
	if err != nil {
		return nil, err
	}

	cadenceArgs := []string{
		callerAddress,
		cadenceByteArrayString(decodedAddress),
		cadenceByteArrayString(data),
	}

	scriptArgs, err := arguments.ParseWithoutType(cadenceArgs, runCode, "")
	if err != nil {
		return nil, fmt.Errorf("can't parse cadence arguments: %w", err)
	}

	val, err := flow.ExecuteScript(
		context.Background(),
		flowkit.Script{
			Code: runCode,
			Args: scriptArgs,
		},
		flowkit.ScriptQuery{Latest: true},
	)
	if err != nil {
		return nil, err
	}

	cadenceArray, ok := val.(cadence.Array)
	if !ok {
		return nil, fmt.Errorf("script doesn't return byte array as it should")
	}

	byteArray := make([]byte, len(cadenceArray.Values))
	for i, x := range cadenceArray.Values {
		byteArray[i] = x.ToGoValue().(byte)
	}

	funcABI := ABI.Methods[funcName]
	unpackedValues, err := funcABI.Outputs.Unpack(byteArray)
	if err != nil {
		return nil, fmt.Errorf("unpack failed: %w", err)
	}

	fmt.Println("Result: ", unpackedValues)

	return nil, nil
}

func cadenceByteArrayString(data []byte) string {
	raw := fmt.Sprintf("%v", data)
	return strings.ReplaceAll(raw, " ", ",")
}

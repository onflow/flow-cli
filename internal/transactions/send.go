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

package transactions

import (
	"encoding/hex"
	"fmt"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-go-sdk"

	"github.com/spf13/cobra"

	"github.com/onflow/cadence"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsSend struct {
	ArgsJSON string   `default:"" flag:"args-json" info:"arguments in JSON-Cadence format"`
	Arg      []string `default:"" flag:"arg" info:"⚠️  Deprecated: use command arguments"`
	Signer   []string `default:"emulator-account" flag:"signer" info:"Account name(s) from configuration used to sign the transaction"`
	GasLimit uint64   `default:"1000" flag:"gas-limit" info:"transaction gas limit"`
	Include  []string `default:"" flag:"include" info:"Fields to include in the output"`
	Exclude  []string `default:"" flag:"exclude" info:"Fields to exclude from the output (events)"`
}

var sendFlags = flagsSend{}

var SendCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "send <code filename> [<argument> <argument> ...]",
		Short:   "Send a transaction",
		Args:    cobra.MinimumNArgs(1),
		Example: `flow transactions send tx.cdc "Hello world"`,
	},
	Flags: &sendFlags,
	RunS:  send,
}

func send(
	args []string,
	readerWriter flowkit.ReaderWriter,
	globalFlags command.GlobalFlags,
	srv *services.Services,
	state *flowkit.State,
) (command.Result, error) {
	codeFilename := args[0]

	var signed *flowkit.Transaction
	var signers []*flowkit.Account

	//validate all signers
	for _, signerName := range sendFlags.Signer {
		// use service account by default
		if signerName == config.DefaultEmulatorServiceAccountName {
			signerName = state.Config().Emulators.Default().ServiceAccount
		}
		signer, err := state.Accounts().ByName(signerName)
		if err != nil {
			return nil, fmt.Errorf("signer account: [%s] doesn't exists in configuration", signerName)
		}
		signers = append(signers, signer)
	}

	code, err := readerWriter.ReadFile(codeFilename)
	if err != nil {
		return nil, fmt.Errorf("error loading transaction file: %w", err)
	}

	//find authorizer count from code
	authorizerCount := flowkit.GetAuthorizerCount(codeFilename, code)

	var seen map[flow.Address]any = make(map[flow.Address]any)
	var authorizers []flow.Address
	for _, auth := range signers {
		if _, ok := seen[auth.Address()]; ok {
			continue
		}
		authorizers = append(authorizers, auth.Address())
	}

	if len(authorizers) < authorizerCount {
		return nil, fmt.Errorf("invalid number of authorizers, expected: %d", authorizerCount)
	}

	//remove extra signers
	authorizers = authorizers[:authorizerCount]

	//proposer is the first signer
	proposer := signers[0]

	//payer signs last
	payer := signers[len(signers)-1]

	if len(sendFlags.Arg) != 0 {
		fmt.Println("⚠️  DEPRECATION WARNING: use transaction arguments as command arguments: send <code filename> [<argument> <argument> ...]")
	}

	var transactionArgs []cadence.Value
	if sendFlags.ArgsJSON != "" {
		transactionArgs, err = flowkit.ParseArgumentsJSON(sendFlags.ArgsJSON)
	} else {
		transactionArgs, err = flowkit.ParseArgumentsWithoutType(codeFilename, code, args[1:])
	}
	if err != nil {
		return nil, fmt.Errorf("error parsing transaction arguments: %w", err)
	}

	//payload generation
	build, err := services.Transactions.Build(
		proposer.Address(),
		authorizers,
		payer.Address(),
		proposer.Key().Index(),
		code,
		codeFilename,
		buildFlags.GasLimit,
		transactionArgs,
		globalFlags.Network,
		true,
	)
	if err != nil {
		return nil, err
	}

	encoded := build.FlowTransaction().Encode()
	payload := []byte(hex.EncodeToString(encoded))

	for _, signer := range signers {
		signed, err = services.Transactions.Sign(signer, payload, true)
		if err != nil {
			return nil, err
		}
		payload = []byte(hex.EncodeToString(signed.FlowTransaction().Encode()))
	}

	tx, result, err := services.Transactions.SendSigned(
		payload,
		true,
	)

	if err != nil {
		return nil, err
	}

	return &TransactionResult{
		result:  result,
		tx:      tx,
		include: sendFlags.Include,
		exclude: sendFlags.Exclude,
	}, nil
}

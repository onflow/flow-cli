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
	"fmt"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"

	"github.com/onflow/cadence"
	"github.com/spf13/cobra"
)

type flagsSend struct {
	ArgsJSON  string   `default:"" flag:"args-json" info:"arguments in JSON-Cadence format"`
	Signer    string   `default:"" flag:"signer" info:"Account name from configuration used to sign the transaction as proposer, payer and suthorizer"`
	Proposer  string   `default:"" flag:"proposer" info:"Account name from configuration used as proposer"`
	Payer     string   `default:"" flag:"payer" info:"Account name from configuration used as payer"`
	Autorizer []string `default:"" flag:"authorizer" info:"Name of a single or multiple comma-separated accounts used as authorizers from configuration"`
	Include   []string `default:"" flag:"include" info:"Fields to include in the output"`
	Exclude   []string `default:"" flag:"exclude" info:"Fields to exclude from the output (events)"`
	GasLimit  uint64   `default:"1000" flag:"gas-limit" info:"transaction gas limit"`
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
) (result command.Result, err error) {
	codeFilename := args[0]

	proposerName := sendFlags.Proposer
	var proposer *flowkit.Account
	if proposerName != "" {
		proposer, err = state.Accounts().ByName(proposerName)
		if err != nil {
			return nil, fmt.Errorf("proposer account: [%s] doesn't exists in configuration", proposerName)
		}
	}

	payerName := sendFlags.Payer
	var payer *flowkit.Account
	if payerName != "" {
		payer, err = state.Accounts().ByName(payerName)
		if err != nil {
			return nil, fmt.Errorf("payer account: [%s] doesn't exists in configuration", payerName)
		}
	}

	var authorizers []*flowkit.Account
	for _, authorizerName := range sendFlags.Autorizer {
		authorizer, err := state.Accounts().ByName(authorizerName)
		if err != nil {
			return nil, fmt.Errorf("authorizer account: [%s] doesn't exists in configuration", authorizerName)
		}
		authorizers = append(authorizers, authorizer)
	}

	signerName := sendFlags.Signer

	if signerName == "" && proposer == nil && payer == nil && len(authorizers) == 0 {
		signerName = state.Config().Emulators.Default().ServiceAccount
	}

	if signerName != "" {
		if proposer != nil || payer != nil || len(authorizers) > 0 {
			return nil, fmt.Errorf("signer flag cannot be combined with payer/proposer/authorizer flags")
		}
		signer, err := state.Accounts().ByName(signerName)
		if err != nil {
			return nil, fmt.Errorf("signer account: [%s] doesn't exists in configuration", signerName)
		}
		proposer = signer
		payer = signer
		authorizers = append(authorizers, signer)
	}

	code, err := readerWriter.ReadFile(codeFilename)
	if err != nil {
		return nil, fmt.Errorf("error loading transaction file: %w", err)
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

	roles, err := services.NewTransactionAccountRoles(proposer, payer, authorizers)
	if err != nil {
		return nil, fmt.Errorf("error parsing transaction roles: %w", err)
	}

	tx, txResult, err := srv.Transactions.Send(
		roles,
		flowkit.NewScript(code, transactionArgs, codeFilename),
		sendFlags.GasLimit,
		globalFlags.Network)

	if err != nil {
		return nil, err
	}

	return &TransactionResult{
		result:  txResult,
		tx:      tx,
		include: sendFlags.Include,
		exclude: sendFlags.Exclude,
	}, nil
}

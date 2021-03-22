/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
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

package utils

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	cli "github.com/onflow/flow-cli/flow"
	"github.com/onflow/flow-go-sdk"
)

func LoadTransactionPayloadFromFile(filename string) *flow.Transaction {
	partialTxHex, err := ioutil.ReadFile(filename)
	if err != nil {
		cli.Exitf(1, "Failed to read partial transaction from %s: %v", filename, err)
	}
	partialTxBytes, err := hex.DecodeString(string(partialTxHex))
	if err != nil {
		cli.Exitf(1, "Failed to decode partial transaction from %s: %v", filename, err)
	}
	tx, err := flow.DecodeTransaction(partialTxBytes)
	if err != nil {
		cli.Exitf(1, "Failed to decode transaction from %s: %v", filename, err)
	}

	return tx
}

func NewTransactionWithCodeArgsAuthorizers(code string, args string, authorizers []string) *flow.Transaction {
	tx := flow.NewTransaction()
	if code != "" {
		codeBytes, err := ioutil.ReadFile(code)
		if err != nil {
			cli.Exitf(1, "Failed to read transaction script from %s: %v", code, err)
		}
		tx.SetScript(codeBytes)
	}

	// Arguments
	if args != "" {
		transactionArguments, err := cli.ParseArguments(args)
		if err != nil {
			cli.Exitf(1, "Invalid arguments passed: %s", args)
		}

		for _, arg := range transactionArguments {
			err := tx.AddArgument(arg)

			if err != nil {
				cli.Exitf(1, "Failed to add %s argument to a transaction ", args)
			}
		}
	}

	if len(authorizers) > 0 {
		for _, authorizer := range authorizers {
			authorizerAddress := flow.HexToAddress(authorizer)
			if authorizerAddress == flow.EmptyAddress {
				cli.Exitf(1, "Invalid authorizer address provided %s", authorizer)
			}

			tx.AddAuthorizer(authorizerAddress)
		}
	}

	return tx
}

// StdInYesOrError
// This function currently reads from stdin, and will error on anything other that
// y or yes (case insensitive).
// If there is ever a use case where we do not want to exit right away on a non-yes
// response form the user, this function will need to be changed
func StdInYesOrError() bool {
	buf := bufio.NewReader(os.Stdin)
	fmt.Print("> ")
	input, err := buf.ReadString('\n')
	input = strings.TrimSpace(input)
	if err != nil {
		cli.Exitf(1, "Unable to read input %s", err)
	} else if strings.EqualFold(input, "y") || strings.EqualFold(input, "yes") {
		return true
	}
	cli.Exitf(1, "Input not accepted (%s), exiting", input)
	return false
}

func DisplayTransactionForConfirmation(tx *flow.Transaction, autoConfirm bool) {
	// Display authorizers
	if len(tx.Authorizers) == 0 {
		fmt.Println("No authorizers")
	} else {
		fmt.Printf("Authorizers (%d):\n", len(tx.Authorizers))
		for i, authorizer := range tx.Authorizers {
			fmt.Printf(cli.Indent+"Authorizer %d: %s\n", i, authorizer)
		}
	}
	fmt.Println()
	// Display Arguments
	if len(tx.Arguments) == 0 {
		fmt.Println("No arguments")
	} else {
		fmt.Printf("Arguments (%d):\n", len(tx.Arguments))
		for i, argument := range tx.Arguments {
			fmt.Printf(cli.Indent+"Argument %d: %s\n", i, argument)
		}
	}
	fmt.Println()
	// Display Code
	fmt.Println("Script:")
	fmt.Println(string(tx.Script))
	// Display Payer
	fmt.Println("===")
	fmt.Printf("Proposer Address: %s\n", tx.ProposalKey.Address)
	fmt.Printf("Proposer Key Index: %d\n", tx.ProposalKey.KeyIndex)
	fmt.Printf("Payer Address: %s\n", tx.Payer)
	fmt.Println("===")
	if len(tx.PayloadSignatures) == 0 {
		fmt.Println("No payload signatures")
	} else {
		fmt.Printf("Payload Signatures (%d):\n", len(tx.PayloadSignatures))
		for i, sig := range tx.PayloadSignatures {
			fmt.Printf(cli.Indent+"Signature %d Signer Address: %s\n", i, sig.Address)
			fmt.Printf(cli.Indent+"Signature %d Signer Key Index: %d\n", i, sig.KeyIndex)
		}
	}
	fmt.Println("===")
	if autoConfirm {
		fmt.Println("Auto confirming correctness of payload")
	} else {
		fmt.Println("Does this look correct? (y/n)")
		StdInYesOrError()
		fmt.Println("Payload contents verified")
	}
}

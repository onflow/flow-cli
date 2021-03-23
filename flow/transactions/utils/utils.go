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
	"encoding/hex"
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

func AssertEmptyOnLoadingPayload(shouldBeEmpty string, paramName string) {
	if strings.TrimSpace(shouldBeEmpty) != "" {
		cli.Exitf(1, "Loading transaction payload from file, but %s provided which cannot be used", paramName)
	}
}

func ValidateKeyPreReq(account *cli.Account) {
	if account == nil {
		cli.Exitf(1, "A specified key was not found")
	}
	if account.KeyType == cli.KeyTypeHex {
		// Always Valid
		return
	} else if account.KeyType == cli.KeyTypeKMS {
		// Check GOOGLE_APPLICATION_CREDENTIALS
		googleAppCreds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		if len(googleAppCreds) == 0 {
			if len(account.KeyContext["projectId"]) == 0 {
				cli.Exitf(1, "Could not get GOOGLE_APPLICATION_CREDENTIALS, no google service account json provided but private key type is KMS", account.Address)
			}
			cli.GcloudApplicationSignin(account.KeyContext["projectId"])
		}
		return
	}
	cli.Exitf(1, "Failed to validate %s key for %s", account.KeyType, account.Address)

}

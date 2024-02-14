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

package migration

import (
	"bytes"
	"fmt"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit"
	"github.com/onflow/flowkit/accounts"

	"github.com/onflow/flow-cli/internal/util"
)

// TODO: update these  once deployed
var migrationContractStagingAddress = map[string]string{
	"testnet": "0xa983fecbed621163",
	"mainnet": "0xa983fecbed621163",
}

// MigrationContractStagingAddress returns the address of the migration contract on the given network
func MigrationContractStagingAddress(network string) flow.Address {
	return flow.HexToAddress(migrationContractStagingAddress[network])
}

func getAccountByContractName(state *flowkit.State, contractName string, network string) (*accounts.Account, error) {
	deployments := state.Deployments().ByNetwork(network)
	var accountName string
	for _, d := range deployments {
		for _, c := range d.Contracts {
			if c.Name == contractName {
				accountName = d.Account
				break
			}
		}
	}
	if accountName == "" {
		return nil, fmt.Errorf("contract not found in state")
	}

	accs := state.Accounts()
	if accs == nil {
		return nil, fmt.Errorf("no accounts found in state")
	}

	var account *accounts.Account
	for _, a := range *accs {
		if accountName == a.Name {
			account = &a
			break
		}
	}
	if account == nil {
		return nil, fmt.Errorf("account %s not found in state", accountName)
	}

	return account, nil
}

type migrationResult struct {
	result  string
	message string
	key     accounts.Key
}

func (s *migrationResult) JSON() any {
	return map[string]string{
		"signature": fmt.Sprintf("%x", s.result),
		"message":   s.message,
		"hashAlgo":  s.key.HashAlgo().String(),
		"sigAlgo":   s.key.SigAlgo().String(),
	}
}

func (s *migrationResult) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)

	_, _ = fmt.Fprintf(writer, "Signature \t %x\n", s.result)
	_, _ = fmt.Fprintf(writer, "Message \t %s\n", s.message)
	_, _ = fmt.Fprintf(writer, "Hash Algorithm \t %s\n", s.key.HashAlgo())
	_, _ = fmt.Fprintf(writer, "Signature Algorithm \t %s\n", s.key.SigAlgo())

	_ = writer.Flush()
	return b.String()
}

func (s *migrationResult) Oneliner() string {

	return fmt.Sprintf(
		"signature: %x, message: %s, hashAlgo: %s, sigAlgo: %s",
		s.result, s.message, s.key.HashAlgo(), s.key.SigAlgo(),
	)
}

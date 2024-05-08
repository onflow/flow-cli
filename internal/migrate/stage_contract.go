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

/*
TODO: handle the broken dependency graph case.

e.g. Foo -> Bar -> Baz
but only Foo & Baz are staged, so how to build the contract graph?
*/

package migrate

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/output"
	"github.com/onflow/flowkit/v2/project"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
)

type stagingResult struct {
	// Error will be nil if the contract was successfully staged
	Contracts map[common.AddressLocation]error
}

var _ command.ResultWithExitCode = &stagingResult{}

var stageContractflags struct {
	All            bool     `default:"false" flag:"all" info:"Stage all contracts"`
	Accounts       []string `default:"" flag:"account" info:"Accounts to stage the contract under"`
	SkipValidation bool     `default:"false" flag:"skip-validation" info:"Do not validate the contract code against staged dependencies"`
}

var stageContractCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "stage-contract <CONTRACT_NAME>",
		Short:   "stage a contract for migration",
		Example: `flow migrate stage-contract HelloWorld`,
		Args:    cobra.MinimumNArgs(1),
	},
	Flags: &stageContractflags,
	RunS:  stageContract,
}

func buildContract(state *flowkit.State, flow flowkit.Services, contract *config.Contract) (*project.Contract, error) {
	contractName := contract.Name

	replacedCode, err := replaceImportsIfExists(state, flow, contract.Location)
	if err != nil {
		return nil, fmt.Errorf("failed to replace imports: %w", err)
	}

	account, err := getAccountByContractName(state, contractName, flow.Network())
	if err != nil {
		return nil, fmt.Errorf("failed to get account by contract name: %w", err)
	}

	return project.NewContract(contractName, filepath.Clean(contract.Location), replacedCode, account.Address, account.Name, nil), nil
}

func stageContract(
	args []string,
	globalFlags command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	err := checkNetwork(flow.Network())
	if err != nil {
		return nil, err
	}

	var results map[common.AddressLocation]error
	s := newStagingService(flow, state, logger, !stageContractflags.SkipValidation)

	// Stage all contracts
	if stageContractflags.All {
		if len(args) > 0 || len(stageContractflags.Accounts) > 0 {
			return nil, fmt.Errorf("cannot use --all flag with contract names or --accounts flag")
		}

		results, err = s.StageContracts(context.Background(), nil)
	}

	// Filter by contract names
	if len(args) > 0 {
		if len(stageContractflags.Accounts) > 0 {
			return nil, fmt.Errorf("cannot use --account flag with contract names")
		}

		results, err = s.StageContracts(context.Background(), func(c *project.Contract) bool {
			for _, name := range args {
				if c.Name == name {
					return true
				}
			}
			return false
		})
	}

	// Filter by accounts
	if len(stageContractflags.Accounts) > 0 {
		results, err = s.StageContracts(context.Background(), func(c *project.Contract) bool {
			for _, account := range stageContractflags.Accounts {
				if c.AccountName == account {
					return true
				}
			}
			return false
		})
	}

	if err != nil {
		return nil, err
	}

	// Print the results
	return &stagingResult{
		Contracts: results,
	}, nil
}

func (r *stagingResult) ExitCode() int {
	for _, err := range r.Contracts {
		if err != nil {
			return 1
		}
	}
	return 0
}

func (s *stagingResult) String() string {
	if len(s.Contracts) == 0 {
		return "no contracts staged"
	}

	sb := &strings.Builder{}

	// First, print the failing contracts
	for location, err := range s.Contracts {
		if err != nil {
			sb.WriteString(fmt.Sprintf("failed to stage contract %s: %s\n", location, err))
		}
	}

	// Then, print the successfully staged contracts
	for location, err := range s.Contracts {
		if err == nil {
			sb.WriteString(fmt.Sprintf("staged contract %s\n", location))
		}
	}

	sb.WriteString(fmt.Sprintf("staged %d contracts", len(s.Contracts)))

	return sb.String()
}

func (s *stagingResult) JSON() interface{} {
	return s
}

func (r *stagingResult) Oneliner() string {
	if len(r.Contracts) == 0 {
		return "no contracts staged"
	}
	return fmt.Sprintf("staged %d contracts", len(r.Contracts))
}

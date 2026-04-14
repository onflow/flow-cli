/*
 * Flow CLI
 *
 * Copyright Flow Foundation
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

package diffcontract

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/spf13/cobra"

	flowsdk "github.com/onflow/flow-go-sdk"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"
	"github.com/onflow/flowkit/v2/project"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

type diffContractFlags struct {
	Quiet bool `default:"false" flag:"quiet" info:"Exit with non-zero code if contracts differ, without output"`
}

var diffFlags = diffContractFlags{}

var DiffContractCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "diff-contract <file-or-url> [address]",
		Short:   "Diff a local contract against a deployed one",
		Example: "flow diff-contract ./MyContract.cdc\nflow diff-contract ./MyContract.cdc 0xf8d6e0586b0a20c7\nflow diff-contract https://example.com/MyContract.cdc my-account --network testnet",
		Args:    cobra.RangeArgs(1, 2),
		GroupID: "tools",
	},
	Flags: &diffFlags,
	RunS:  diffContract,
}

func init() {
	DiffContractCommand.Cmd.Flags().BoolVarP(&diffFlags.Quiet, "quiet", "q", false, "Exit with non-zero code if contracts differ, without output")
}

func diffContract(
	args []string,
	globalFlags command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	source := args[0]

	// Read source code from file or URL
	var code []byte
	var location string
	var err error

	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		code, err = fetchURL(source)
		if err != nil {
			return nil, fmt.Errorf("error fetching contract from URL: %w", err)
		}
		location = source
	} else {
		code, err = state.ReadFile(source)
		if err != nil {
			return nil, fmt.Errorf("error loading contract file: %w", err)
		}
		location = source
	}

	// Extract contract name from source
	program, err := project.NewProgram(code, nil, location)
	if err != nil {
		return nil, fmt.Errorf("error parsing contract source: %w", err)
	}

	contractName, err := program.Name()
	if err != nil {
		return nil, fmt.Errorf("error extracting contract name: %w", err)
	}

	// Resolve imports in source code
	ctx := context.Background()
	resolved, err := flow.ReplaceImportsInScript(ctx, flowkit.Script{
		Code:     code,
		Location: location,
	})
	if err != nil {
		return nil, fmt.Errorf("error resolving imports: %w", err)
	}

	// Resolve target address: from argument or from flow.json deployments
	var address flowsdk.Address
	if len(args) >= 2 {
		address, err = util.ResolveAddressOrAccountNameForNetworks(args[1], state, []string{globalFlags.Network})
		if err != nil {
			return nil, err
		}
	} else {
		address, err = resolveAddressFromConfig(state, contractName, globalFlags.Network)
		if err != nil {
			return nil, err
		}
	}

	// Fetch deployed contract
	logger.StartProgress(fmt.Sprintf("Fetching contract '%s' from %s...", contractName, address))
	defer logger.StopProgress()

	account, err := flow.GetAccount(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("error fetching account: %w", err)
	}

	deployedCode, ok := account.Contracts[contractName]
	if !ok {
		return nil, fmt.Errorf("contract '%s' not found on account %s", contractName, address)
	}

	// Normalize and diff
	localCode := util.NormalizeLineEndings(string(resolved.Code))
	remoteCode := util.NormalizeLineEndings(string(deployedCode))

	identical := localCode == remoteCode

	exitCode := 0
	if !identical {
		exitCode = 1
	}

	diffText := ""
	if !identical {
		localLabel := source
		remoteLabel := fmt.Sprintf("0x%s/%s (deployed)", address, contractName)
		diff := difflib.UnifiedDiff{
			A:        difflib.SplitLines(remoteCode),
			B:        difflib.SplitLines(localCode),
			FromFile: remoteLabel,
			ToFile:   localLabel,
			Context:  3,
		}
		diffText, err = difflib.GetUnifiedDiffString(diff)
		if err != nil {
			return nil, fmt.Errorf("error computing diff: %w", err)
		}
	}

	return &diffContractResult{
		diff:         diffText,
		contractName: contractName,
		address:      address.String(),
		identical:    identical,
		quiet:        diffFlags.Quiet,
		exitCode:     exitCode,
	}, nil
}

// resolveAddressFromConfig looks up the address for a contract in flow.json
// by checking deployments first, then contract aliases for the given network.
func resolveAddressFromConfig(state *flowkit.State, contractName string, network string) (flowsdk.Address, error) {
	// Check deployments
	deployments := state.Deployments().ByNetwork(network)
	for _, deployment := range deployments {
		for _, contract := range deployment.Contracts {
			if contract.Name == contractName {
				account, err := state.Accounts().ByName(deployment.Account)
				if err != nil {
					return flowsdk.EmptyAddress, fmt.Errorf("account '%s' from deployment not found in configuration: %w", deployment.Account, err)
				}
				return account.Address, nil
			}
		}
	}

	// Check contract aliases
	contract, err := state.Contracts().ByName(contractName)
	if err == nil && contract != nil {
		if alias := contract.Aliases.ByNetwork(network); alias != nil {
			return alias.Address, nil
		}
	}

	return flowsdk.EmptyAddress, fmt.Errorf("contract '%s' not found in deployments or aliases for network '%s' in flow.json, specify an address explicitly", contractName, network)
}

func fetchURL(url string) ([]byte, error) {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// diffContractResult implements command.ResultWithExitCode
type diffContractResult struct {
	diff         string
	contractName string
	address      string
	identical    bool
	quiet        bool
	exitCode     int
}

var _ command.ResultWithExitCode = &diffContractResult{}

func (r *diffContractResult) String() string {
	if r.quiet {
		return ""
	}
	if r.identical {
		return fmt.Sprintf("Contract '%s' on 0x%s is up to date", r.contractName, r.address)
	}
	return r.diff
}

func (r *diffContractResult) Oneliner() string {
	if r.identical {
		return "identical"
	}
	return "different"
}

func (r *diffContractResult) JSON() any {
	result := map[string]any{
		"contract":  r.contractName,
		"address":   r.address,
		"identical": r.identical,
	}
	if !r.identical {
		result["diff"] = r.diff
	}
	return result
}

func (r *diffContractResult) ExitCode() int {
	return r.exitCode
}

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

package accounts

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/onflow/cadence"
	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/accounts"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/gateway"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

type flagsList struct{}

var listFlags = flagsList{}

var listCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "list",
		Short:   "List all accounts from flow.json and validate them against networks",
		Example: "flow accounts list",
		Args:    cobra.NoArgs,
	},
	Flags: &listFlags,
	RunS:  list,
}

type accountOnNetwork struct {
	Name      string   `json:"name"`
	Address   string   `json:"address"`
	Balance   string   `json:"balance,omitempty"`
	Contracts []string `json:"contracts,omitempty"`
	Exists    bool     `json:"exists"`
	Error     string   `json:"error,omitempty"`
}

type networkResult struct {
	Name     string             `json:"name"`
	Host     string             `json:"host"`
	Accounts []accountOnNetwork `json:"accounts"`
	Warning  string             `json:"warning,omitempty"`
}

type accountsListResult struct {
	Networks         []networkResult    `json:"networks"`
	AccountsNotFound []accountOnNetwork `json:"accounts_not_found"`
}

func (r *accountsListResult) JSON() any {
	return r
}

func (r *accountsListResult) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)

	_, _ = fmt.Fprintf(writer, "Accounts by Network:\n\n")

	for _, network := range r.Networks {
		_, _ = fmt.Fprintf(writer, "%s:\n", network.Name)

		if network.Warning != "" {
			_, _ = fmt.Fprintf(writer, "  ⚠️  Warning: %s\n", network.Warning)
		}

		if len(network.Accounts) == 0 {
			_, _ = fmt.Fprintf(writer, "  No accounts found\n")
		} else {
			for _, account := range network.Accounts {
				if account.Exists {
					_, _ = fmt.Fprintf(writer, "  - %s (%s): %s FLOW\n",
						account.Name, account.Address, account.Balance)
				} else {
					status := "Account not found"
					if account.Error != "" {
						status = account.Error
					}
					_, _ = fmt.Fprintf(writer, "  - %s (%s): %s\n",
						account.Name, account.Address, status)
				}
			}
		}
		_, _ = fmt.Fprintf(writer, "\n")
	}

	if len(r.AccountsNotFound) > 0 {
		_, _ = fmt.Fprintf(writer, "Accounts not found on any network:\n")
		for _, account := range r.AccountsNotFound {
			_, _ = fmt.Fprintf(writer, "  - %s (%s)\n", account.Name, account.Address)
		}
	}

	_ = writer.Flush()
	return b.String()
}

func (r *accountsListResult) Oneliner() string {
	totalAccounts := 0
	totalNetworks := len(r.Networks)

	for _, network := range r.Networks {
		for _, account := range network.Accounts {
			if account.Exists {
				totalAccounts++
			}
		}
	}

	return fmt.Sprintf("%d accounts found across %d networks", totalAccounts, totalNetworks)
}

func isEmulatorRunning(host string) bool {
	conn, err := net.DialTimeout("tcp", host, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func validateAccountOnNetwork(account *accounts.Account, network *config.Network, logger output.Logger) accountOnNetwork {
	result := accountOnNetwork{
		Name:    account.Name,
		Address: account.Address.String(),
		Exists:  false,
	}

	// Check if emulator is running before trying to connect
	if network.Name == "emulator" || strings.Contains(network.Host, "127.0.0.1") || strings.Contains(network.Host, "localhost") {
		if !isEmulatorRunning(network.Host) {
			result.Error = fmt.Sprintf("Emulator not running on %s", network.Host)
			return result
		}
	}

	var gw gateway.Gateway
	var err error

	gw, err = gateway.NewGrpcGateway(*network)

	if err != nil {
		result.Error = fmt.Sprintf("Failed to create gateway: %v", err)
		return result
	}

	flow := flowkit.NewFlowkit(nil, *network, gw, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	flowAccount, err := flow.GetAccount(ctx, account.Address)
	if err != nil {
		if strings.Contains(err.Error(), "invalid address") {
			result.Error = "Invalid address for this network"
		} else if strings.Contains(err.Error(), "connection refused") {
			result.Error = "Network unreachable"
		} else if strings.Contains(err.Error(), "not found") {
			result.Error = "Account not found"
		} else {
			result.Error = "Connection failed"
		}
		return result
	}

	result.Exists = true
	result.Balance = cadence.UFix64(flowAccount.Balance).String()

	return result
}

func list(
	_ []string,
	_ command.GlobalFlags,
	logger output.Logger,
	_ flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	if state == nil {
		return nil, fmt.Errorf("flow.json not found, please run 'flow init' first")
	}

	accounts := state.Accounts()
	if accounts == nil || len(*accounts) == 0 {
		return &accountsListResult{
			Networks:         []networkResult{},
			AccountsNotFound: []accountOnNetwork{},
		}, nil
	}

	networks := state.Networks()
	if networks == nil || len(*networks) == 0 {
		return nil, fmt.Errorf("no networks configured in flow.json")
	}

	result := &accountsListResult{
		Networks:         make([]networkResult, 0, len(*networks)),
		AccountsNotFound: []accountOnNetwork{},
	}

	accountsFound := make(map[string]bool)

	for _, network := range *networks {
		networkRes := networkResult{
			Name:     network.Name,
			Host:     network.Host,
			Accounts: make([]accountOnNetwork, 0, len(*accounts)),
		}

		if network.Name == "emulator" || strings.Contains(network.Host, "127.0.0.1") || strings.Contains(network.Host, "localhost") {
			if !isEmulatorRunning(network.Host) {
				networkRes.Warning = fmt.Sprintf("Emulator not running on %s", network.Host)
			}
		}

		logger.StartProgress(fmt.Sprintf("Checking accounts on %s...", network.Name))

		// Check each account on this network
		for _, account := range *accounts {
			accountResult := validateAccountOnNetwork(&account, &network, logger)
			networkRes.Accounts = append(networkRes.Accounts, accountResult)

			if accountResult.Exists {
				accountsFound[account.Name] = true
			}
		}

		logger.StopProgress()
		result.Networks = append(result.Networks, networkRes)
	}

	// Find accounts not found on any network
	for _, account := range *accounts {
		if !accountsFound[account.Name] {
			result.AccountsNotFound = append(result.AccountsNotFound, accountOnNetwork{
				Name:    account.Name,
				Address: account.Address.String(),
				Exists:  false,
			})
		}
	}

	return result, nil
}

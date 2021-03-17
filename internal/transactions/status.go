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

package transactions

import (
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flow"
	"github.com/onflow/flow-cli/pkg/flow/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type flagsStatus struct {
	Host   string `flag:"host" info:"Flow Access API host address"`
	Sealed bool   `default:"true" flag:"sealed" info:"Wait for a sealed result"`
	Code   bool   `default:"false" flag:"code" info:"Display transaction code"`
}

type cmdStatus struct {
	cmd   *cobra.Command
	flags flagsStatus
}

// NewStatusCmd new status command
func NewStatusCmd() command.Command {
	return &cmdStatus{
		cmd: &cobra.Command{
			Use:   "status <tx_id>",
			Short: "Get the transaction status",
			Args:  cobra.ExactArgs(1),
		},
	}
}

// Run command
func (s *cmdStatus) Run(
	cmd *cobra.Command,
	args []string,
	project *flow.Project,
	services *services.Services,
) (command.Result, error) {
	tx, result, err := services.Transactions.GetStatus(args[0], s.flags.Sealed)
	return &TransactionResult{
		result: result,
		tx:     tx,
		code:   s.flags.Code,
	}, err
}

// GetFlags for command
func (s *cmdStatus) GetFlags() *sconfig.Config {
	return sconfig.New(&s.flags)
}

// GetCmd gets command
func (s *cmdStatus) GetCmd() *cobra.Command {
	return s.cmd
}

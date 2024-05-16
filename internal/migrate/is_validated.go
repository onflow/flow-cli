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

package migrate

import (
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/logrusorgru/aurora/v4"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/migrate/validator"
	"github.com/onflow/flow-cli/internal/util"
)

const moreInformationMessage = "For more information, please find the latest full migration report on GitHub (https://github.com/onflow/cadence/tree/master/migrations_data).\n\nNew reports are generated after each weekly emulated migration and your contract's status may change, so please actively monitor this status and stay tuned for the latest announcements until the migration deadline."

type validationResult struct {
	Timestamp time.Time
	Status    validator.ContractUpdateStatus
	Network   string
}

var isValidatedflags struct{}

var IsValidatedCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "is-validated <CONTRACT_NAME>",
		Short:   "checks to see if the contract has passed the last emulated migration",
		Example: `flow migrate is-validated HelloWorld`,
		Args:    cobra.MinimumNArgs(1),
	},
	Flags: &isValidatedflags,
	RunS:  isValidated,
}

func isValidated(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	repoService := github.NewClient(nil).Repositories
	v := validator.NewValidator(repoService, flow.Network(), state, logger)

	contractName := args[0]
	s, ts, err := v.Validate(contractName)
	if err != nil {
		return nil, err
	}

	return validationResult{
		Status:    s,
		Timestamp: *ts,
		Network:   flow.Network().Name,
	}, nil

}

func (v validationResult) String() string {
	status := v.Status

	builder := strings.Builder{}
	builder.WriteString("Last emulated migration report was created at ")
	builder.WriteString(v.Timestamp.Format(time.RFC3339))
	builder.WriteString("\n\n")

	statusBuilder := strings.Builder{}
	emoji := "✅ "
	statusColor := aurora.Green
	if status.IsFailure() {
		emoji = "❌ "
		statusColor = aurora.Red
	}

	statusBuilder.WriteString(util.PrintEmoji(emoji))
	statusBuilder.WriteString("The contract has ")

	if status.IsFailure() {
		statusBuilder.WriteString("FAILED")
	} else {
		statusBuilder.WriteString("PASSED")
	}
	statusBuilder.WriteString(" the last emulated migration")

	statusBuilder.WriteString("\n\n - Account: ")
	statusBuilder.WriteString(status.AccountAddress)
	statusBuilder.WriteString("\n - Contract: ")
	statusBuilder.WriteString(status.ContractName)
	statusBuilder.WriteString("\n - Network: ")
	statusBuilder.WriteString(v.Network)
	statusBuilder.WriteString("\n\n")

	// Write colored status
	builder.WriteString(statusColor(statusBuilder.String()).String())

	if status.Error != "" {
		builder.WriteString(status.Error)
		builder.WriteString("\n")
	}

	if status.IsFailure() {
		builder.WriteString(aurora.Red(">> Please review the error and re-stage the contract to resolve these issues if necessary").String())
		builder.WriteString("\n\n")
	}

	builder.WriteString(moreInformationMessage)
	return builder.String()
}

func (v validationResult) JSON() interface{} {
	return v
}

func (v validationResult) Oneliner() string {
	if v.Status.IsFailure() {
		return util.MessageWithEmojiPrefix("❌", "FAILED")
	}
	return util.MessageWithEmojiPrefix("✅", "FAIlED")
}

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

// Package main implements the entry point for the Flow CLI.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/accounts"
	"github.com/onflow/flow-cli/internal/blocks"
	"github.com/onflow/flow-cli/internal/cadence"
	"github.com/onflow/flow-cli/internal/collections"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/config"
	"github.com/onflow/flow-cli/internal/dependencymanager"
	"github.com/onflow/flow-cli/internal/emulator"
	"github.com/onflow/flow-cli/internal/events"
	evm "github.com/onflow/flow-cli/internal/evm"
	"github.com/onflow/flow-cli/internal/keys"
	"github.com/onflow/flow-cli/internal/migrate"
	"github.com/onflow/flow-cli/internal/project"
	"github.com/onflow/flow-cli/internal/quick"
	"github.com/onflow/flow-cli/internal/scripts"
	"github.com/onflow/flow-cli/internal/settings"
	"github.com/onflow/flow-cli/internal/signatures"
	"github.com/onflow/flow-cli/internal/snapshot"
	"github.com/onflow/flow-cli/internal/status"
	"github.com/onflow/flow-cli/internal/super"
	"github.com/onflow/flow-cli/internal/test"
	"github.com/onflow/flow-cli/internal/tools"
	"github.com/onflow/flow-cli/internal/transactions"
	"github.com/onflow/flow-cli/internal/util"
	"github.com/onflow/flow-cli/internal/version"
)

func main() {
	var cmd = &cobra.Command{
		Use:              "flow",
		TraverseChildren: true,
		// Messaging for Cadence 1.0 upgrade
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			outputFlag, _ := cmd.Flags().GetString("output")
			// If output is set to json, do not append any message
			if outputFlag != "json" {

				width := 80
				url := "https://cadence-lang.org/docs/cadence_migration_guide"

				// Function to center text within a given width
				centerText := func(text string, width int) string {
					space := (width - len(text)) / 2
					if space < 0 {
						space = 0
					}
					return fmt.Sprintf("%s%s%s", strings.Repeat(" ", space), text, strings.Repeat(" ", space))
				}

				fmt.Fprintln(os.Stderr, strings.Repeat("+", width))
				fmt.Fprintln(os.Stderr, centerText("⚠ Upgrade to Cadence 1.0", width))
				fmt.Fprintln(os.Stderr, centerText("The Crescendo network upgrade, including Cadence 1.0, is coming soon.", width))
				fmt.Fprintln(os.Stderr, centerText("You may need to update your existing contracts to support this change.", width))
				fmt.Fprintln(os.Stderr, centerText("Please visit our migration guide here:", width))
				fmt.Fprintln(os.Stderr, centerText(url, width))
				fmt.Fprintln(os.Stderr, strings.Repeat("+", width))

			}
		},
	}

	// quick commands
	quick.InitCommand.AddToParent(cmd)
	quick.DeployCommand.AddToParent(cmd)
	quick.RunCommand.AddToParent(cmd)

	// single commands
	status.Command.AddToParent(cmd)
	tools.DevWallet.AddToParent(cmd)
	tools.Flowser.AddToParent(cmd)
	test.TestCommand.AddToParent(cmd)

	// super commands
	super.SetupCommand.AddToParent(cmd)
	super.DevCommand.AddToParent(cmd)

	// structured commands
	cmd.AddCommand(settings.Cmd)
	cmd.AddCommand(cadence.Cmd)
	cmd.AddCommand(version.Cmd)
	cmd.AddCommand(emulator.Cmd)
	cmd.AddCommand(accounts.Cmd)
	cmd.AddCommand(scripts.Cmd)
	cmd.AddCommand(transactions.Cmd)
	cmd.AddCommand(keys.Cmd)
	cmd.AddCommand(events.Cmd)
	cmd.AddCommand(blocks.Cmd)
	cmd.AddCommand(collections.Cmd)
	cmd.AddCommand(project.Cmd)
	cmd.AddCommand(config.Cmd)
	cmd.AddCommand(signatures.Cmd)
	cmd.AddCommand(snapshot.Cmd)
	cmd.AddCommand(super.FlixCmd)
	cmd.AddCommand(super.GenerateCommand)
	cmd.AddCommand(dependencymanager.Cmd)
	cmd.AddCommand(migrate.Cmd)
	cmd.AddCommand(evm.Cmd)

	command.InitFlags(cmd)
	cmd.AddGroup(&cobra.Group{
		ID:    "super",
		Title: "🔥 Super Commands",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "resources",
		Title: "📦 Flow Entities",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "interactions",
		Title: "💬 Flow Interactions",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "tools",
		Title: "🔨 Flow Tools",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "project",
		Title: "🏄 Flow Project",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "security",
		Title: "🔒 Flow Security",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "manager",
		Title: "🔗 Dependency Manager",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "migrate",
		Title: "📦 Migration to 1.0",
	})

	cmd.SetUsageTemplate(command.UsageTemplate)

	// Don't print usage on error
	cmd.SilenceUsage = true
	// Don't print errors on error (we handle them)
	cmd.SilenceErrors = true

	if err := cmd.Execute(); err != nil {
		util.Exit(1, err.Error())
	}
}

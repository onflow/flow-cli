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

package dependencymanager

import (
	"fmt"

	"github.com/onflow/flow-cli/internal/util"

	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
)

var installFlags = Flags{}

var installCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "install",
		Short:   "Install contract and dependencies.",
		Example: "flow dependencies install",
		Args:    cobra.ArbitraryArgs,
	},
	Flags: &installFlags,
	RunS:  install,
}

func install(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	logger.Info(util.MessageWithEmojiPrefix("🔄", "Installing dependencies from flow.json..."))

	installer, err := NewDependencyInstaller(logger, state, true, "", installFlags)
	if err != nil {
		logger.Error(fmt.Sprintf("Error: %v", err))
		return nil, err
	}

	// Add any dependencies passed as arguments
	if args != nil && len(args) > 0 {
		for _, arg := range args {
			if err := installer.AddBySourceString(arg, ""); err != nil {
				logger.Error(fmt.Sprintf("Error adding dependency: %v", err))
				return nil, err
			}
		}
	}

	// Install dependencies
	if err := installer.Install(); err != nil {
		logger.Error(fmt.Sprintf("Error installing dependencies: %v", err))
		return nil, err
	}

	return nil, nil
}

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

package dependencymanager

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
)

type installFlagsCollection struct {
	skipDeployments bool `default:"false" flag:"skip-deployments" info:"Skip adding the dependency to deployments"`
}

var installFlags = installFlagsCollection{}

var installCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "install",
		Short:   "Install contract and dependencies.",
		Example: "flow dependencies install",
	},
	Flags: &installFlags,
	RunS:  install,
}

func install(
	_ []string,
	_ command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (result command.Result, err error) {
	logger.Info("🔄 Installing dependencies from flow.json...")

	installer, err := NewDependencyInstaller(logger, state, installFlags.skipDeployments)
	if err != nil {
		logger.Error(fmt.Sprintf("Error: %v", err))
		return nil, err
	}

	if err := installer.Install(); err != nil {
		logger.Error(fmt.Sprintf("Error: %v", err))
		return nil, err
	}

	logger.Info("✅  Dependency installation complete. Check your flow.json")

	return nil, nil
}

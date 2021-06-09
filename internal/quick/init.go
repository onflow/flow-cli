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

package quick

import (
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/config"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

// TODO(sideninja) workaround - init needed to be copied in order to work else there is flag duplicate error

var initFlag = config.FlagsInit{}

var InitCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "init",
		Short:   "Initialize a new configuration",
		Example: "flow project init",
	},
	Flags: &initFlag,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
		proj *flowkit.Project,
	) (command.Result, error) {
		proj, err := services.Project.Init(
			initFlag.Reset,
			initFlag.Global,
			initFlag.ServiceKeySigAlgo,
			initFlag.ServiceKeyHashAlgo,
			initFlag.ServicePrivateKey,
		)
		if err != nil {
			return nil, err
		}

		return &config.InitResult{Project: proj}, nil
	},
}

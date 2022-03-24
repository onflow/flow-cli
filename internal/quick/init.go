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

package quick

import (
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/config"
)

// TODO(sideninja) workaround - init needed to be copied in order to work else there is flag duplicate error

var InitCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "init",
		Short:   "Initialize a new configuration",
		Example: "flow project init",
		Annotations: map[string]string{
			"HotCommand": "true",
		},
	},
	Flags: &config.InitFlag,
	Run:   config.Initialise, // TODO(sideninja) workaround - init needed to be copied in order to work else there is flag duplicate error
}

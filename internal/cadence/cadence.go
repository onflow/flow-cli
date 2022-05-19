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

package cadence

import (
	"github.com/onflow/cadence/runtime/cmd/execute"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/cadence/languageserver"
)

var Cmd = &cobra.Command{
	Use:   "cadence",
	Short: "Execute Cadence code",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			execute.Execute(args)
		} else {
			execute.RunREPL()
		}
	},
}

func init() {
	Cmd.AddCommand(languageserver.Cmd)
}

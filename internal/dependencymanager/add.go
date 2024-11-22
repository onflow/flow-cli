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

	"github.com/onflow/flowkit/v2"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/util"

	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
)

var addCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:        "add",
		Short:      "This command has been deprecated.",
		Long:       "The 'add' command has been deprecated. Please use the 'install' command instead.",
		Deprecated: "This command is deprecated. Use 'install' to manage dependencies.",
	},
	RunS:  add,
	Flags: &struct{}{},
}

func add(
	_ []string,
	_ command.GlobalFlags,
	logger output.Logger,
	_ flowkit.Services,
	_ *flowkit.State,
) (command.Result, error) {
	logger.Info(fmt.Sprintf("%s The 'add' command has been deprecated. Please use 'install' instead.", util.PrintEmoji("⚠️")))
	return nil, nil
}

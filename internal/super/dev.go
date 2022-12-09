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

package super

import (
	"fmt"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/config"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var DevCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "dev",
		Short:   "Monitor your project during development", // todo improve
		Args:    cobra.ExactArgs(0),
		Example: "flow dev",
	},
	Flags: &config.InitFlag,
	RunS:  dev,
}

func dev(
	args []string,
	readerWriter flowkit.ReaderWriter,
	globalFlags command.GlobalFlags,
	services *services.Services,
	state *flowkit.State,
) (command.Result, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	proj := projectFiles{projectDir: filepath.Join(dir, "cadence")}
	fmt.Println(proj.Contracts())

	return nil, nil
}

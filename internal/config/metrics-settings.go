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

package config

import (
	"fmt"

	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/onflow/flow-cli/pkg/flowkit/util"

	"github.com/onflow/flow-cli/internal/command"

	"github.com/spf13/cobra"
)

type flagsCommandMetrics struct{}

var commandMetricsFlags = flagsCommandMetrics{}

var MetricsSettings = &command.Command{
	Cmd: &cobra.Command{
		Use:     "metrics",
		Short:   "Configure command usage metrics settings",
		Example: "flow config metrics disable",
		Args:    cobra.ExactArgs(1),
	},
	Flags: &commandMetricsFlags,
	Run:   handleMetricsSettings,
}

func handleMetricsSettings(
	args []string,
	loader flowkit.ReaderWriter,
	_ command.GlobalFlags,
	_ *services.Services,
) (command.Result, error) {
	enabled := args[0] == "enable"
	configPath, err := util.AddToConfig(loader, enabled)
	if err != nil {
		return nil, err
	}

	output := fmt.Sprintf("Metrics tracking is %sd and its settings is stored in %s \n", args[0], configPath) +
		"Please note that you will also need to opt out of metrics tracking on other devices that you also use flow-cli on \n"

	return &Result{output}, nil
}

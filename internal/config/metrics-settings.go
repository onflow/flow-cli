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
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/onflow/flow-cli/pkg/flowkit/util"
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
	_ flowkit.ReaderWriter,
	_ command.GlobalFlags,
	_ *services.Services,
) (command.Result, error) {
	if args[0] == "disable" {
		err := util.SetUserMetricsSettings(false)
		if err != nil {
			return nil, err
		}
	} else if args[0] == "enable" {
		err := util.SetUserMetricsSettings(true)
		if err != nil {
			return nil, err
		}
	}

	return &Result{
		fmt.Sprintf("flow cli commands tracking %sd", args[0]),
	}, nil
}

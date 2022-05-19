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

package command

import (
	"fmt"

	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/util"
)

// GlobalFlags contains all global flags definitions.
type GlobalFlags struct {
	Filter         string
	Format         string
	Save           string
	Host           string
	HostNetworkKey string
	Log            string
	Network        string
	Yes            bool
	ConfigPaths    []string
}

// Flags initialized to default values.
var Flags = GlobalFlags{
	Filter:         "",
	Format:         formatText,
	Save:           "",
	Host:           "",
	HostNetworkKey: "",
	Network:        config.DefaultEmulatorNetwork().Name,
	Log:            logLevelInfo,
	Yes:            false,
	ConfigPaths:    config.DefaultPaths(),
}

// InitFlags init all the global persistent flags.
func InitFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(
		&Flags.Filter,
		"filter",
		"x",
		Flags.Filter,
		"Filter result values by property name",
	)

	cmd.PersistentFlags().StringVarP(
		&Flags.Host,
		"host",
		"",
		Flags.Host,
		"Flow Access API host address",
	)

	cmd.PersistentFlags().StringVarP(
		&Flags.HostNetworkKey,
		"network-key",
		"",
		Flags.HostNetworkKey,
		"Flow Access API host network key for secure client connections",
	)

	cmd.PersistentFlags().StringVarP(
		&Flags.Format,
		"output",
		"o",
		Flags.Format,
		"Output format, options: \"text\", \"json\", \"inline\"",
	)

	cmd.PersistentFlags().StringVarP(
		&Flags.Save,
		"save",
		"s",
		Flags.Save,
		"Save result to a filename",
	)

	cmd.PersistentFlags().StringVarP(
		&Flags.Log,
		"log",
		"l",
		Flags.Log,
		"Log level, options: \"debug\", \"info\", \"error\", \"none\"",
	)

	cmd.PersistentFlags().StringSliceVarP(
		&Flags.ConfigPaths,
		"config-path",
		"f",
		Flags.ConfigPaths,
		"Path to flow configuration file",
	)

	cmd.PersistentFlags().StringVarP(
		&Flags.Network,
		"network",
		"n",
		Flags.Network,
		"Network from configuration file",
	)

	cmd.PersistentFlags().BoolVarP(
		&Flags.Yes,
		"yes",
		"y",
		Flags.Yes,
		"Approve any prompts",
	)
}

// bindFlags bind all the flags needed.
func bindFlags(command Command) {
	err := sconfig.New(command.Flags).
		FromEnvironment(util.EnvPrefix).
		BindFlags(command.Cmd.PersistentFlags()).
		Parse()
	if err != nil {
		fmt.Println(err)
	}
}

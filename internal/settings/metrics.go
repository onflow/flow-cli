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

package settings

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	settings "github.com/onflow/flow-cli/internal/settings/globalSettings"
)

var MetricsSettings = &cobra.Command{
	Use:       "metrics",
	Short:     "Configure command usage metrics settings",
	Example:   "flow config metrics disable \nflow config metrics enable",
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"enable", "disable"},
	RunE:      handleMetricsSettings,
}

// handleMetricsSettings sets global settings for metrics
func handleMetricsSettings(
	_ *cobra.Command,
	args []string,
) error {
	if args[0] == "enable" {
		settings.GlobalSettings.MetricsEnabled = true
	} else if args[0] == "disable" {
		settings.GlobalSettings.MetricsEnabled = false
	} else {
		return errors.New("Invalid metrics argument '" + args[0] + "'")
	}

	fmt.Println(fmt.Sprintf(
		"Command usage tracking is %sd. Setting were updated in %s \n",
		args[0],
		settings.GlobalSettings.GetFileName()))

	return nil
}

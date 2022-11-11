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
	"fmt"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

const enable = "enable"
const disable = "disable"

var MetricsSettings = &cobra.Command{
	Use:       "metrics",
	Short:     "Configure command usage metrics settings",
	Example:   "flow settings metrics disable \nflow settings metrics enable",
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	ValidArgs: []string{enable, disable},
	RunE:      handleMetricsSettings,
}

// handleMetricsSettings sets global settings for metrics
func handleMetricsSettings(
	_ *cobra.Command,
	args []string,
) error {
	enabled := args[0] == enable
	if err := Set(MetricsEnabled, enabled); err != nil {
		return errors.Wrap(err, "failed to update metrics setting")
	}

	fmt.Println(fmt.Sprintf(
		"Command usage tracking is %sd. Setting were updated in %s \n",
		args[0],
		FileName()))

	return nil
}

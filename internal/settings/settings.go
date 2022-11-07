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
	settings "github.com/onflow/flow-cli/internal/settings/globalSettings"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:               "settings",
	Short:             "Manage persisted global settings",
	PersistentPostRun: handleSettingsChanged,
	TraverseChildren:  true,
}

func init() {
	settings.GlobalSettings.Import()
	Cmd.AddCommand(MetricsSettings)
}

func handleSettingsChanged(_ *cobra.Command, _ []string) {
	settings.GlobalSettings.Export()
}

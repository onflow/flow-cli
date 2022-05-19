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

package version

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/build"
)

var Cmd = &cobra.Command{
	Use:   "version",
	Short: "View version and commit information",
	Run: func(cmd *cobra.Command, args []string) {
		semver := build.Semver()
		commit := build.Commit()

		// Print version/commit strings if they are known
		if build.IsDefined(semver) {
			fmt.Printf("Version: %s\n", semver)
		}
		if build.IsDefined(commit) {
			fmt.Printf("Commit: %s\n", commit)
		}
		// If no version info is known print a message to indicate this.
		if !build.IsDefined(semver) && !build.IsDefined(commit) {
			fmt.Printf("Version information unknown!\n")
		}
	},
}

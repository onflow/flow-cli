/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
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
	"github.com/spf13/cobra"
)

var ViewCmd = &cobra.Command{
	Use:              "view <account|contract|emulator|network|deployment>",
	Short:            "View the resources in the configuration",
	Example:          "flow config view account \nflow config view account <accountname> \nflow config view contract \nflow config view contract <contractname> \nflow config view emulator \nflow config view emulator <emulatorname> \nflow config view network \nflow config view network <networkname> \nflow config view deployment \nflow config view deployment <networkname>",
	Args:             cobra.ExactArgs(1),
	TraverseChildren: true,
}

func init() {
	ViewAccountCommand.AddToParent(ViewCmd)
	ViewEmulatorCommand.AddToParent(ViewCmd)
	ViewContractCommand.AddToParent(ViewCmd)
	ViewDeploymentCommand.AddToParent(ViewCmd)
	ViewNetworkCommand.AddToParent(ViewCmd)

}

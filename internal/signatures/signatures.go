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

package signatures

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "signatures",
	Short:            "Signature verification and creation",
	TraverseChildren: true,
	GroupID:          "security",
}

func init() {
	generateCommand.AddToParent(Cmd)
	verifyCommand.AddToParent(Cmd)
}

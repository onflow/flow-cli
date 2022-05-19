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

package snapshot

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "snapshot",
	Short:            "Utilities to download the latest finalized protocol state snapshot",
	TraverseChildren: true,
}

func init() {
	SaveCommand.AddToParent(Cmd)
}

// SaveResult represents the result of the snapshot save command.
type SaveResult struct {
	OutputPath string
}

func (r *SaveResult) JSON() interface{} {
	return map[string]string{"path": r.OutputPath}
}

func (r *SaveResult) String() string {
	return fmt.Sprintf("snapshot saved: %s", r.OutputPath)
}

func (r *SaveResult) Oneliner() string {
	return fmt.Sprintf("snapshot saved: %s", r.OutputPath)
}

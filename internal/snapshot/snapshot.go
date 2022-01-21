/*
 * Flow CLI
 *
 * Copyright 2019-2022 Dapper Labs, Inc.
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
	Short:            "Utility to download the latest finalized protocol state snapshot",
	TraverseChildren: true,
}

func init() {
	DownloadCommand.AddToParent(Cmd)
}

// DownloadResult represent result snapshot download command.
type DownloadResult struct {
	OutputPath string
}

func (r *DownloadResult) JSON() interface{} {
	return map[string]string{"output-path": r.OutputPath}
}

func (r *DownloadResult) String() string {
	return fmt.Sprintf("output path: %s", r.OutputPath)
}

func (r *DownloadResult) Oneliner() string {
	return fmt.Sprintf("output path: %s", r.OutputPath)
}

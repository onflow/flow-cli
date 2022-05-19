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
	"path/filepath"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"

	"github.com/spf13/cobra"
)

var SaveCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "save",
		Short:   "Get the latest finalized protocol snapshot",
		Example: "flow snapshot save /tmp/snapshot.json",
		Args:    cobra.ExactArgs(1),
	},
	Flags: &struct{}{},
	Run:   save,
}

func save(
	args []string,
	readerWriter flowkit.ReaderWriter,
	_ command.GlobalFlags,
	services *services.Services,
) (command.Result, error) {
	fileName := args[0]

	snapshotBytes, err := services.Snapshot.GetLatestProtocolStateSnapshot()
	if err != nil {
		return nil, err
	}

	outputPath, err := filepath.Abs(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute output path for protocol snapshot")
	}

	err = readerWriter.WriteFile(outputPath, snapshotBytes, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write protocol snapshot file to %s: %w", outputPath, err)
	}

	return &SaveResult{OutputPath: outputPath}, nil
}

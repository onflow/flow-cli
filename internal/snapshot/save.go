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

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
)

var saveCommand = &command.Command{
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
	_ command.GlobalFlags,
	logger output.Logger,
	writer flowkit.ReaderWriter,
	flow flowkit.Services,
) (command.Result, error) {
	fileName := args[0]

	logger.StartProgress("Downloading protocol snapshot...")
	if !flow.Gateway().SecureConnection() {
		logger.Info(fmt.Sprintf("%s warning: using insecure client connection to download snapshot, you should use a secure network configuration...", output.WarningEmoji()))
	}

	snapshotBytes, err := flow.Gateway().GetLatestProtocolStateSnapshot()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest finalized protocol snapshot from gateway: %w", err)
	}

	logger.StopProgress()

	outputPath, err := filepath.Abs(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute output path for protocol snapshot")
	}

	err = writer.WriteFile(outputPath, snapshotBytes, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write protocol snapshot file to %s: %w", outputPath, err)
	}

	return &saveResult{OutputPath: outputPath}, nil
}

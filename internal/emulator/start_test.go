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

package emulator

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flowkit/v2"
)

func Test_PersistentPreRunE_ForkFlag(t *testing.T) {
	// Create a temporary directory for flow.json
	tempDir := t.TempDir()
	flowJSONPath := filepath.Join(tempDir, "flow.json")

	// Create a sample flow.json with networks
	flowJSONContent := `{
		"networks": {
			"mainnet": "access.mainnet.nodes.onflow.org:9000",
			"testnet": "access.devnet.nodes.onflow.org:9000"
		}
	}`
	err := os.WriteFile(flowJSONPath, []byte(flowJSONContent), 0644)
	require.NoError(t, err)

	// Create a command with the fork flag
	cmd := &cobra.Command{}
	cmd.Flags().String("fork", "", "")
	cmd.Flags().String("fork-host", "", "")
	cmd.Flags().Bool("skip-tx-validation", false, "")

	// Set the PersistentPreRunE function (copied from init)
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		fh, err := cmd.Flags().GetString("fork-host")
		if err != nil {
			return err
		}
		if fh != "" {
			return nil
		}
		forkOpt, err := cmd.Flags().GetString("fork")
		if err != nil {
			return err
		}
		if forkOpt == "" {
			return nil
		}
		loader := &afero.Afero{Fs: afero.NewOsFs()}
		state, err := flowkit.Load([]string{flowJSONPath}, loader) // Use the test path
		if err != nil {
			return fmt.Errorf("failed to load flow.json: %w", err)
		}

		// Resolve network endpoint from flow.json
		network, err := state.Networks().ByName(forkOpt)
		if err != nil {
			return fmt.Errorf("network %q not found in flow.json", forkOpt)
		}
		host := network.Host
		if host == "" {
			return fmt.Errorf("network %q has no host configured", forkOpt)
		}

		// Set fork-host flag
		if err := cmd.Flags().Set("fork-host", host); err != nil {
			return err
		}

		// Automatically disable signature validation when forking
		if err := cmd.Flags().Set("skip-tx-validation", "true"); err != nil {
			return err
		}

		return nil
	}

	// Test case 1: Fork with mainnet
	cmd.Flags().Set("fork", "mainnet")
	err = cmd.PersistentPreRunE(cmd, []string{})
	assert.NoError(t, err)

	forkHost, _ := cmd.Flags().GetString("fork-host")
	assert.Equal(t, "access.mainnet.nodes.onflow.org:9000", forkHost)

	skipValidation, _ := cmd.Flags().GetBool("skip-tx-validation")
	assert.True(t, skipValidation)

	// Reset flags for next test
	cmd.Flags().Set("fork-host", "")
	cmd.Flags().Set("skip-tx-validation", "false")

	// Test case 2: Fork with testnet
	cmd.Flags().Set("fork", "testnet")
	err = cmd.PersistentPreRunE(cmd, []string{})
	assert.NoError(t, err)

	forkHost, _ = cmd.Flags().GetString("fork-host")
	assert.Equal(t, "access.devnet.nodes.onflow.org:9000", forkHost)

	skipValidation, _ = cmd.Flags().GetBool("skip-tx-validation")
	assert.True(t, skipValidation)
}

func Test_PersistentPreRunE_ForkFlag_Errors(t *testing.T) {
	// Create a temporary directory for flow.json
	tempDir := t.TempDir()
	flowJSONPath := filepath.Join(tempDir, "flow.json")

	// Create a flow.json without the network
	flowJSONContent := `{
		"networks": {
			"mainnet": "access.mainnet.nodes.onflow.org:9000"
		}
	}`
	err := os.WriteFile(flowJSONPath, []byte(flowJSONContent), 0644)
	require.NoError(t, err)

	// Create a command
	cmd := &cobra.Command{}
	cmd.Flags().String("fork", "", "")
	cmd.Flags().String("fork-host", "", "")

	// Set the PersistentPreRunE function
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		fh, err := cmd.Flags().GetString("fork-host")
		if err != nil {
			return err
		}
		if fh != "" {
			return nil
		}
		forkOpt, err := cmd.Flags().GetString("fork")
		if err != nil {
			return err
		}
		if forkOpt == "" {
			return nil
		}
		loader := &afero.Afero{Fs: afero.NewOsFs()}
		state, err := flowkit.Load([]string{flowJSONPath}, loader)
		if err != nil {
			return fmt.Errorf("failed to load flow.json: %w", err)
		}

		// Resolve network endpoint from flow.json
		network, err := state.Networks().ByName(forkOpt)
		if err != nil {
			return fmt.Errorf("network %q not found in flow.json", forkOpt)
		}
		host := network.Host
		if host == "" {
			return fmt.Errorf("network %q has no host configured", forkOpt)
		}

		// Set fork-host flag
		if err := cmd.Flags().Set("fork-host", host); err != nil {
			return err
		}

		// Automatically disable signature validation when forking
		if err := cmd.Flags().Set("skip-tx-validation", "true"); err != nil {
			return err
		}

		return nil
	}

	// Test case: Network not found
	cmd.Flags().Set("fork", "nonexistent")
	err = cmd.PersistentPreRunE(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "network \"nonexistent\" not found")
}

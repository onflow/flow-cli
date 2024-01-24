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

package command_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
)

func TestInitFlags(t *testing.T) {
	cmd := &cobra.Command{}
	command.InitFlags(cmd)

	flags := []struct {
		name     string
		expected string
	}{
		{"filter", command.Flags.Filter},
		{"format", command.Flags.Format},
		{"save", command.Flags.Save},
		{"host", command.Flags.Host},
		{"network-key", command.Flags.HostNetworkKey},
		{"network", command.Flags.Network},
		{"log", command.Flags.Log},
		{"yes", strconv.FormatBool(command.Flags.Yes)},
		{"config-path", fmt.Sprintf("[%s]", strings.Join(command.Flags.ConfigPaths, ","))},
		{"skip-version-check", strconv.FormatBool(command.Flags.SkipVersionCheck)},
	}

	for _, flag := range flags {
		f := cmd.PersistentFlags().Lookup(flag.name)
		if f == nil {
			t.Errorf("Flag %s was not initialized", flag.name)
		} else if f.DefValue != flag.expected {
			t.Errorf("Flag %s was not initialized with correct default value. Value: %s, Expected: %s", flag.name, f.Value.String(), flag.expected)
		}
	}
}

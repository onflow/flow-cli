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
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/build"
	"github.com/onflow/flow-cli/internal/command"
)

type versionCmd struct {
	Version      string
	Commit       string
	Dependencies []debug.Module
}

// Print prints the version information in the given format.
func (c versionCmd) Print(format string) error {
	switch format {
	case command.FormatInline, command.FormatText:
		var txtBuilder strings.Builder
		txtBuilder.WriteString(fmt.Sprintf("Version: %s\n", c.Version))
		txtBuilder.WriteString(fmt.Sprintf("Commit: %s\n", c.Commit))

		txtBuilder.WriteString("\nFlow Package Dependencies \n")
		for _, dep := range c.Dependencies {
			txtBuilder.WriteString(fmt.Sprintf("%s %s\n", dep.Path, dep.Version))
		}

		fmt.Println(txtBuilder.String())

		return nil

	case command.FormatJSON:
		jsonRes, err := c.MarshalJSON()
		if err != nil {
			return err
		}

		fmt.Println(string(jsonRes))

		return nil

	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// MarshalJSON returns the JSON encoding of the cmdPrint.
func (c *versionCmd) MarshalJSON() ([]byte, error) {
	js := struct {
		Version      string `json:"version"`
		Commit       string `json:"commit"`
		Dependencies []struct {
			Package string `json:"package"`
			Version string `json:"version"`
		} `json:"dependencies"`
	}{
		Version: c.Version,
		Commit:  c.Commit,
	}

	for _, dep := range c.Dependencies {
		js.Dependencies = append(js.Dependencies, struct {
			Package string `json:"package"`
			Version string `json:"version"`
		}{
			Package: dep.Path,
			Version: dep.Version,
		})
	}

	return json.Marshal(js)
}

var Cmd = &cobra.Command{
	Use:   "version",
	Short: "View version and commit information",
	RunE: func(cmd *cobra.Command, args []string) error {
		semver := build.Semver()
		commit := build.Commit()

		v := &versionCmd{
			Version: semver,
			Commit:  commit,
		}

		bi, ok := debug.ReadBuildInfo()
		if !ok {
			return fmt.Errorf("failed to read build info")
		}

		// only add dependencies from github.com/onflow
		for _, dep := range bi.Deps {
			if strings.Contains(dep.Path, "github.com/onflow/") {
				v.Dependencies = append(v.Dependencies, *dep)
			}
		}

		if err := v.Print(command.Flags.Format); err != nil {
			return err
		}

		return nil
	},
}

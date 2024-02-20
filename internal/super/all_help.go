/*
 * Flow CLI
 *
 * Copyright 2024 Flow Foundation
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

package super

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var AllHelp = &cobra.Command{
	Use:     "allHelp",
	Short:   "Outputs help for all the CLI commands",
	Example: "flow cheat sheet",
	Run: func(
		c *cobra.Command,
		args []string,
	) {
		root := c.Root()
		r, err := generateCS(root)
		if err != nil {
			fmt.Printf("Error generating cheat sheet: %s", root.Name())
		}
		fmt.Println(r)
	},
}

func generateCS(c *cobra.Command) (string, error) {
	cmd := c.Root()
	// Slice to hold the help texts of all commands
	var helpTexts []string
	helpTexts = append(helpTexts, "```")

	// Generate the help texts
	generateCommandHelpTexts(cmd, &helpTexts)
	helpTexts = append(helpTexts, "```")

	commands := strings.Join(helpTexts, "\n\n------------\n\n")

	return commands, nil

}

// Recursive function to traverse all commands and subcommands,
// capturing the help text for each.
func generateCommandHelpTexts(cmd *cobra.Command, result *[]string) {
	removeGlobalFlags := true
	// version is last command show global flags
	if cmd.Name() == "version" {
		removeGlobalFlags = false
	}
	*result = append(*result, getCommandHelpText(cmd, removeGlobalFlags))

	// Recursively process each subcommand
	for _, subCmd := range cmd.Commands() {
		generateCommandHelpTexts(subCmd, result)
	}
}

// Helper function to execute the help command for a given cobra.Command
// and return the output as a string.
func getCommandHelpText(cmd *cobra.Command, removeGlobalFlags bool) string {
	// Create a new bytes buffer to capture the output
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	if err := cmd.Help(); err != nil {
		fmt.Printf("Error generating help for %s", cmd.Name())
	}
	cmd.SetOut(nil)
	helpText := buf.String()

	if removeGlobalFlags {
		helpText = removeGlobalFlagsSection(helpText)
	}

	return removeANSICodes(helpText)
}

func removeGlobalFlagsSection(helpText string) string {
	// Define the start of the Global Flags section
	startMarker := "Global Flags:\n"
	start := strings.Index(helpText, startMarker)
	if start != -1 {
		// Attempt to find the end of the Global Flags section by looking for the next section or end of text
		// This may need adjustment based on your actual help text formatting.
		end := strings.Index(helpText[start+len(startMarker):], "\n\n")
		if end != -1 {
			// Remove the Global Flags section
			helpText = helpText[:start] + helpText[start+end+len(startMarker)+2:]
		} else {
			// If there's no section after Global Flags, remove starting from the marker to the end of the text
			helpText = helpText[:start]
		}
	}
	return helpText
}

func removeANSICodes(input string) string {
	re := regexp.MustCompile(`\x1B\[[0-9;]*[a-zA-Z]`)
	return re.ReplaceAllString(input, "")
}

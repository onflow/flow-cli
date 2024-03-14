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
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var ansiRegex = regexp.MustCompile(`\x1B\[[0-9;]*[a-zA-Z]`)

var AllHelp = &cobra.Command{
	Use:     "all-help",
	Short:   "Outputs help for all the CLI commands",
	Example: "flow cheatsheet",
	Run: func(
		c *cobra.Command,
		args []string,
	) {
		root := c.Root()
		r, err := generateCS(root)
		if err != nil {
			fmt.Printf("Error generating cheatsheet: %s", root.Name())
		}
		fmt.Println(r)
	},
}

func generateCS(c *cobra.Command) (string, error) {
	cmd := c.Root()
	var helpTexts strings.Builder

	// Generate the help texts
	generateCommandHelpTexts(cmd, &helpTexts)
	globalFlags := getGlobalFlagsHelp(cmd)
	helpTexts.WriteString("\n\n---------------\n\n")
	helpTexts.WriteString(globalFlags)

	return helpTexts.String(), nil

}

// Recursive function to traverse all commands and subcommands,
// capturing the help text for each.
func generateCommandHelpTexts(cmd *cobra.Command, result *strings.Builder) {

	result.WriteString("\n\n---------------\n\n")
	result.WriteString(getCommandHelpText(cmd))

	// Recursively process each subcommand
	for _, subCmd := range cmd.Commands() {
		generateCommandHelpTexts(subCmd, result)
	}
}

// Helper function to execute the help command for a given cobra.Command
// and return the output as a string.
func getCommandHelpText(cmd *cobra.Command) string {
	var sb strings.Builder
	cmd.SetOut(&sb)

	if err := cmd.Help(); err != nil {
		fmt.Printf("Error generating help for %s", cmd.Name())
	}

	cmd.SetOut(nil)
	helpText := sb.String()
	helpText = removeGlobalFlagsSection(helpText)
	return ansiRegex.ReplaceAllString(helpText, "")
}

func removeGlobalFlagsSection(helpText string) string {
	// Define the start of the Global Flags section
	const startMarker = "Global Flags:\n"
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

func getGlobalFlagsHelp(rootCmd *cobra.Command) string {
	// Initialize a strings.Builder to capture the output
	var globalFlagsHelp strings.Builder

	// Create a dummy command to capture global flags
	var dummyCmd cobra.Command
	// Assign the persistent flags of the rootCmd to the dummy command
	dummyCmd.PersistentFlags().AddFlagSet(rootCmd.PersistentFlags())

	// Set the output of the dummy command's flag set to the strings.Builder
	dummyCmd.SetOut(&globalFlagsHelp)
	_ = dummyCmd.Usage()
	helpText := globalFlagsHelp.String()
	helpText = strings.Replace(helpText, "Usage:\n", "", 1)
	helpText = strings.Replace(helpText, "Flags:", "Global Flags:", 1)

	return helpText
}

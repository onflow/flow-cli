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

package command

import (
	"fmt"
	"text/template"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/common/branding"
)

var (
	// Header styles
	logoStyle = lipgloss.NewStyle().
			Foreground(branding.FlowGreen).
			Bold(true)

	welcomeStyle = lipgloss.NewStyle().
			Foreground(branding.PurpleText).
			Bold(true)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(branding.GrayText).
			Italic(true).
			MarginLeft(2)

	// Command group styles
	groupTitleStyle = lipgloss.NewStyle().
			Foreground(branding.FlowGreen).
			Bold(true)

	commandNameStyle = lipgloss.NewStyle().
				Foreground(branding.PurpleText).
				Bold(true)

	commandDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("7"))

	// Section styles
	sectionTitleStyle = lipgloss.NewStyle().
				Foreground(branding.FlowGreen).
				Bold(true)

	// Footer style
	footerStyle = lipgloss.NewStyle().
			Foreground(branding.GrayText).
			Italic(true)
)

// Template functions for styling
var templateFuncs = template.FuncMap{
	"styleFlowHeader": func() string {
		logo := logoStyle.Render(branding.FlowASCII)
		welcome := welcomeStyle.Render("👋 Welcome Flow developer!")
		subtitle := subtitleStyle.Render("If you are starting a new flow project use our super commands, start by running 'flow init'.")
		return fmt.Sprintf("%s\n%s\n%s\n", logo, welcome, subtitle)
	},
	"styleGroupTitle": func(title string) string {
		return groupTitleStyle.Render(title)
	},
	"styleCommandName": func(name string) string {
		return commandNameStyle.Render(name)
	},
	"styleCommandDesc": func(desc string) string {
		return commandDescStyle.Render(desc)
	},
	"styleSectionTitle": func(title string) string {
		return sectionTitleStyle.Render(title)
	},
	"styleFooter": func(text string) string {
		return footerStyle.Render(text)
	},
}

func InitTemplateFunc(cmd *cobra.Command) {
	for name, fn := range templateFuncs {
		cobra.AddTemplateFunc(name, fn)
	}
}

var UsageTemplate = `{{if (eq .Name "flow")}}{{styleFlowHeader}}{{else}}Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

{{styleSectionTitle "Available Commands:"}}{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{styleCommandName (rpad .Name .NamePadding)}} {{styleCommandDesc .Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{styleGroupTitle .Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{styleCommandName (rpad .Name .NamePadding)}} {{styleCommandDesc .Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

{{styleSectionTitle "Additional Commands:"}}{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{styleCommandName (rpad .Name .NamePadding)}} {{styleCommandDesc .Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

{{styleSectionTitle "Flags:"}}
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

{{styleSectionTitle "Global Flags:"}}
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

{{styleSectionTitle "Additional help topics:"}}{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{styleCommandName (rpad .CommandPath .CommandPathPadding)}} {{styleCommandDesc .Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

{{styleFooter (printf "Use \"%s [command] --help\" for more information about a command." .CommandPath)}}{{end}}`

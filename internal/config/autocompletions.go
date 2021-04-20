package config

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/spf13/cobra"
)

var CompletionCmd = &cobra.Command{
	Use:                   "setup-completions [powershell]",
	Short:                 "Setup command autocompletion",
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"powershell"},
	Args:                  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		shell := ""
		if len(args) == 1 {
			shell = args[0]
		}

		if shell == "powershell" {
			cmd.Root().GenPowerShellCompletion(os.Stdout)
		} else {
			shell, shellOS := output.AutocompletionPrompt()

			if shell == "bash" && shellOS == "MacOS" {
				cmd.Root().GenBashCompletionFile("/usr/local/etc/bash_completion.d/flow")

				fmt.Printf("Flow command completions installed in: /usr/local/etc/bash_completion.d/flow\n")
				fmt.Printf("You will need to start a new shell for this setup to take effect.\n\n")
			} else if shell == "bash" && shellOS == "Linux" {
				cmd.Root().GenBashCompletionFile("/etc/bash_completion.d/flow")

				fmt.Printf("Flow command completions installed in: /etc/bash_completion.d/flow\n")
				fmt.Printf("You will need to start a new shell for this setup to take effect.\n\n")
			} else if shell == "zsh" {
				c := exec.Command("zsh", "-c ", `echo -n ${fpath[1]}`)
				path, _ := c.Output()
				cmd.Root().GenZshCompletionFile(fmt.Sprintf("%s/_flow", path))

				fmt.Printf("Flow command completions installed in: '%s/_flow'\n", path)
				fmt.Printf("You will need to start a new shell for this setup to take effect.\n\n")
			}
		}
	},
}

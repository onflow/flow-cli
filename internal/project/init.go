package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
)

type flagsInit struct{}

var initFlag = flagsInit{}

var InitCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:   "init",
		Short: "⚠️  No longer supported: use 'flow init' command",
	},
	Flags: &initFlag,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {
		return nil, fmt.Errorf("⚠️  No longer supported: use 'flow init' command.")
	},
}

// InitResult result structure
type InitResult struct {
	*project.Project
}

// JSON convert result to JSON
func (r *InitResult) JSON() interface{} {
	account, _ := r.Project.EmulatorServiceAccount()
	result := make(map[string]string)

	result["serviceAccount"] = account.Address().Hex()

	return result
}

// String convert result to string
func (r *InitResult) String() string {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, '\t', tabwriter.AlignRight)
	account, _ := r.Project.EmulatorServiceAccount()

	fmt.Fprintf(writer, "Configuration initialized\n")
	fmt.Fprintf(writer, "Service account: %s\n\n", util.Bold("0x"+account.Address().String()))
	fmt.Fprintf(writer,
		"Start emulator by running: %s \nReset configuration using: %s\n",
		util.Bold("'flow emulator'"),
		util.Bold("'flow project init --reset'"),
	)

	writer.Flush()
	return b.String()
}

// Oneliner show result as one liner grep friendly
func (r *InitResult) Oneliner() string {
	return ""
}

package config

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"github.com/onflow/flow-cli/pkg/flowcli/util"
)

// todo implement top level command config

// InitResult result structure
type InitResult struct {
	*project.Project
}

// JSON convert result to JSON
func (r *InitResult) JSON() interface{} {
	return r
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
		util.Bold("'flow init --reset'"),
	)

	writer.Flush()
	return b.String()
}

// Oneliner show result as one liner grep friendly
func (r *InitResult) Oneliner() string {
	return ""
}

package project

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/onflow/flow-cli/cmd"
	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/sharedlib/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type flagsInit struct {
	ServicePrivateKey  string `flag:"service-priv-key" info:"Service account private key"`
	ServiceKeySigAlgo  string `default:"ECDSA_P256" flag:"service-sig-algo" info:"Service account key signature algorithm"`
	ServiceKeyHashAlgo string `default:"SHA3_256" flag:"service-hash-algo" info:"Service account key hash algorithm"`
	Reset              bool   `default:"false" flag:"reset" info:"Reset flow.json config file"`
}

type cmdInit struct {
	cmd   *cobra.Command
	flags flagsInit
}

// NewExecuteScriptCmd creates new script command
func NewInitCmd() cmd.Command {
	return &cmdInit{
		cmd: &cobra.Command{
			Use:   "init",
			Short: "Initialize a new account profile",
		},
	}
}

// Run script command
func (s *cmdInit) Run(
	cmd *cobra.Command,
	args []string,
	project *cli.Project,
	services *services.Services,
) (cmd.Result, error) {
	project, err := services.Project.Init(
		s.flags.Reset,
		s.flags.ServiceKeySigAlgo,
		s.flags.ServiceKeyHashAlgo,
		s.flags.ServicePrivateKey,
	)
	return &InitResult{project}, err
}

// GetFlags for script
func (s *cmdInit) GetFlags() *sconfig.Config {
	return sconfig.New(&s.flags)
}

// GetCmd get command
func (s *cmdInit) GetCmd() *cobra.Command {
	return s.cmd
}

type InitResult struct {
	*cli.Project
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
	fmt.Fprintf(writer, "Service account: %s\n\n", cli.Bold("0x"+account.Address().String()))
	fmt.Fprintf(writer,
		"Start emulator by runing: %s \nReset configuration using: %s.\n",
		cli.Bold("flow emulator start"),
		cli.Bold("flow init --rest"),
	)

	writer.Flush()
	return b.String()
}

// Oneliner show result as one liner grep friendly
func (r *InitResult) Oneliner() string {
	return ""
}

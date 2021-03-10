package keys

import (
	"github.com/onflow/flow-cli/cmd"
	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/sharedlib/services"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"
)

type flagsGenerate struct {
	Seed     string `flag:"seed,s" info:"Deterministic seed phrase"`
	SigAlgo  string `default:"ECDSA_P256" flag:"algo,a" info:"Signature algorithm"`
	HashAlgo string `flag:"hashalgo" info:"hash algorithm for the key"`
	KeyIndex int    `flag:"index" info:"index of the key on the account"`
}

type cmdGenerate struct {
	cmd   *cobra.Command
	flags flagsGenerate
}

// NewGenerateCmd return new command
func NewGenerateCmd() cmd.Command {
	return &cmdGenerate{
		cmd: &cobra.Command{
			Use:   "generate",
			Short: "Generate a new key-pair",
		},
	}
}

func (a *cmdGenerate) Run(
	cmd *cobra.Command,
	args []string,
	project *cli.Project,
	services *services.Services,
) (cmd.Result, error) {

	keys, err := services.Keys.Generate(a.flags.Seed, a.flags.SigAlgo)
	return &KeyResult{keys}, err
}

func (a *cmdGenerate) GetFlags() *sconfig.Config {
	return sconfig.New(&a.flags)
}

func (a *cmdGenerate) GetCmd() *cobra.Command {
	return a.cmd
}

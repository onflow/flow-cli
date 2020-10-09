package status

import (
	"log"

	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"

	cli "github.com/dapperlabs/flow-cli/flow"
)

type Config struct {
	Host   string `default:"127.0.0.1:3569" flag:"host" info:"Flow Access API host address"`
	Sealed bool   `default:"true" flag:"sealed" info:"Wait for a sealed result"`
}

var conf Config

var Cmd = &cobra.Command{
	Use:   "status <tx_id>",
	Short: "Get the transaction status",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cli.GetTransactionResult(conf.Host, args[0], conf.Sealed)
	},
}

func init() {
	initConfig()
}

func initConfig() {
	err := sconfig.New(&conf).
		FromEnvironment(cli.EnvPrefix).
		BindFlags(Cmd.PersistentFlags()).
		Parse()
	if err != nil {
		log.Fatal(err)
	}
}

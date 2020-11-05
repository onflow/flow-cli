package execute

import (
	"io/ioutil"
	"log"

	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"

	cli "github.com/dapperlabs/flow-cli/flow"
)

type Config struct {
	Host string `flag:"host" info:"Flow Access API host address"`
}

var conf Config

var Cmd = &cobra.Command{
	Use:   "execute <script.cdc>",
	Short: "Execute a script",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		code, err := ioutil.ReadFile(args[0])
		if err != nil {
			cli.Exitf(1, "Failed to read script from %s", args[0])
		}
		projectConf := new(cli.Config)
		if conf.Host == "" {
			projectConf = cli.LoadConfig()
		}
		cli.ExecuteScript(projectConf.HostWithOverride(conf.Host), code)
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

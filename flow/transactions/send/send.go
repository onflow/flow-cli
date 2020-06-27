package send

import (
	"io/ioutil"
	"log"

	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"

	cli "github.com/dapperlabs/flow-cli/flow"
)

type Config struct {
	Signer  string `default:"service" flag:"signer,s"`
	Code    string `flag:"code,c"`
	Host    string `default:"127.0.0.1:3569" flag:"host" info:"Flow Observation API host address"`
	Results bool   `default:"false" flag:"results" info:"Wether or not to wait for the results of the transaction"`
}

var conf Config

var Cmd = &cobra.Command{
	Use:   "send",
	Short: "Send a transaction",
	Run: func(cmd *cobra.Command, args []string) {
		projectConf := cli.LoadConfig()

		signerAccount := projectConf.Accounts[conf.Signer]

		var (
			code []byte
			err  error
		)

		if conf.Code != "" {
			code, err = ioutil.ReadFile(conf.Code)
			if err != nil {
				cli.Exitf(1, "Failed to read transaction script from %s", conf.Code)
			}
		}

		cli.SendTransaction(conf.Host, signerAccount, code, conf.Results)
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

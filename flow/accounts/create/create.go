package create

import (
	"io/ioutil"
	"log"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/templates"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"

	cli "github.com/dapperlabs/flow-cli/flow"
)

type Config struct {
	Signer string   `default:"service" flag:"signer,s"`
	Keys   []string `flag:"key,k" info:"public keys to attach to account"`
	Code   string   `flag:"code,c" info:"path to a file containing code for the account"`
	Host   string   `default:"127.0.0.1:3569" flag:"host" info:"Flow Observation API host address"`
}

var conf Config

var Cmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new account",
	Run: func(cmd *cobra.Command, args []string) {
		projectConf := cli.LoadConfig()

		signerAccount := projectConf.Accounts[conf.Signer]

		accountKeys := make([]*flow.AccountKey, len(conf.Keys))

		for i, publicKeyHex := range conf.Keys {
			publicKey := cli.MustDecodePublicKeyHex(cli.DefaultSigAlgo, publicKeyHex)
			accountKeys[i] = &flow.AccountKey{
				PublicKey: publicKey,
				SigAlgo:   cli.DefaultSigAlgo,
				HashAlgo:  cli.DefaultHashAlgo,
				Weight:    flow.AccountKeyWeightThreshold,
			}
		}

		var (
			code []byte
			err  error
		)

		if conf.Code != "" {
			code, err = ioutil.ReadFile(conf.Code)
			if err != nil {
				cli.Exitf(1, "Failed to read Cadence code from %s", conf.Code)
			}
		}

		script, err := templates.CreateAccount(accountKeys, code)
		if err != nil {
			cli.Exit(1, "Failed to generate transaction script")
		}

		cli.SendTransaction(conf.Host, signerAccount, script)
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

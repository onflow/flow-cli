package create

import (
	"io/ioutil"
	"log"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/templates"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"

	cli "github.com/dapperlabs/flow-cli/flow"
)

type Config struct {
	Signer   string   `default:"service" flag:"signer,s"`
	Keys     []string `flag:"key,k" info:"public keys to attach to account"`
	SigAlgo  string   `default:"ECDSA_P256" flag:"sig-algo" info:"signature algorithm used to generate the keys"`
	HashAlgo string   `default:"SHA3_256" flag:"hash-algo" info:"hash used for the digest"`
	Code     string   `flag:"code,c" info:"path to a file containing code for the account"`
	Host     string   `flag:"host" info:"Flow Observation API host address"`
	Results  bool     `default:"false" flag:"results" info:"Wether or not to wait for the results of the transaction"`
}

var conf Config

var Cmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new account",
	Run: func(cmd *cobra.Command, args []string) {
		projectConf := cli.LoadConfig()

		signerAccount := projectConf.Accounts[conf.Signer]

		accountKeys := make([]*flow.AccountKey, len(conf.Keys))

		sigAlgo := crypto.StringToSignatureAlgorithm(conf.SigAlgo)
		if sigAlgo == crypto.UnknownSignatureAlgorithm {
			cli.Exitf(1, "Failed to determine signature algorithm from %s", conf.SigAlgo)
		}
		hashAlgo := crypto.StringToHashAlgorithm(conf.HashAlgo)
		if hashAlgo == crypto.UnknownHashAlgorithm {
			cli.Exitf(1, "Failed to determine hash algorithm from %s", conf.HashAlgo)
		}

		for i, publicKeyHex := range conf.Keys {
			publicKey := cli.MustDecodePublicKeyHex(cli.DefaultSigAlgo, publicKeyHex)
			accountKeys[i] = &flow.AccountKey{
				PublicKey: publicKey,
				SigAlgo:   sigAlgo,
				HashAlgo:  hashAlgo,
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

		tx := templates.CreateAccount(accountKeys, code, signerAccount.Address)

		cli.SendTransaction(projectConf.HostWithOverride(conf.Host), signerAccount, tx, conf.Results)
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

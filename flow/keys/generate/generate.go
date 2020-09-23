package generate

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"

	cli "github.com/dapperlabs/flow-cli/flow"
)

type Config struct {
	Seed    string `flag:"seed,s" info:"deterministic seed phrase"`
	SigAlgo string `default:"ECDSA_P256" flag:"algo,a" info:"signature algorithm"`
}

var conf Config

var Cmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a new key-pair",
	Run: func(cmd *cobra.Command, args []string) {
		var seed []byte
		if conf.Seed == "" {
			seed = cli.RandomSeed(crypto.MinSeedLength)
		} else {
			seed = []byte(conf.Seed)
		}

		sigAlgo := crypto.StringToSignatureAlgorithm(conf.SigAlgo)
		if sigAlgo == crypto.UnknownSignatureAlgorithm {
			cli.Exitf(1, "Invalid signature algorithm: %s", conf.SigAlgo)
		}

		fmt.Printf(
			"Generating key pair with signature algorithm:                 %s\n...\n",
			sigAlgo,
		)

		privateKey, err := crypto.GeneratePrivateKey(sigAlgo, seed)
		if err != nil {
			cli.Exitf(1, "Failed to generate private key: %v", err)
		}

		fmt.Printf(
			"\U0001F510 Private key (\u26A0\uFE0F\u202F store safely and don't share with anyone): %s\n",
			hex.EncodeToString(privateKey.Encode()),
		)
		fmt.Printf(
			"\U0001F54A️\uFE0F\u202F Encoded public key (share freely):                         %x\n",
			privateKey.PublicKey().Encode(),
		)
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

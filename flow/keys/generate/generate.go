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
	Seed string `flag:"seed,s" info:"deterministic seed phrase"`
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

		// Abstract out for now incase we want to allow choosing sig alg,
		// and so we can print out what we're using to generate the key pair
		sigAlg := cli.DefaultSigAlgo

		fmt.Printf(
			"Generating key pair with signature algorithm:                 %s\n...\n",
			sigAlg,
		)

		privateKey, err := crypto.GeneratePrivateKey(sigAlg, seed)
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

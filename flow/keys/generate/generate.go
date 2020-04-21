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
			seed = cli.RandomSeed(crypto.MinSeedLengthECDSA_P256)
		} else {
			seed = []byte(conf.Seed)
		}

		privateKey, err := crypto.GeneratePrivateKey(cli.DefaultSigAlgo, seed)
		if err != nil {
			cli.Exitf(1, "Failed to generate private key: %v", err)
		}

		prKeyBytes := privateKey.Encode()

		fmt.Printf("Generated a new private key:\n")
		fmt.Println(hex.EncodeToString(prKeyBytes))
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

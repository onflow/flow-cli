package hex

import (
	"log"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"

	cli "github.com/onflow/flow-cli/flow"
)

type Config struct {
	Name      string `flag:"name" info:"name of the key"`
	Address   string `flag:"address" info:"flow address of the account"`
	SigAlgo   string `flag:"sigalgo" info:"signature algorithm for the key"`
	HashAlgo  string `flag:"hashalgo" info:"hash algorithm for the key"`
	KeyIndex  int    `flag:"index" info:"index of the key on the account"`
	KeyHex    string `flag:"privatekey" info:"private key in hex format"`
	Overwrite bool   `flag:"overwrite" info:"bool indicating if we should overwrite an existing config with the same name in the config file"`
}

var conf Config

var Cmd = &cobra.Command{
	Use:     "hex",
	Short:   "Save a hex key to the config file",
	Example: "flow keys save hex --name test --address 8c5303eaa26202d6 --sigalgo ECDSA_secp256k1 --hashalgo SHA2_256 --index 0 --privatekey <HEX_PRIVATEKEY>",
	Run: func(cmd *cobra.Command, args []string) {
		projectConf := cli.LoadConfig()

		_, accountExists := projectConf.Accounts[conf.Name]
		if accountExists && !conf.Overwrite {
			cli.Exitf(1, "%s already exists in the config, and overwrite is false", conf.Name)
		}

		// Populate account
		account := &cli.Account{
			KeyType:    cli.KeyTypeHex,
			Address:    flow.HexToAddress(conf.Address),
			SigAlgo:    crypto.StringToSignatureAlgorithm(conf.SigAlgo),
			HashAlgo:   crypto.StringToHashAlgorithm(conf.HashAlgo),
			KeyIndex:   conf.KeyIndex,
			KeyContext: map[string]string{"privateKey": conf.KeyHex},
		}
		privateKey, err := crypto.DecodePrivateKeyHex(account.SigAlgo, conf.KeyHex)
		if err != nil {
			cli.Exitf(1, "key hex could not be parsed")
		}

		account.PrivateKey = privateKey

		// Validate account
		err = account.LoadSigner()
		if err != nil {
			cli.Exitf(1, "provide key could not be loaded as a valid signer %s", conf.KeyHex)
		}

		projectConf.Accounts[conf.Name] = account

		err = cli.SaveConfig(projectConf)
		if err != nil {
			cli.Exitf(1, "could not save config file %s", cli.ConfigPath)
		}

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

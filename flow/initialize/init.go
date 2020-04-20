package initialize

import (
	"fmt"
	"log"

	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/psiemens/sconfig"
	"github.com/spf13/cobra"

	cli "github.com/dapperlabs/flow-cli/flow"
)

type Config struct {
	RootPrivateKey  string `flag:"root-key,k" info:"root account pirvate key"`
	RootKeySigAlgo  string `default:"ECDSA_P256" flag:"root-key-sig-algo" info:"root account key signature algorithm"`
	RootKeyHashAlgo string `default:"SHA3_256" flag:"root-key-hash-algo" info:"root account key hash algorithm"`
	Reset           bool   `default:"false" flag:"reset" info:"reset flow.json config file"`
}

var (
	conf Config
)

var Cmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new account profile",
	Run: func(cmd *cobra.Command, args []string) {
		if !cli.ConfigExists() || conf.Reset {
			var pconf *cli.Config
			if len(conf.RootPrivateKey) > 0 {
				rootKeySigAlgo := crypto.StringToSignatureAlgorithm(conf.RootKeySigAlgo)
				rootKeyHashAlgo := crypto.StringToHashAlgorithm(conf.RootKeyHashAlgo)
				rootKey := cli.MustDecodePrivateKeyHex(rootKeySigAlgo, conf.RootPrivateKey)
				pconf = InitProjectWithRootKey(rootKey, rootKeyHashAlgo)
			} else {
				pconf = InitProject()
			}
			rootAcct := pconf.RootAccount()

			fmt.Printf("⚙️   Flow client initialized with root account:\n\n")
			fmt.Printf("👤  Address: 0x%x\n", rootAcct.Address.Bytes())
			fmt.Printf("ℹ️   Start the emulator with this root account by running: flow emulator start\n")
		} else {
			fmt.Printf("⚠️   Flow configuration file already exists! Begin by running: flow emulator start\n")
		}
	},
}

// InitProject generates a new root key and saves project config.
func InitProject() *cli.Config {
	seed := cli.RandomSeed(crypto.MinSeedLengthECDSA_P256)

	rootKey, err := crypto.GeneratePrivateKey(crypto.ECDSA_P256, seed)
	if err != nil {
		cli.Exitf(1, "Failed to generate private key: %v", err)
	}

	return InitProjectWithRootKey(rootKey, crypto.SHA3_256)
}

// InitProjectWithRootKey creates and saves a new project config
// using the specified root key.
func InitProjectWithRootKey(privateKey crypto.PrivateKey, hashAlgo crypto.HashAlgorithm) *cli.Config {
	pconf := cli.NewConfig()
	pconf.SetRootAccountKey(privateKey, hashAlgo)
	cli.MustSaveConfig(pconf)
	return pconf
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

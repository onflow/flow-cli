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
	ServicePrivateKey  string `flag:"service-priv-key" info:"service account private key"`
	ServiceKeySigAlgo  string `default:"ECDSA_P256" flag:"service-sig-algo" info:"service account key signature algorithm"`
	ServiceKeyHashAlgo string `default:"SHA3_256" flag:"service-hash-algo" info:"service account key hash algorithm"`
	Reset              bool   `default:"false" flag:"reset" info:"reset flow.json config file"`
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
			if len(conf.ServicePrivateKey) > 0 {
				serviceKeySigAlgo := crypto.StringToSignatureAlgorithm(conf.ServiceKeySigAlgo)
				serviceKeyHashAlgo := crypto.StringToHashAlgorithm(conf.ServiceKeyHashAlgo)
				serviceKey := cli.MustDecodePrivateKeyHex(serviceKeySigAlgo, conf.ServicePrivateKey)
				pconf = InitProjectWithServiceKey(serviceKey, serviceKeyHashAlgo)
			} else {
				pconf = InitProject()
			}
			serviceAcct := pconf.ServiceAccount()

			fmt.Printf("⚙️   Flow client initialized with service account:\n\n")
			fmt.Printf("👤  Address: 0x%x\n", serviceAcct.Address.Bytes())
			fmt.Printf("ℹ️   Start the emulator with this service account by running: flow emulator start\n")
		} else {
			fmt.Printf("⚠️   Flow configuration file already exists! Begin by running: flow emulator start\n")
		}
	},
}

// InitProject generates a new service key and saves project config.
func InitProject() *cli.Config {
	seed := cli.RandomSeed(crypto.MinSeedLength)

	serviceKey, err := crypto.GeneratePrivateKey(crypto.ECDSA_P256, seed)
	if err != nil {
		cli.Exitf(1, "Failed to generate private key: %v", err)
	}

	return InitProjectWithServiceKey(serviceKey, crypto.SHA3_256)
}

// InitProjectWithServiceKey creates and saves a new project config
// using the specified service key.
func InitProjectWithServiceKey(privateKey crypto.PrivateKey, hashAlgo crypto.HashAlgorithm) *cli.Config {
	pconf := cli.NewConfig()
	pconf.SetServiceAccountKey(privateKey, hashAlgo)
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

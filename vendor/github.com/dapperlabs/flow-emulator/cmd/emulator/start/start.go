package start

import (
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/prometheus/common/log"
	"github.com/psiemens/sconfig"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	emulator "github.com/dapperlabs/flow-emulator"
	"github.com/dapperlabs/flow-emulator/server"
)

type Config struct {
	Port            int           `default:"3569" flag:"port,p" info:"port to run RPC server"`
	HTTPPort        int           `default:"8080" flag:"http-port" info:"port to run HTTP server"`
	Verbose         bool          `default:"false" flag:"verbose,v" info:"enable verbose logging"`
	BlockTime       time.Duration `flag:"block-time,b" info:"time between sealed blocks"`
	RootPrivateKey  string        `flag:"root-priv-key" info:"root account private key"`
	RootPublicKey   string        `flag:"root-pub-key" info:"root account public key"`
	RootKeySigAlgo  string        `default:"ECDSA_P256" flag:"root-sig-algo" info:"root account key signature algorithm"`
	RootKeyHashAlgo string        `default:"SHA3_256" flag:"root-hash-algo" info:"root account key hash algorithm"`
	Init            bool          `default:"false" flag:"init" info:"whether to initialize a new account profile"`
	GRPCDebug       bool          `default:"false" flag:"grpc-debug" info:"enable gRPC server reflection for debugging with grpc_cli"`
	Persist         bool          `default:"false" flag:"persist" info:"enable persistent storage"`
	DBPath          string        `default:"./flowdb" flag:"dbpath" info:"path to database directory"`
}

const (
	EnvPrefix                 = "FLOW"
	DefaultRootPrivateKeySeed = emulator.DefaultRootPrivateKeySeed
)

var (
	logger *logrus.Logger
	conf   Config
)

var Cmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the Flow emulator server",
	Run: func(cmd *cobra.Command, args []string) {
		var (
			rootPrivateKey  crypto.PrivateKey
			rootPublicKey   crypto.PublicKey
			rootKeySigAlgo  crypto.SignatureAlgorithm
			rootKeyHashAlgo crypto.HashAlgorithm
			err             error
		)

		rootKeySigAlgo = crypto.StringToSignatureAlgorithm(conf.RootKeySigAlgo)
		rootKeyHashAlgo = crypto.StringToHashAlgorithm(conf.RootKeyHashAlgo)

		if len(conf.RootPublicKey) > 0 {
			rootPublicKey, err = crypto.DecodePublicKeyHex(rootKeySigAlgo, conf.RootPublicKey)
			if err != nil {
				Exit(1, err.Error())
			}
		} else if len(conf.RootPrivateKey) > 0 {
			rootPrivateKey, err = crypto.DecodePrivateKeyHex(rootKeySigAlgo, conf.RootPrivateKey)
			if err != nil {
				Exit(1, err.Error())
			}

			rootPublicKey = rootPrivateKey.PublicKey()
		} else {
			rootPrivateKey, _ = crypto.GeneratePrivateKey(crypto.ECDSA_P256, []byte(DefaultRootPrivateKeySeed))

			rootPublicKey = rootPrivateKey.PublicKey()
			rootKeySigAlgo = rootPrivateKey.Algorithm()
			rootKeyHashAlgo = crypto.SHA3_256
		}

		if conf.Verbose {
			logger.SetLevel(logrus.DebugLevel)
		}

		rootAddress := flow.HexToAddress("01")
		rootFields := logrus.Fields{
			"rootAddress":  rootAddress.Hex(),
			"rootPubKey":   hex.EncodeToString(rootPublicKey.Encode()),
			"rootSigAlgo":  rootKeySigAlgo,
			"rootHashAlgo": rootKeyHashAlgo,
		}

		if rootPrivateKey != (crypto.PrivateKey{}) {
			rootFields["rootPrivKey"] = hex.EncodeToString(rootPrivateKey.Encode())
		}

		logger.WithFields(rootFields).Infof("⚙️   Using root account 0x%s", rootAddress.Hex())

		serverConf := &server.Config{
			GRPCPort:  conf.Port,
			GRPCDebug: conf.GRPCDebug,
			HTTPPort:  conf.HTTPPort,
			// TODO: allow headers to be parsed from environment
			HTTPHeaders:     nil,
			BlockTime:       conf.BlockTime,
			RootPublicKey:   rootPublicKey,
			RootKeySigAlgo:  rootKeySigAlgo,
			RootKeyHashAlgo: rootKeyHashAlgo,
			Persist:         conf.Persist,
			DBPath:          conf.DBPath,
		}

		emu := server.NewEmulatorServer(logger, serverConf)
		emu.Start()
	},
}

func init() {
	initLogger()
	initConfig()
}

func initLogger() {
	logger = logrus.New()
	logger.Formatter = new(logrus.TextFormatter)
	logger.Out = os.Stdout
}

func initConfig() {
	err := sconfig.New(&conf).
		FromEnvironment(EnvPrefix).
		BindFlags(Cmd.PersistentFlags()).
		Parse()
	if err != nil {
		log.Fatal(err)
	}
}

func Exit(code int, msg string) {
	fmt.Println(msg)
	os.Exit(code)
}

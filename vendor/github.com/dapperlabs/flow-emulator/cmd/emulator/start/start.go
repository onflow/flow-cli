package start

import (
	"encoding/hex"
	"fmt"
	"os"
	"time"

	sdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/prometheus/common/log"
	"github.com/psiemens/sconfig"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/dapperlabs/flow-emulator/server"
)

type Config struct {
	Port               int           `default:"3569" flag:"port,p" info:"port to run RPC server"`
	HTTPPort           int           `default:"8080" flag:"http-port" info:"port to run HTTP server"`
	Verbose            bool          `default:"false" flag:"verbose,v" info:"enable verbose logging"`
	BlockTime          time.Duration `flag:"block-time,b" info:"time between sealed blocks"`
	ServicePrivateKey  string        `flag:"service-priv-key" info:"service account private key"`
	ServicePublicKey   string        `flag:"service-pub-key" info:"service account public key"`
	ServiceKeySigAlgo  string        `default:"ECDSA_P256" flag:"service-sig-algo" info:"service account key signature algorithm"`
	ServiceKeyHashAlgo string        `default:"SHA3_256" flag:"service-hash-algo" info:"service account key hash algorithm"`
	Init               bool          `default:"false" flag:"init" info:"whether to initialize a new account profile"`
	GRPCDebug          bool          `default:"false" flag:"grpc-debug" info:"enable gRPC server reflection for debugging with grpc_cli"`
	Persist            bool          `default:"false" flag:"persist" info:"enable persistent storage"`
	DBPath             string        `default:"./flowdb" flag:"dbpath" info:"path to database directory"`
}

const EnvPrefix = "FLOW"

var (
	logger *logrus.Logger
	conf   Config
)

type serviceKeyFunc func(init bool) (crypto.PrivateKey, crypto.SignatureAlgorithm, crypto.HashAlgorithm)

func Cmd(getServiceKey serviceKeyFunc) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Starts the Flow emulator server",
		Run: func(cmd *cobra.Command, args []string) {
			var (
				servicePrivateKey  crypto.PrivateKey
				servicePublicKey   crypto.PublicKey
				serviceKeySigAlgo  crypto.SignatureAlgorithm
				serviceKeyHashAlgo crypto.HashAlgorithm
				err                error
			)

			serviceKeySigAlgo = crypto.StringToSignatureAlgorithm(conf.ServiceKeySigAlgo)
			serviceKeyHashAlgo = crypto.StringToHashAlgorithm(conf.ServiceKeyHashAlgo)

			if len(conf.ServicePublicKey) > 0 {
				servicePublicKey, err = crypto.DecodePublicKeyHex(serviceKeySigAlgo, conf.ServicePublicKey)
				if err != nil {
					Exit(1, err.Error())
				}
			} else if len(conf.ServicePrivateKey) > 0 {
				servicePrivateKey, err = crypto.DecodePrivateKeyHex(serviceKeySigAlgo, conf.ServicePrivateKey)
				if err != nil {
					Exit(1, err.Error())
				}

				servicePublicKey = servicePrivateKey.PublicKey()
			} else {
				servicePrivateKey, serviceKeySigAlgo, serviceKeyHashAlgo = getServiceKey(conf.Init)
				servicePublicKey = servicePrivateKey.PublicKey()
			}

			if conf.Verbose {
				logger.SetLevel(logrus.DebugLevel)
			}

			serviceAddress := sdk.ServiceAddress(sdk.Emulator)
			serviceFields := logrus.Fields{
				"serviceAddress":  serviceAddress.Hex(),
				"servicePubKey":   hex.EncodeToString(servicePublicKey.Encode()),
				"serviceSigAlgo":  serviceKeySigAlgo,
				"serviceHashAlgo": serviceKeyHashAlgo,
			}

			if servicePrivateKey != (crypto.PrivateKey{}) {
				serviceFields["servicePrivKey"] = hex.EncodeToString(servicePrivateKey.Encode())
			}

			logger.WithFields(serviceFields).Infof("⚙️   Using service account 0x%s", serviceAddress.Hex())

			serverConf := &server.Config{
				GRPCPort:  conf.Port,
				GRPCDebug: conf.GRPCDebug,
				HTTPPort:  conf.HTTPPort,
				// TODO: allow headers to be parsed from environment
				HTTPHeaders:        nil,
				BlockTime:          conf.BlockTime,
				ServicePublicKey:   servicePublicKey,
				ServiceKeySigAlgo:  serviceKeySigAlgo,
				ServiceKeyHashAlgo: serviceKeyHashAlgo,
				Persist:            conf.Persist,
				DBPath:             conf.DBPath,
			}

			emu := server.NewEmulatorServer(logger, serverConf)
			emu.Start()
		},
	}

	initConfig(cmd)

	return cmd
}

func init() {
	initLogger()
}

func initLogger() {
	logger = logrus.New()
	logger.Formatter = new(logrus.TextFormatter)
	logger.Out = os.Stdout
}

func initConfig(cmd *cobra.Command) {
	err := sconfig.New(&conf).
		FromEnvironment(EnvPrefix).
		BindFlags(cmd.PersistentFlags()).
		Parse()
	if err != nil {
		log.Fatal(err)
	}
}

func Exit(code int, msg string) {
	fmt.Println(msg)
	os.Exit(code)
}

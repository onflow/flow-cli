package emulator

import (
	"fmt"

	emulator "github.com/dapperlabs/flow-emulator"
	"github.com/dapperlabs/flow-emulator/cmd/emulator/start"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/cobra"

	cli "github.com/dapperlabs/flow-cli/flow"
	"github.com/dapperlabs/flow-cli/flow/initialize"
)

var Cmd = &cobra.Command{
	Use:              "emulator",
	Short:            "Flow emulator server",
	TraverseChildren: true,
}




func configuredServiceKey(
	init bool,
	sigAlgo crypto.SignatureAlgorithm,
	hashAlgo crypto.HashAlgorithm,
) (
	crypto.PrivateKey,
	crypto.SignatureAlgorithm,
	crypto.HashAlgorithm,
) {
	if sigAlgo == crypto.UnknownSignatureAlgorithm {
		sigAlgo = emulator.DefaultServiceKeySigAlgo
	}

	if hashAlgo == crypto.UnknownHashAlgorithm {
		hashAlgo = emulator.DefaultServiceKeyHashAlgo
	}

	var serviceAcct *cli.Account

	if init {
		pconf := initialize.InitProject(sigAlgo, hashAlgo)
		serviceAcct = pconf.ServiceAccount()

		fmt.Printf("⚙️   Flow client initialized with service account:\n\n")
		fmt.Printf("👤  Address: 0x%s\n", serviceAcct.Address)
	} else {
		serviceAcct = cli.LoadConfig().ServiceAccount()
	}

	return serviceAcct.PrivateKey, serviceAcct.SigAlgo, serviceAcct.HashAlgo
}

func init() {
	Cmd.AddCommand(start.Cmd(configuredServiceKey))
}

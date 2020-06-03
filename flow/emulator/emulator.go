package emulator

import (
	"fmt"

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

func configuredServiceKey(init bool) (crypto.PrivateKey, crypto.SignatureAlgorithm, crypto.HashAlgorithm) {
	var serviceAcct *cli.Account

	if init {
		pconf := initialize.InitProject()
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

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

func configuredRootKey(init bool) (crypto.PrivateKey, crypto.SignatureAlgorithm, crypto.HashAlgorithm) {
	var rootAcct *cli.Account

	if init {
		pconf := initialize.InitProject()
		rootAcct = pconf.RootAccount()

		fmt.Printf("⚙️   Flow client initialized with root account:\n\n")
		fmt.Printf("👤  Address: 0x%s\n", rootAcct.Address)
	} else {
		rootAcct = cli.LoadConfig().RootAccount()
	}

	return rootAcct.PrivateKey, rootAcct.SigAlgo, rootAcct.HashAlgo
}

func init() {
	Cmd.AddCommand(start.Cmd(configuredRootKey))
}

package evm

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "evm",
	Short:            "Interact with Flow EVM",
	TraverseChildren: true,
}

func init() {
	deployCommand.AddToParent(Cmd)
	createCommand.AddToParent(Cmd)
	getCommand.AddToParent(Cmd)
	runCommand.AddToParent(Cmd)
}

type evmResult struct {
}

func (k *evmResult) JSON() any {
	return nil
}

func (k *evmResult) String() string {
	return ""
}

func (k *evmResult) Oneliner() string {
	return ""
}

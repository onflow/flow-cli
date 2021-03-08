package accounts

import (
	"bytes"
	"fmt"
	"github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"
	"strings"
	"text/tabwriter"
)

var Cmd = &cobra.Command{
	Use:              "accounts",
	Short:            "Utilities to manage accounts",
	TraverseChildren: true,
}

type AccountResult struct {
	*flow.Account
}

func (r *AccountResult) JSON() interface{} {
	return r
}

func (r *AccountResult) String() string {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, '\t', tabwriter.AlignRight)

	fmt.Fprintf(writer, "Address\t %s\n", r.Address)
	fmt.Fprintf(writer, "Balance\t %d\n", r.Balance)

	fmt.Fprintf(writer, "Keys\t %d\n", len(r.Keys))

	for i, key := range r.Keys {
		fmt.Fprintf(writer, "\nKey %d\tPublic Key\t %x\n", i, key.PublicKey.Encode())
		fmt.Fprintf(writer, "\tWeight\t %d\n", key.Weight)
		fmt.Fprintf(writer, "\tSignature Algorithm\t %s\n", key.SigAlgo)
		fmt.Fprintf(writer, "\tHash Algorithm\t %s\n", key.HashAlgo)
		fmt.Fprintf(writer, "\n")
	}

	fmt.Fprintf(writer, "\nCode\t\t %s", strings.ReplaceAll(string(r.Code), "\n", "\n\t "))
	fmt.Fprintf(writer, "\n")

	writer.Flush()

	return b.String()
}

func (r *AccountResult) Oneliner() string {
	return fmt.Sprintf("Address: %s, Balance: %v, Keys: %s", r.Address, r.Balance, r.Keys[0].PublicKey)
}

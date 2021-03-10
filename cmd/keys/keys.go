package keys

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"text/tabwriter"

	"github.com/onflow/flow-go-sdk/crypto"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "keys",
	Short:            "Utilities to manage keys",
	TraverseChildren: true,
}

// KeyResult represent result from all account commands
type KeyResult struct {
	*crypto.PrivateKey
}

// JSON convert result to JSON
func (k *KeyResult) JSON() interface{} {
	result := make(map[string]string, 0)
	result["Private"] = hex.EncodeToString(k.PublicKey().Encode())
	result["Public"] = hex.EncodeToString(k.Encode())

	return result
}

// String convert result to string
func (k *KeyResult) String() string {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, '\t', tabwriter.AlignRight)
	fmt.Fprintf(writer, "🔴️ Store Private Key safely and don't share with anyone! \n")
	fmt.Fprintf(writer, "Public Key \t %x \n", k.PublicKey().Encode())
	fmt.Fprintf(writer, "Private Key \t %x \n", k.Encode())
	writer.Flush()

	return b.String()
}

// Oneliner show result as one liner grep friendly
func (k *KeyResult) Oneliner() string {
	return fmt.Sprintf("")
}

func (k *KeyResult) ToConfig() string {
	return ""
}

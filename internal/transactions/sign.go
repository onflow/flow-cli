package transactions

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/onflow/flow-cli/pkg/flowcli/project"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
	"github.com/spf13/cobra"
)

type flagsSign struct {
	ArgsJSON              string   `default:"" flag:"args-json" info:"arguments in JSON-Cadence format"`
	Args                  []string `default:"" flag:"arg" info:"argument in Type:Value format"`
	Signer                string   `default:"emulator-account" flag:"signer"`
	Payload               string   `flag:"payload" info:"path to the transaction payload file"`
	Proposer              string   `default:"" flag:"proposer"`
	Role                  string   `default:"authorizer" flag:"role"`
	AdditionalAuthorizers []string `flag:"additional-authorizers" info:"Additional authorizer addresses to add to the transaction"`
	PayerAddress          string   `flag:"payer-address" info:"Specify payer of the transaction. Defaults to current signer."`
}

var signFlags = flagsSign{}

var SignCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "sign <optional code filename>",
		Short:   "Sign a transaction",
		Example: `flow transactions sign`,
		Args:    cobra.MaximumNArgs(1),
	},
	Flags: &signFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {
		if sendFlags.Code != "" {
			return nil, fmt.Errorf("⚠️  DEPRECATED: use filename argument")
		}

		if sendFlags.Results {
			return nil, fmt.Errorf("⚠️  DEPRECATED: all transactions will provide results")
		}

		codeFilename := ""
		if len(args) > 0 {
			codeFilename = args[0]
		}

		signed, err := services.Transactions.Sign(
			signFlags.Signer,
			signFlags.Proposer,
			signFlags.PayerAddress,
			signFlags.AdditionalAuthorizers,
			signFlags.Role,
			codeFilename,
			signFlags.Payload,
			signFlags.Args,
			signFlags.ArgsJSON,
		)
		if err != nil {
			return nil, err
		}

		return &SignResult{
			signed: signed,
		}, nil
	},
}

type SignResult struct {
	signed *project.Transaction
}

// JSON convert result to JSON
func (r *SignResult) JSON() interface{} {
	tx := r.signed.FlowTransaction()
	result := make(map[string]string)
	result["Payload"] = fmt.Sprintf("%x", tx.Encode())
	result["Authorizers"] = fmt.Sprintf("%s", tx.Authorizers)

	return result
}

// String convert result to string
func (r *SignResult) String() string {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, '\t', tabwriter.AlignRight)
	tx := r.signed.FlowTransaction()

	fmt.Fprintf(writer, "Authorizers\t%s\n", tx.Authorizers)
	fmt.Fprintf(writer, "Payer\t%s\n", tx.Payer)

	fmt.Fprintf(writer,
		"\nProposal Key:\t\n    Address\t%s\n    Index\t%v\n    Sequence\t%v\n",
		tx.ProposalKey.Address, tx.ProposalKey.KeyIndex, tx.ProposalKey.SequenceNumber,
	)

	for i, e := range tx.PayloadSignatures {
		fmt.Fprintf(writer, "\nPayload Signature %v:\n", i)
		fmt.Fprintf(writer, "    Address\t%s\n", e.Address)
		fmt.Fprintf(writer, "    Signature\t%x\n", e.Signature)
		fmt.Fprintf(writer, "    Key Index\t%v\n", e.KeyIndex)
	}

	for i, e := range tx.EnvelopeSignatures {
		fmt.Fprintf(writer, "\nEnvelope Signature %v:\n", i)
		fmt.Fprintf(writer, "    Address\t%s\n", e.Address)
		fmt.Fprintf(writer, "    Signature\t%s\n", e.Signature)
		fmt.Fprintf(writer, "    Key Index\t%s\n", e.KeyIndex)
	}

	fmt.Fprintf(writer, "\n\nTransaction Payload:\n%x", tx.Encode())

	writer.Flush()
	return b.String()
}

// Oneliner show result as one liner grep friendly
func (r *SignResult) Oneliner() string {
	tx := r.signed.FlowTransaction()
	return fmt.Sprintf("Payload: %x, Authorizers: %s", tx.Encode(), tx.Authorizers)
}

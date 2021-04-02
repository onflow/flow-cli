package output

import (
	"fmt"

	"github.com/gosuri/uilive"
	"github.com/manifoldco/promptui"

	"github.com/onflow/flow-cli/pkg/flowcli/project"
)

func ApproveTransactionPrompt(transaction *project.Transaction) bool {
	writer := uilive.New()
	tx := transaction.FlowTransaction()

	fmt.Fprintf(writer, "\n")
	fmt.Fprintf(writer, "Hash\t%s\n", tx.ID())
	fmt.Fprintf(writer, "Payer\t%x\n", tx.Payer)
	fmt.Fprintf(writer, "Authorizers\t%s\n", tx.Authorizers)

	fmt.Fprintf(writer,
		"\nProposal Key:\t\n    Address\t%s\n    Index\t%v\n    Sequence\t%v\n",
		tx.ProposalKey.Address, tx.ProposalKey.KeyIndex, tx.ProposalKey.SequenceNumber,
	)

	if len(tx.PayloadSignatures) == 0 {
		fmt.Fprintf(writer, "\nNo Payload Signatures\n")
	}

	if len(tx.EnvelopeSignatures) == 0 {
		fmt.Fprintf(writer, "\nNo Envelope Signatures\n")
	}

	for i, e := range tx.PayloadSignatures {
		fmt.Fprintf(writer, "\nPayload Signature %v:\n", i)
		fmt.Fprintf(writer, "    Address\t%s\n", e.Address)
		fmt.Fprintf(writer, "    Signature\t%x\n", e.Signature)
		fmt.Fprintf(writer, "    Key Index\t%d\n", e.KeyIndex)
	}

	for i, e := range tx.EnvelopeSignatures {
		fmt.Fprintf(writer, "\nEnvelope Signature %v:\n", i)
		fmt.Fprintf(writer, "    Address\t%s\n", e.Address)
		fmt.Fprintf(writer, "    Signature\t%x\n", e.Signature)
		fmt.Fprintf(writer, "    Key Index\t%d\n", e.KeyIndex)
	}

	if tx.Script != nil {
		if len(tx.Arguments) == 0 {
			fmt.Fprintf(writer, "\n\nArguments\tNo arguments\n")
		} else {
			fmt.Fprintf(writer, "\n\nArguments (%d):\n", len(tx.Arguments))
			for i, argument := range tx.Arguments {
				fmt.Fprintf(writer, "    - Argument %d: %s\n", i, argument)
			}
		}

		fmt.Fprintf(writer, "\nCode\n\n%s\n", tx.Script)
	}

	fmt.Fprintf(writer, "\n\n")
	writer.Flush()

	prompt := promptui.Select{
		Label: "⚠️  Do you want to sign this transaction?",
		Items: []string{"No", "Yes"},
	}

	_, result, _ := prompt.Run()

	fmt.Fprintf(writer, "\r\r")
	writer.Flush()

	return result == "Yes"
}

package signatures

import (
	"bytes"
	"context"
	"fmt"

	"github.com/onflow/flow-cli/pkg/flowkit/util"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/spf13/cobra"
)

type flagsGenerate struct {
	Signer string `default:"emulator-account" flag:"signer" info:"name of the account used to sign"`
}

var generateFlags = flagsGenerate{}

var GenerateCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "generate <payload>",
		Short:   "Generate the payload signature",
		Example: "flow signatures generate 'The quick brown fox jumps over the lazy dog' --signer alice",
		Args:    cobra.ExactArgs(1),
	},
	Flags: &generateFlags,
	RunS:  sign,
}

func sign(
	args []string,
	_ flowkit.ReaderWriter,
	_ command.GlobalFlags,
	_ *services.Services,
	state *flowkit.State,
) (command.Result, error) {
	payload := []byte(args[0])
	accountName := generateFlags.Signer
	acc, err := state.Accounts().ByName(accountName)
	if err != nil {
		return nil, err
	}

	s, err := acc.Key().Signer(context.Background())
	if err != nil {
		return nil, err
	}

	signed, err := s.Sign(payload)
	if err != nil {
		return nil, err
	}

	return &SignatureResult{
		result:  string(signed),
		payload: string(payload),
		key:     acc.Key(),
	}, nil
}

type SignatureResult struct {
	result  string
	payload string
	key     flowkit.AccountKey
}

func (s *SignatureResult) pubKey() string {
	pkey, err := s.key.PrivateKey()
	if err == nil {
		return (*pkey).PublicKey().String()
	}

	return "ERR"
}

func (s *SignatureResult) JSON() interface{} {
	return map[string]string{
		"signature": fmt.Sprintf("%x", s.result),
		"payload":   fmt.Sprintf("%s", s.payload),
		"hashAlgo":  fmt.Sprintf("%s", s.key.HashAlgo()),
		"sigAlgo":   fmt.Sprintf("%s", s.key.SigAlgo()),
		"pubKey":    fmt.Sprintf("%s", s.pubKey()),
	}
}
func (s *SignatureResult) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)

	_, _ = fmt.Fprintf(writer, "Signature \t %x\n", s.result)
	_, _ = fmt.Fprintf(writer, "Payload \t %s\n", s.payload)
	_, _ = fmt.Fprintf(writer, "Public Key \t %s\n", s.pubKey())
	_, _ = fmt.Fprintf(writer, "Hash Algorithm \t %s\n", s.key.HashAlgo())
	_, _ = fmt.Fprintf(writer, "Signature Algorithm \t %s\n", s.key.SigAlgo())

	_ = writer.Flush()
	return b.String()
}

func (s *SignatureResult) Oneliner() string {

	return fmt.Sprintf(
		"signature: %x, payload: %s, hashAlgo: %s, sigAlgo: %s, pubKey: %s",
		s.result, s.payload, s.key.HashAlgo(), s.key.SigAlgo(), s.pubKey(),
	)
}

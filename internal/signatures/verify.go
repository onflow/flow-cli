package signatures

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/spf13/cobra"
)

type flagsVerify struct {
	Key      string `flag:"key" info:"Public keys for signature verification"`
	SigAlgo  string `flag:"sig-algo" default:"ECDSA_P256" info:"Signature algorithm used to create the public key"`
	HashAlgo string `flag:"hash-algo" default:"SHA3_256" info:"Hashing algorithm used to create signature"`
}

var verifyFlags = flagsVerify{}

var VerifyCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "verify <payload> <signature>",
		Short:   "Verify the signature",
		Example: "flow signatures verify 99fa...25b --key af3...52d",
		Args:    cobra.ExactArgs(2),
	},
	Flags: &verifyFlags,
	RunS:  verify,
}

func verify(
	args []string,
	_ flowkit.ReaderWriter,
	_ command.GlobalFlags,
	_ *services.Services,
	state *flowkit.State,
) (command.Result, error) {
	payload := []byte(args[0])

	sig, err := hex.DecodeString(strings.ReplaceAll(args[1], "0x", ""))
	if err != nil {
		return nil, fmt.Errorf("invalid payload signature: %w", err)
	}

	key, err := hex.DecodeString(strings.ReplaceAll(verifyFlags.Key, "0x", ""))
	if err != nil {
		return nil, fmt.Errorf("invalid public key: %w", err)
	}

	sigAlgo := crypto.StringToSignatureAlgorithm(verifyFlags.SigAlgo)
	hashAlgo := crypto.StringToHashAlgorithm(verifyFlags.HashAlgo)

	pkey, err := crypto.DecodePublicKey(sigAlgo, key)
	if err != nil {
		return nil, fmt.Errorf("invalid public key: %w", err)
	}

	hasher, err := crypto.NewHasher(hashAlgo)
	if err != nil {
		return nil, err
	}

	valid, err := pkey.Verify(sig, payload, hasher)
	if err != nil {
		return nil, err
	}

	return &VerificationResult{
		valid:     valid,
		payload:   payload,
		signature: sig,
	}, nil
}

type VerificationResult struct {
	valid     bool
	payload   []byte
	signature []byte
}

func (s *VerificationResult) JSON() interface{} {
	return map[string]string{
		"valid":     fmt.Sprintf("%v", s.valid),
		"payload":   fmt.Sprintf("%s", s.payload),
		"signature": fmt.Sprintf("%s", s.signature),
	}
}
func (s *VerificationResult) String() string {
	return fmt.Sprintf(
		"valid: \t\t%v\npayload: \t%s\nsignature: \t%x\n",
		s.valid, s.payload, s.signature,
	)
}

func (s *VerificationResult) Oneliner() string {
	return fmt.Sprintf(
		"valid: %v, payload: %s, signature: %x",
		s.valid, s.payload, s.signature,
	)
}

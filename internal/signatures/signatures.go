package signatures

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "signatures",
	Short:            "Signature verification and creation",
	TraverseChildren: true,
}

func init() {
	SignCommand.AddToParent(Cmd)
}

type SignatureResult struct {
	result string
}

func (s *SignatureResult) JSON() interface{} {
	return map[string]string{
		"result": fmt.Sprintf("%x", s.result),
	}
}
func (s *SignatureResult) String() string {
	return fmt.Sprintf("%x", s.result)
}

func (s *SignatureResult) Oneliner() string {
	return fmt.Sprintf("%x", s.result)
}

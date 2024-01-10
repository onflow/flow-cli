package command_test

import (
	"testing"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/spf13/cobra"
)

func TestAddToParent(t *testing.T) {
    cmd := &cobra.Command{}
    c := &command.Command{
        Cmd: cmd,
    }

    c.AddToParent(cmd)

    if c.Cmd.Run == nil {
        t.Errorf("Run function was not initialized")
    }
}
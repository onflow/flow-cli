package command_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/spf13/cobra"
)

func TestInitFlags(t *testing.T) {
    cmd := &cobra.Command{}
    command.InitFlags(cmd)

    flags := []struct {
        name     string
        expected string
    }{
        {"filter", command.Flags.Filter},
        {"format", command.Flags.Format},
        {"save", command.Flags.Save},
        {"host", command.Flags.Host},
        {"network-key", command.Flags.HostNetworkKey},
        {"network", command.Flags.Network},
        {"log", command.Flags.Log},
        {"yes", strconv.FormatBool(command.Flags.Yes)},
        {"config-path", fmt.Sprintf("[%s]",strings.Join(command.Flags.ConfigPaths, ","))},
        {"skip-version-check", strconv.FormatBool(command.Flags.SkipVersionCheck)},
    }

    for _, flag := range flags {
        f := cmd.PersistentFlags().Lookup(flag.name)
        if f == nil {
            t.Errorf("Flag %s was not initialized", flag.name)
        } else if f.DefValue != flag.expected {
            t.Errorf("Flag %s was not initialized with correct default value. Value: %s, Expected: %s", flag.name, f.Value.String(), flag.expected)
        }
    }
}
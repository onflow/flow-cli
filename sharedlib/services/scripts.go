package services

import (
	"github.com/onflow/flow-cli/sharedlib/lib"
	"github.com/onflow/flow-cli/sharedlib/util"

	"github.com/onflow/cadence"

	"github.com/onflow/flow-cli/flow/cli"
	"github.com/onflow/flow-cli/sharedlib/gateway"
)

// Scripts service handles all interactions for scripts
type Scripts struct {
	gateway gateway.Gateway
	project cli.Project
}

// NewScripts create new script service
func NewScripts(gateway gateway.Gateway, project cli.Project) *Scripts {
	return &Scripts{
		gateway: gateway,
		project: project,
	}
}

// Execute script
func (s *Scripts) Execute(scriptFilename string, args []string) (cadence.Value, error) {
	script, err := util.LoadFile(scriptFilename)
	if err != nil {
		return nil, err
	}

	scriptArgs, err := lib.ParseArgumentsCommaSplit(args)
	if err != nil {
		return nil, err
	}

	return s.gateway.ExecuteScript(script, scriptArgs)
}

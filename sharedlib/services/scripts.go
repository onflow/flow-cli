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
	project *cli.Project
	logger  util.Logger
}

// NewScripts create new script service
func NewScripts(
	gateway gateway.Gateway,
	project *cli.Project,
	logger util.Logger,
) *Scripts {
	return &Scripts{
		gateway: gateway,
		project: project,
		logger:  logger,
	}
}

// Execute script
func (s *Scripts) Execute(scriptFilename string, args []string, argsJSON string) (cadence.Value, error) {
	script, err := util.LoadFile(scriptFilename)
	if err != nil {
		return nil, err
	}

	scriptArgs, err := lib.ParseArguments(args, argsJSON)
	if err != nil {
		return nil, err
	}

	return s.gateway.ExecuteScript(script, scriptArgs)
}

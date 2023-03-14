package project

import (
	"github.com/onflow/flow-cli/internal/util"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ProjectDeploy(t *testing.T) {
	srv, state, _ := util.TestMocks(t)

	t.Run("Fail contract errors", func(t *testing.T) {
		srv.DeployProject.Return(nil, &flowkit.ProjectDeploymentError{})
		_, err := deploy([]string{}, util.NoFlags, util.NoLogger, srv.Mock, state)
		assert.EqualError(t, err, "failed deploying all contracts")
	})

}

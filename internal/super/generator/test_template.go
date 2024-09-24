package generator

import (
	"path/filepath"

	"github.com/onflow/flow-cli/internal/util"
	"github.com/onflow/flowkit/v2"
)

type TestTemplate struct {
	Name         string
	TemplatePath string
	Data         map[string]interface{}
}

var _ TemplateItem = TestTemplate{}

// GetName returns the name of the script or transaction
func (o TestTemplate) GetName() string {
	return o.Name
}

// GetTemplate returns an empty string for scripts and transactions
func (o TestTemplate) GetTemplatePath() string {
	if o.TemplatePath == "" {
		return "contract_init_test.cdc.tmpl"
	}

	return o.TemplatePath
}

// GetData returns the data of the script or transaction
func (o TestTemplate) GetData() map[string]interface{} {
	return o.Data
}

func (o TestTemplate) GetTargetPath() string {
	return filepath.Join(DefaultCadenceDirectory, "tests", util.AddCDCExtension(o.Name))
}

func (o TestTemplate) UpdateState(state *flowkit.State) error {
	return nil
}

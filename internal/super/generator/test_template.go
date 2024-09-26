package generator

import (
	"path/filepath"

	"github.com/onflow/flow-cli/internal/util"
)

const (
	DefaultTestDirectory = "tests"
)

type TestTemplate struct {
	Name         string
	TemplatePath string
	Data         map[string]interface{}
}

var _ TemplateItem = TestTemplate{}

func (o TestTemplate) GetType() string {
	return "test"
}

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
	return filepath.Join(DefaultCadenceDirectory, DefaultTestDirectory, util.AddCDCExtension(o.Name))
}

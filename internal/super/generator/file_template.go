package generator

import (
	"github.com/onflow/flowkit/v2"
)

// FileTemplate is a template for raw
type FileTemplate struct {
	TemplatePath string
	TargetPath   string
	Data         map[string]interface{}
}

func NewFileTemplate(
	templatePath string,
	targetPath string,
	data map[string]interface{},
) FileTemplate {
	return FileTemplate{
		TemplatePath: templatePath,
		TargetPath:   targetPath,
		Data:         data,
	}
}

var _ TemplateItem = FileTemplate{}

// GetType returns the type of the contract
func (c FileTemplate) GetType() string {
	return "file"
}

func (c FileTemplate) GetTemplatePath() string {
	return c.TemplatePath
}

// GetData returns the data of the contract
func (c FileTemplate) GetData() map[string]interface{} {
	return c.Data
}

func (c FileTemplate) GetTargetPath() string {
	return c.TargetPath
}

func (c FileTemplate) UpdateState(state *flowkit.State) error {
	return nil
}

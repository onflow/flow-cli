/*
 * Flow CLI
 *
 * Copyright Flow Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package generator

import (
	"bytes"
	"embed"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/onflow/flowkit/v2"

	"github.com/onflow/flowkit/v2/output"
)

const (
	DefaultCadenceDirectory = "cadence"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

// TemplateItem is an interface for different template types
type TemplateItem interface {
	GetTemplatePath() string
	GetData() map[string]interface{}
	GetTargetPath() string
	UpdateState(state *flowkit.State) error
}

//go:generate mockery --name Generator --case underscore
type Generator interface {
	// Create generates files from the provided template items
	Create(items ...TemplateItem) error
}

type GeneratorImpl struct {
	directory   string
	state       *flowkit.State
	logger      output.Logger
	disableLogs bool
	saveState   bool
}

func NewGenerator(
	directory string,
	state *flowkit.State,
	logger output.Logger,
	disableLogs,
	saveState bool,
) *GeneratorImpl {
	return &GeneratorImpl{
		directory:   directory,
		state:       state,
		logger:      logger,
		disableLogs: disableLogs,
		saveState:   saveState,
	}
}

func (g *GeneratorImpl) Create(items ...TemplateItem) error {
	for _, item := range items {
		err := g.generate(item)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *GeneratorImpl) generate(item TemplateItem) error {
	rootDir := g.directory

	targetRelativeToRoot := item.GetTargetPath()
	templatePath := item.GetTemplatePath()
	data := item.GetData()

	fileData := map[string]interface{}{}
	for k, v := range data {
		fileData[k] = v
	}

	outputContent, err := g.processTemplate(templatePath, fileData)
	if err != nil {
		// TODO, better error based on template type
		return fmt.Errorf("error generating template: %w", err)
	}

	targetPath := filepath.Join(rootDir, targetRelativeToRoot)
	targetDirectory := filepath.Dir(targetPath)

	// Check file existence
	if _, err := g.state.ReaderWriter().ReadFile(targetPath); err == nil {
		return fmt.Errorf("file already exists: %s", targetPath)
	}

	// Ensure the directory exists
	if err := g.state.ReaderWriter().MkdirAll(targetDirectory, 0755); err != nil {
		return fmt.Errorf("error creating directories: %w", err)
	}

	// Write files
	err = g.state.ReaderWriter().WriteFile(targetPath, []byte(outputContent), 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	if !g.disableLogs {
		// TODO: Add more detailed logging
		g.logger.Info(fmt.Sprintf("Generated %s", targetPath))
	}

	// Call template state update function if it exists
	err = item.UpdateState(g.state)
	if err != nil {
		return err
	}

	return nil
}

// processTemplate reads a template file from the embedded filesystem and processes it with the provided data
// If you don't need to provide data, pass nil
func (g *GeneratorImpl) processTemplate(templatePath string, data map[string]interface{}) (string, error) {
	templateData, err := templatesFS.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file: %w", err)
	}

	tmpl, err := template.New("template").Parse(string(templateData))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var executedTemplate bytes.Buffer
	// Execute the template with the provided data or nil if no data is needed
	if err = tmpl.Execute(&executedTemplate, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return executedTemplate.String(), nil
}

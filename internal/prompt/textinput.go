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

package prompt

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// textInputModel is now private, only accessible within the 'prompt' package.
type textInputModel struct {
	textInput    textinput.Model
	err          error
	customMsg    string
	validate     func(string) error
	defaultValue string
	cancelled    bool
}

// newTextInput is a private function that initializes a new text input model.
func newTextInput(customMsg, placeholder, defaultValue string, validate func(string) error) textInputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 30
	if defaultValue != "" {
		ti.SetValue(defaultValue)
	}

	return textInputModel{
		textInput:    ti,
		customMsg:    customMsg,
		validate:     validate,
		defaultValue: defaultValue,
	}
}

func (m textInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m textInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			// Validate input before quitting
			if m.validate != nil {
				if err := m.validate(m.textInput.Value()); err != nil {
					m.err = err
					return m, nil
				}
			}
			m.err = nil
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit
		}
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		// Clear error when user types
		if m.err != nil {
			m.err = nil
		}
		return m, cmd
	}

	return m, nil
}

func (m textInputModel) View() string {
	view := fmt.Sprintf("%s\n\n%s\n\n%s", m.customMsg, m.textInput.View(), "(Enter to submit, Esc to quit)")

	if m.err != nil {
		view = fmt.Sprintf("%s\n\n❌ %s", view, m.err.Error())
	}

	return view
}

// RunTextInput remains public. It's the entry point for external usage.
func RunTextInput(customMsg, placeholder string) (string, error) {
	return RunTextInputWithValidation(customMsg, placeholder, "", nil)
}

// RunTextInputWithValidation runs a text input with validation and optional default value
func RunTextInputWithValidation(customMsg, placeholder, defaultValue string, validate func(string) error) (string, error) {
	model := newTextInput(customMsg, placeholder, defaultValue, validate)
	p := tea.NewProgram(model)

	if finalModel, err := p.Run(); err != nil {
		return "", err
	} else {
		final := finalModel.(textInputModel)
		if final.cancelled {
			os.Exit(-1)
		}
		return final.textInput.Value(), nil
	}
}

/*
 * Flow CLI
 *
 * Copyright 2019 Dapper Labs, Inc.
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
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// optionSelectModel represents the prompt state but is now private
type optionSelectModel struct {
	message  string           // message to display
	cursor   int              // position of the cursor
	choices  []string         // items on the list
	selected map[int]struct{} // which items are selected
}

// selectOptions creates a prompt for selecting multiple options but is now private
func selectOptions(options []string, message string) optionSelectModel {
	return optionSelectModel{
		message:  message,
		choices:  options,
		selected: make(map[int]struct{}),
	}
}

func (m optionSelectModel) Init() tea.Cmd {
	return nil // No initial command
}

func (m optionSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc: // Quit the program
			return m, tea.Quit

		case tea.KeyUp: // Navigate up
			if m.cursor > 0 {
				m.cursor--
			}

		case tea.KeyDown: // Navigate down
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case tea.KeySpace: // Select an item
			// Toggle selection
			if _, ok := m.selected[m.cursor]; ok {
				delete(m.selected, m.cursor) // Deselect
			} else {
				m.selected[m.cursor] = struct{}{} // Select
			}

		case tea.KeyEnter: // Confirm selection
			return m, tea.Quit // Quit and process selections in main
		}
	}

	return m, nil
}

func (m optionSelectModel) View() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s\n", m.message))
	b.WriteString("Use arrow keys to navigate, space to select, enter to confirm or skip, q to quit:\n\n")
	for i, choice := range m.choices {
		if m.cursor == i {
			b.WriteString("> ")
		} else {
			b.WriteString("  ")
		}
		// Mark selected items
		if _, ok := m.selected[i]; ok {
			b.WriteString("[x] ")
		} else {
			b.WriteString("[ ] ")
		}
		b.WriteString(choice + "\n")
	}
	return b.String()
}

// RunSelectOptions remains public and is the interface for external usage.
func RunSelectOptions(options []string, message string) ([]string, error) {
	model := selectOptions(options, message)
	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	final := finalModel.(optionSelectModel)
	selectedChoices := make([]string, 0)
	for i := range final.selected {
		selectedChoices = append(selectedChoices, final.choices[i])
	}
	return selectedChoices, nil
}

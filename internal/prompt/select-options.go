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

// OptionSelectModel represents the prompt state
type OptionSelectModel struct {
	message  string           // message to display
	cursor   int              // position of the cursor
	Choices  []string         // items on the list
	Selected map[int]struct{} // which items are selected
}

// SelectOptions creates a prompt for selecting multiple options
func SelectOptions(options []string, message string) OptionSelectModel {
	return OptionSelectModel{
		message:  message,
		Choices:  options,
		Selected: make(map[int]struct{}),
	}
}

func (m OptionSelectModel) Init() tea.Cmd {
	return nil // No initial command
}

func (m OptionSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(m.Choices)-1 {
				m.cursor++
			}

		case tea.KeySpace: // Select an item
			// Toggle selection
			if _, ok := m.Selected[m.cursor]; ok {
				delete(m.Selected, m.cursor) // Deselect
			} else {
				m.Selected[m.cursor] = struct{}{} // Select
			}

		case tea.KeyEnter: // Confirm selection
			return m, tea.Quit // Quit and process selections in main
		}
	}

	return m, nil
}

func (m OptionSelectModel) View() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s.\n", m.message))
	b.WriteString("Use arrow keys to navigate, space to select, enter to confirm, q to quit:\n\n")
	for i, choice := range m.Choices {
		if m.cursor == i {
			b.WriteString("> ")
		} else {
			b.WriteString("  ")
		}
		// Mark selected items
		if _, ok := m.Selected[i]; ok {
			b.WriteString("[x] ")
		} else {
			b.WriteString("[ ] ")
		}
		b.WriteString(choice + "\n")
	}
	return b.String()
}

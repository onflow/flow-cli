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
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	"github.com/onflow/flow-cli/common/branding"
)

// SectionItem represents an item with a section header
type SectionItem struct {
	Name         string
	Section      string
	IsHeader     bool
	IsSelectable bool
}

// sectionSelectModel handles selection across multiple sections with headers
type sectionSelectModel struct {
	message  string           // message to display
	cursor   int              // position of the cursor
	items    []SectionItem    // items with section info
	selected map[int]struct{} // which items are selected
	footer   string           // optional footer message
}

// optionSelectModel represents the prompt state but is now private
type optionSelectModel struct {
	message  string           // message to display
	cursor   int              // position of the cursor
	choices  []string         // items on the list
	selected map[int]struct{} // which items are selected
	footer   string           // optional footer message
}

// selectOptions creates a prompt for selecting multiple options but is now private
func selectOptions(options []string, message string) optionSelectModel {
	return optionSelectModel{
		message:  message,
		choices:  options,
		selected: make(map[int]struct{}),
	}
}

// selectOptionsWithFooter creates a prompt for selecting multiple options with footer message
func selectOptionsWithFooter(options []string, message string, footer string) optionSelectModel {
	return optionSelectModel{
		message:  message,
		choices:  options,
		selected: make(map[int]struct{}),
		footer:   footer,
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
	b.WriteString(fmt.Sprintf("%s\n", branding.MessageStyle.Render(m.message)))
	b.WriteString(branding.GrayStyle.Render("Use arrow keys to navigate, space to select, enter to confirm or skip, q to quit:") + "\n\n")
	for i, choice := range m.choices {
		if m.cursor == i {
			b.WriteString(branding.GreenStyle.Render("> "))
		} else {
			b.WriteString("  ")
		}
		// Mark selected items
		if _, ok := m.selected[i]; ok {
			b.WriteString(branding.GreenStyle.Render("[x] "))
		} else {
			b.WriteString("[ ] ")
		}

		// Style the choice text if it's selected
		if m.cursor == i {
			b.WriteString(branding.GreenStyle.Render(choice) + "\n")
		} else {
			b.WriteString(choice + "\n")
		}
	}

	// Add footer message if present
	if m.footer != "" {
		b.WriteString("\n" + branding.GrayStyle.Render(m.footer))
	}

	return b.String()
}

// RunSelectOptions remains public and is the interface for external usage.
func RunSelectOptions(options []string, message string) ([]string, error) {
	return RunSelectOptionsWithFooter(options, message, "")
}

// RunSelectOptionsWithFooter runs the selection prompt with an optional footer message.
func RunSelectOptionsWithFooter(options []string, message string, footer string) ([]string, error) {
	// Non-interactive fallback for CI: return no selection
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return []string{}, nil
	}
	model := selectOptionsWithFooter(options, message, footer)
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

// singleSelectModel represents a single-choice prompt
type singleSelectModel struct {
	message   string   // message to display
	cursor    int      // position of the cursor
	choices   []string // items on the list
	selected  int      // which item is selected (-1 for none)
	cancelled bool     // whether the user cancelled
}

// newSingleSelect creates a single-select prompt
func newSingleSelect(options []string, message string) singleSelectModel {
	return singleSelectModel{
		message:  message,
		choices:  options,
		selected: -1,
	}
}

func (m singleSelectModel) Init() tea.Cmd {
	return nil
}

func (m singleSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit

		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}

		case tea.KeyDown:
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case tea.KeyEnter:
			m.selected = m.cursor
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m singleSelectModel) View() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s\n\n", branding.MessageStyle.Render(m.message)))

	for i, choice := range m.choices {
		if m.cursor == i {
			b.WriteString(branding.GreenStyle.Render("> "))
			b.WriteString(branding.GreenStyle.Render(choice) + "\n")
		} else {
			b.WriteString("  ")
			b.WriteString(choice + "\n")
		}
	}

	b.WriteString("\n" + branding.GrayStyle.Render("Use arrow keys to navigate, enter to select, esc to cancel"))
	return b.String()
}

// RunSingleSelect runs a single-choice selection prompt
func RunSingleSelect(options []string, message string) (string, error) {
	// Non-interactive fallback for CI: default to first option (safe default)
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		if len(options) == 0 {
			return "", fmt.Errorf("no options provided")
		}
		return options[0], nil
	}
	model := newSingleSelect(options, message)
	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	final := finalModel.(singleSelectModel)
	if final.cancelled {
		return "", fmt.Errorf("selection cancelled")
	}

	if final.selected >= 0 && final.selected < len(final.choices) {
		return final.choices[final.selected], nil
	}

	return "", fmt.Errorf("no selection made")
}

// RunSelectSections runs a selection prompt with grouped sections and headers
func RunSelectSections(sections map[string][]string, message string, footer string) ([]string, error) {
	// Non-interactive fallback for CI: return no selection
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return []string{}, nil
	}

	// Core Contracts first, then others alphabetically
	orderedSectionNames := make([]string, 0, len(sections))
	if _, exists := sections["Core Contracts"]; exists {
		orderedSectionNames = append(orderedSectionNames, "Core Contracts")
	}

	for sectionName := range sections {
		if sectionName != "Core Contracts" {
			orderedSectionNames = append(orderedSectionNames, sectionName)
		}
	}

	var items []SectionItem
	for _, sectionName := range orderedSectionNames {
		options := sections[sectionName]
		if len(options) == 0 {
			continue
		}

		items = append(items, SectionItem{
			Name:         sectionName,
			Section:      sectionName,
			IsHeader:     true,
			IsSelectable: false,
		})

		for _, option := range options {
			items = append(items, SectionItem{
				Name:         option,
				Section:      sectionName,
				IsHeader:     false,
				IsSelectable: true,
			})
		}
	}

	model := sectionSelectModel{
		message:  message,
		items:    items,
		selected: make(map[int]struct{}),
		footer:   footer,
	}

	for i, item := range items {
		if item.IsSelectable {
			model.cursor = i
			break
		}
	}

	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	final := finalModel.(sectionSelectModel)
	selectedChoices := make([]string, 0)
	for i := range final.selected {
		if i < len(final.items) && final.items[i].IsSelectable {
			selectedChoices = append(selectedChoices, final.items[i].Name)
		}
	}
	return selectedChoices, nil
}

func (m sectionSelectModel) Init() tea.Cmd {
	return nil
}

func (m sectionSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyUp:
			for i := m.cursor - 1; i >= 0; i-- {
				if m.items[i].IsSelectable {
					m.cursor = i
					break
				}
			}

		case tea.KeyDown:
			for i := m.cursor + 1; i < len(m.items); i++ {
				if m.items[i].IsSelectable {
					m.cursor = i
					break
				}
			}

		case tea.KeySpace:
			if m.cursor < len(m.items) && m.items[m.cursor].IsSelectable {
				if _, ok := m.selected[m.cursor]; ok {
					delete(m.selected, m.cursor)
				} else {
					m.selected[m.cursor] = struct{}{}
				}
			}

		case tea.KeyEnter:
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m sectionSelectModel) View() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s\n", branding.MessageStyle.Render(m.message)))
	b.WriteString(branding.GrayStyle.Render("Use arrow keys to navigate, space to select, enter to confirm or skip, q to quit:") + "\n\n")

	for i, item := range m.items {
		if item.IsHeader {
			b.WriteString(branding.PurpleStyle.Render(fmt.Sprintf("--- %s ---", item.Name)) + "\n")
		} else {
			if m.cursor == i {
				b.WriteString(branding.GreenStyle.Render("> "))
			} else {
				b.WriteString("  ")
			}

			if _, ok := m.selected[i]; ok {
				b.WriteString(branding.GreenStyle.Render("[x] "))
			} else {
				b.WriteString("[ ] ")
			}

			if m.cursor == i {
				b.WriteString(branding.GreenStyle.Render(item.Name) + "\n")
			} else {
				b.WriteString(item.Name + "\n")
			}
		}
	}

	if m.footer != "" {
		b.WriteString("\n" + branding.GrayStyle.Render(m.footer))
	}

	return b.String()
}

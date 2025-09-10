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
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

	"github.com/onflow/flow-cli/common/branding"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

// ContractItem represents a contract that can be selected
type ContractItem struct {
	name        string
	section     string
	description string
	selected    bool
	isHeader    bool
}

func (i ContractItem) Title() string {
	if i.isHeader {
		return branding.PurpleStyle.Render("--- " + i.name + " ---")
	}
	
	prefix := "[ ] "
	if i.selected {
		prefix = branding.GreenStyle.Render("[✓] ")
	}
	
	return prefix + i.name
}

func (i ContractItem) Description() string {
	if i.isHeader {
		return branding.GrayStyle.Render(i.description)
	}
	return branding.GrayStyle.Render(i.section)
}

func (i ContractItem) FilterValue() string {
	return i.name
}

// ContractListModel handles the list-based contract selection
type ContractListModel struct {
	list     list.Model
	items    []ContractItem
	selected map[int]struct{}
	footer   string
	quitting bool
}

func (m ContractListModel) Init() tea.Cmd {
	return nil
}

func (m ContractListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case " ":
			// Toggle selection
			selectedIndex := m.list.Index()
			if selectedIndex < len(m.items) && !m.items[selectedIndex].isHeader {
				if _, exists := m.selected[selectedIndex]; exists {
					delete(m.selected, selectedIndex)
					m.items[selectedIndex].selected = false
				} else {
					m.selected[selectedIndex] = struct{}{}
					m.items[selectedIndex].selected = true
				}
				
				// Update the list items
				listItems := make([]list.Item, len(m.items))
				for i, item := range m.items {
					listItems[i] = item
				}
				m.list.SetItems(listItems)
			}
		case "enter":
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ContractListModel) View() string {
	if m.quitting {
		return ""
	}
	
	view := m.list.View()
	
	// Add footer if present
	if m.footer != "" {
		view += "\n" + branding.GrayStyle.Render(m.footer)
	}
	
	// Add instructions
	instructions := branding.GrayStyle.Render("Space: select/deselect • Enter: confirm • q/Ctrl+C: quit")
	view += "\n" + instructions
	
	return docStyle.Render(view)
}

// RunContractList runs a list-based contract selection prompt
func RunContractList(sections map[string][]string, message string, footer string) ([]string, error) {
	// Non-interactive fallback for CI: return no selection
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return []string{}, nil
	}

	// Build contract items with sections
	var items []ContractItem
	var listItems []list.Item
	
	// Ensure consistent section order: Core Contracts first, then others
	orderedSectionNames := make([]string, 0, len(sections))
	if _, exists := sections["Core Contracts"]; exists {
		orderedSectionNames = append(orderedSectionNames, "Core Contracts")
	}
	
	for sectionName := range sections {
		if sectionName != "Core Contracts" {
			orderedSectionNames = append(orderedSectionNames, sectionName)
		}
	}

	for _, sectionName := range orderedSectionNames {
		contracts := sections[sectionName]
		if len(contracts) == 0 {
			continue
		}

		// Add section header
		headerItem := ContractItem{
			name:        sectionName,
			section:     "",
			description: getSectionDescription(sectionName),
			selected:    false,
			isHeader:    true,
		}
		items = append(items, headerItem)
		listItems = append(listItems, headerItem)

		// Add contracts in section
		for _, contract := range contracts {
			contractItem := ContractItem{
				name:        contract,
				section:     sectionName,
				description: "",
				selected:    false,
				isHeader:    false,
			}
			items = append(items, contractItem)
			listItems = append(listItems, contractItem)
		}
	}

	// Create list model
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = branding.GreenStyle
	delegate.Styles.SelectedDesc = branding.GrayStyle
	
	listModel := list.New(listItems, delegate, 0, 0)
	listModel.Title = message
	listModel.Styles.Title = branding.MessageStyle
	listModel.SetShowStatusBar(false)
	listModel.SetFilteringEnabled(true)
	listModel.SetShowHelp(false)

	model := ContractListModel{
		list:     listModel,
		items:    items,
		selected: make(map[int]struct{}),
		footer:   footer,
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	final := finalModel.(ContractListModel)
	
	// Collect selected contracts
	var selectedContracts []string
	for i := range final.selected {
		if i < len(final.items) && !final.items[i].isHeader {
			selectedContracts = append(selectedContracts, final.items[i].name)
		}
	}

	return selectedContracts, nil
}

func getSectionDescription(sectionName string) string {
	switch sectionName {
	case "Core Contracts":
		return "Essential Flow blockchain system contracts"
	case "DeFi Actions":
		return "DeFi protocol integration contracts for automated actions"
	default:
		return ""
	}
}
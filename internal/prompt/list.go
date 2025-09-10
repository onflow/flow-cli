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

type Item struct {
	name            string
	section         string
	sectionDesc     string
	selected        bool
	isFirstInSection bool
}

func (i Item) Title() string {
	title := ""
	
	if i.isFirstInSection {
		title += branding.PurpleStyle.Render("--- " + i.section + " ---") + "\n"
		if i.sectionDesc != "" {
			title += branding.GrayStyle.Render(i.sectionDesc) + "\n"
		}
	}
	
	prefix := "[ ] "
	if i.selected {
		prefix = branding.GreenStyle.Render("[✓] ")
	}
	
	title += prefix + i.name
	return title
}

func (i Item) Description() string {
	return branding.GrayStyle.Render(i.section)
}

func (i Item) FilterValue() string {
	return i.name
}

type ListModel struct {
	list     list.Model
	items    []Item
	selected map[int]struct{}
	footer   string
	quitting bool
}

func (m ListModel) Init() tea.Cmd {
	return nil
}

func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case " ":
			selectedIndex := m.list.Index()
			if selectedIndex < len(m.items) {
				if _, exists := m.selected[selectedIndex]; exists {
					delete(m.selected, selectedIndex)
					m.items[selectedIndex].selected = false
				} else {
					m.selected[selectedIndex] = struct{}{}
					m.items[selectedIndex].selected = true
				}
				
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

func (m ListModel) View() string {
	if m.quitting {
		return ""
	}
	
	view := m.list.View()
	
	if m.footer != "" {
		view += "\n" + branding.GrayStyle.Render(m.footer)
	}
	
	instructions := branding.GrayStyle.Render("Space: select/deselect • Enter: confirm • q/Ctrl+C: quit")
	view += "\n" + instructions
	
	return docStyle.Render(view)
}

type ListSectionData struct {
	Name        string
	Description string
	Items       []string
}

func RunList(sections []ListSectionData, message string, footer string) ([]string, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return []string{}, nil
	}

	var items []Item
	var listItems []list.Item

	for _, section := range sections {
		if len(section.Items) == 0 {
			continue
		}

		for i, item := range section.Items {
			listItem := Item{
				name:            item,
				section:         section.Name,
				sectionDesc:     section.Description,
				selected:        false,
				isFirstInSection: i == 0,
			}
			items = append(items, listItem)
			listItems = append(listItems, listItem)
		}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = branding.GreenStyle
	delegate.Styles.SelectedDesc = branding.GrayStyle
	
	listModel := list.New(listItems, delegate, 0, 0)
	listModel.Title = message
	listModel.Styles.Title = branding.MessageStyle
	listModel.SetShowStatusBar(false)
	listModel.SetFilteringEnabled(true)
	listModel.SetShowHelp(false)

	model := ListModel{
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

	final := finalModel.(ListModel)
	
	var selectedItems []string
	for i := range final.selected {
		if i < len(final.items) {
			selectedItems = append(selectedItems, final.items[i].name)
		}
	}

	return selectedItems, nil
}


package prompt

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// textInputModel is now private, only accessible within the 'prompt' package.
type textInputModel struct {
	textInput textinput.Model
	err       error
	customMsg string
}

// newTextInput is a private function that initializes a new text input model.
func newTextInput(customMsg, placeholder string) textInputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 30

	return textInputModel{
		textInput: ti,
		customMsg: customMsg,
	}
}

func (m textInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m textInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m textInputModel) View() string {
	return fmt.Sprintf("%s\n\n%s\n\n%s", m.customMsg, m.textInput.View(), "(Enter to submit, Esc to quit)")
}

// RunTextInput remains public. It's the entry point for external usage.
func RunTextInput(customMsg, placeholder string) (string, error) {
	model := newTextInput(customMsg, placeholder)
	p := tea.NewProgram(model)

	if finalModel, err := p.Run(); err != nil {
		return "", err
	} else {
		final := finalModel.(textInputModel)
		return final.textInput.Value(), nil
	}
}

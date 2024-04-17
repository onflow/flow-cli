package prompt

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type TextInputModel struct {
	textInput textinput.Model
	err       error
	customMsg string
}

// NewTextInput initializes a new text input model with a custom message
func NewTextInput(customMsg, placeholder string) TextInputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 30

	return TextInputModel{
		textInput: ti,
		customMsg: customMsg,
	}
}

func (m TextInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m TextInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m TextInputModel) View() string {
	return fmt.Sprintf("%s\n\n%s\n\n%s", m.customMsg, m.textInput.View(), "(Enter to submit, Esc to quit)")
}

// RunTextInput handles running the text input and retrieving the result
func RunTextInput(customMsg, placeholder string) (string, error) {
	model := NewTextInput(customMsg, placeholder)
	p := tea.NewProgram(model)

	if finalModel, err := p.Run(); err != nil {
		return "", err // return the error to handle it outside if necessary
	} else {
		final := finalModel.(TextInputModel)
		return final.textInput.Value(), nil // directly return the input value
	}
}

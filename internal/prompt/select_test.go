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
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestSelectOptions(t *testing.T) {
	t.Run("basic initialization", func(t *testing.T) {
		options := []string{"Option 1", "Option 2", "Option 3"}
		message := "Select options"

		model := selectOptions(options, message)

		assert.Equal(t, message, model.message)
		assert.Equal(t, options, model.choices)
		assert.Equal(t, 0, model.cursor)
		assert.Equal(t, 0, len(model.selected))
	})
}

func TestOptionSelectModel_Init(t *testing.T) {
	model := selectOptions([]string{"A", "B"}, "Test")
	cmd := model.Init()

	assert.Nil(t, cmd)
}

func TestOptionSelectModel_Update(t *testing.T) {
	t.Run("cursor navigation down", func(t *testing.T) {
		model := selectOptions([]string{"A", "B", "C"}, "Test")
		assert.Equal(t, 0, model.cursor)

		updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyDown})
		finalModel := updatedModel.(optionSelectModel)

		assert.Equal(t, 1, finalModel.cursor)
		assert.Nil(t, cmd)
	})

	t.Run("cursor navigation up", func(t *testing.T) {
		model := selectOptions([]string{"A", "B", "C"}, "Test")
		model.cursor = 2

		updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyUp})
		finalModel := updatedModel.(optionSelectModel)

		assert.Equal(t, 1, finalModel.cursor)
		assert.Nil(t, cmd)
	})

	t.Run("cursor stays at bounds", func(t *testing.T) {
		model := selectOptions([]string{"A", "B"}, "Test")

		// Test upper bound
		model.cursor = 1
		updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
		finalModel := updatedModel.(optionSelectModel)
		assert.Equal(t, 1, finalModel.cursor) // Should stay at 1

		// Test lower bound
		model.cursor = 0
		updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
		finalModel = updatedModel.(optionSelectModel)
		assert.Equal(t, 0, finalModel.cursor) // Should stay at 0
	})

	t.Run("space key toggles selection", func(t *testing.T) {
		model := selectOptions([]string{"A", "B", "C"}, "Test")
		model.cursor = 1

		// First space - select
		updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeySpace})
		finalModel := updatedModel.(optionSelectModel)

		_, isSelected := finalModel.selected[1]
		assert.True(t, isSelected)
		assert.Nil(t, cmd)

		// Second space - deselect
		updatedModel, cmd = finalModel.Update(tea.KeyMsg{Type: tea.KeySpace})
		finalModel = updatedModel.(optionSelectModel)

		_, isSelected = finalModel.selected[1]
		assert.False(t, isSelected)
		assert.Nil(t, cmd)
	})

	t.Run("enter key quits", func(t *testing.T) {
		model := selectOptions([]string{"A", "B"}, "Test")

		updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

		assert.NotNil(t, cmd)
		_ = updatedModel.(optionSelectModel) // Should not panic
	})

	t.Run("escape key quits", func(t *testing.T) {
		model := selectOptions([]string{"A", "B"}, "Test")

		updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})

		assert.NotNil(t, cmd)
		_ = updatedModel.(optionSelectModel) // Should not panic
	})

	t.Run("ctrl+c quits", func(t *testing.T) {
		model := selectOptions([]string{"A", "B"}, "Test")

		updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

		assert.NotNil(t, cmd)
		_ = updatedModel.(optionSelectModel) // Should not panic
	})
}

func TestOptionSelectModel_View(t *testing.T) {
	t.Run("displays message and choices", func(t *testing.T) {
		model := selectOptions([]string{"Option A", "Option B"}, "Choose options")

		view := model.View()

		assert.Contains(t, view, "Choose options")
		assert.Contains(t, view, "Option A")
		assert.Contains(t, view, "Option B")
		assert.Contains(t, view, "Use arrow keys to navigate")
	})

	t.Run("shows cursor position", func(t *testing.T) {
		model := selectOptions([]string{"A", "B"}, "Test")
		model.cursor = 1

		view := model.View()

		// Should show cursor on second item
		assert.Contains(t, view, "> [ ] B")
		assert.Contains(t, view, "  [ ] A")
	})

	t.Run("shows selected items", func(t *testing.T) {
		model := selectOptions([]string{"A", "B", "C"}, "Test")
		model.selected[0] = struct{}{}
		model.selected[2] = struct{}{}

		view := model.View()

		assert.Contains(t, view, "[x] A")
		assert.Contains(t, view, "[ ] B")
		assert.Contains(t, view, "[x] C")
	})
}

func TestNewSingleSelect(t *testing.T) {
	t.Run("basic initialization", func(t *testing.T) {
		options := []string{"Yes", "No"}
		message := "Confirm action?"

		model := newSingleSelect(options, message)

		assert.Equal(t, message, model.message)
		assert.Equal(t, options, model.choices)
		assert.Equal(t, 0, model.cursor)
		assert.Equal(t, -1, model.selected)
		assert.False(t, model.cancelled)
	})
}

func TestSingleSelectModel_Init(t *testing.T) {
	model := newSingleSelect([]string{"Yes", "No"}, "Test")
	cmd := model.Init()

	assert.Nil(t, cmd)
}

func TestSingleSelectModel_Update(t *testing.T) {
	t.Run("cursor navigation", func(t *testing.T) {
		model := newSingleSelect([]string{"A", "B", "C"}, "Test")

		// Navigate down
		updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyDown})
		finalModel := updatedModel.(singleSelectModel)
		assert.Equal(t, 1, finalModel.cursor)
		assert.Nil(t, cmd)

		// Navigate up
		updatedModel, cmd = finalModel.Update(tea.KeyMsg{Type: tea.KeyUp})
		finalModel = updatedModel.(singleSelectModel)
		assert.Equal(t, 0, finalModel.cursor)
		assert.Nil(t, cmd)
	})

	t.Run("enter key selects and quits", func(t *testing.T) {
		model := newSingleSelect([]string{"Yes", "No"}, "Test")
		model.cursor = 1 // Position on "No"

		updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		finalModel := updatedModel.(singleSelectModel)

		assert.Equal(t, 1, finalModel.selected)
		assert.False(t, finalModel.cancelled)
		assert.NotNil(t, cmd)
	})

	t.Run("escape key cancels", func(t *testing.T) {
		model := newSingleSelect([]string{"Yes", "No"}, "Test")

		updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
		finalModel := updatedModel.(singleSelectModel)

		assert.True(t, finalModel.cancelled)
		assert.NotNil(t, cmd)
	})

	t.Run("ctrl+c cancels", func(t *testing.T) {
		model := newSingleSelect([]string{"Yes", "No"}, "Test")

		updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
		finalModel := updatedModel.(singleSelectModel)

		assert.True(t, finalModel.cancelled)
		assert.NotNil(t, cmd)
	})
}

func TestSingleSelectModel_View(t *testing.T) {
	t.Run("displays message and choices", func(t *testing.T) {
		model := newSingleSelect([]string{"Yes", "No"}, "Confirm?")

		view := model.View()

		assert.Contains(t, view, "Confirm?")
		assert.Contains(t, view, "Yes")
		assert.Contains(t, view, "No")
		assert.Contains(t, view, "Use arrow keys to navigate")
	})

	t.Run("shows cursor position", func(t *testing.T) {
		model := newSingleSelect([]string{"A", "B"}, "Test")
		model.cursor = 1

		view := model.View()

		// Should show cursor on second item
		assert.Contains(t, view, "> B")
		assert.Contains(t, view, "  A")
	})
}

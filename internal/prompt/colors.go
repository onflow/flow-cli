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

import "github.com/charmbracelet/lipgloss"

// Shared color constants for prompts
var (
	FlowGreen   = lipgloss.Color("#00ff88")
	GrayText    = lipgloss.Color("8")
	PurpleText  = lipgloss.Color("#8b5cf6")
)

// Shared styles for prompts
var (
	GreenStyle    = lipgloss.NewStyle().Foreground(FlowGreen)
	GrayStyle     = lipgloss.NewStyle().Foreground(GrayText)
	PurpleStyle   = lipgloss.NewStyle().Foreground(PurpleText)
	MessageStyle  = PurpleStyle // Use purple for messages/questions
)
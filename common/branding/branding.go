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

package branding

import "github.com/charmbracelet/lipgloss"

// Flow brand colors
var (
	FlowGreen  = lipgloss.Color("#02D87E")
	GrayText   = lipgloss.Color("8")
	PurpleText = lipgloss.Color("#823EE4")
	ErrorRed   = lipgloss.Color("#E55555")
)

// Shared styles for consistent branding
var (
	GreenStyle   = lipgloss.NewStyle().Foreground(FlowGreen)
	GrayStyle    = lipgloss.NewStyle().Foreground(GrayText)
	PurpleStyle  = lipgloss.NewStyle().Foreground(PurpleText)
	MessageStyle = PurpleStyle
	ErrorStyle   = lipgloss.NewStyle().Foreground(ErrorRed).Bold(true)
)

// Flow ASCII art logo
const FlowASCII = `   ___  ___
 /'___\/\_ \
/\ \__/\//\ \     ___   __  __  __
\ \ ,__\ \ \ \   / __` + "`" + `\/\ \/\ \/\ \
 \ \ \_/  \_\ \_/\ \L\ \ \ \_/ \_/ \
  \ \_\   /\____\ \____/\ \___x___/'
   \/_/   \/____/\/___/  \/__//__/
`

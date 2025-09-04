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

import "github.com/onflow/flow-cli/internal/branding"

// Re-export branding colors and styles for backward compatibility
var (
	FlowGreen  = branding.FlowGreen
	GrayText   = branding.GrayText
	PurpleText = branding.PurpleText
)

// Re-export branding styles for backward compatibility
var (
	GreenStyle   = branding.GreenStyle
	GrayStyle    = branding.GrayStyle
	PurpleStyle  = branding.PurpleStyle
	MessageStyle = branding.MessageStyle
)

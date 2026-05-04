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

package settings

import (
	"fmt"
	"os"
	"runtime"
)

const (
	metricsEnabled = "MetricsEnabled"
	flowserPath    = "FlowserPath"
)

// defaults holds the default values for global settings
var defaults = map[string]any{
	metricsEnabled: true,
	flowserPath:    getDefaultInstallDir(),
}

const (
	Darwin  = "darwin"
	Windows = "windows"
	Linux   = "linux"
)

// getDefaultInstallDir returns default installation directory based on the OS.
func getDefaultInstallDir() string {
	switch runtime.GOOS {
	case Darwin:
		return "/Applications"
	case Windows:
		// https://superuser.com/questions/1327037/what-choices-do-i-have-about-where-to-install-software-on-windows-10
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		return fmt.Sprintf(`%s\AppData\Local\Programs`, homeDir)
	case Linux:
		// https://unix.stackexchange.com/questions/127076/into-which-directory-should-i-install-programs-in-linux
		// Use path in users home folder to not require sudo permissions for installation
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		return fmt.Sprintf(`%s/.local/bin`, homeDir)
	default:
		return ""
	}
}

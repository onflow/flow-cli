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

package schedule

import (
	"fmt"

	"github.com/onflow/flow-cli/common/branding"
)

// GetColoredStatus returns a colored status string based on the status code
func GetColoredStatus(status uint8) string {
	statusStr := GetStatusString(status)
	switch status {
	case 0, 1: // Pending, Scheduled
		return branding.GrayStyle.Render(statusStr)
	case 2: // Executing
		return branding.PurpleStyle.Render(statusStr)
	case 3: // Executed
		return branding.GreenStyle.Render(statusStr)
	case 4: // Failed
		return branding.ErrorStyle.Render(statusStr)
	case 5: // Cancelled
		return branding.GrayStyle.Render(statusStr)
	default:
		return statusStr
	}
}

// GetColoredPriority returns a colored priority string based on the priority code
func GetColoredPriority(priority uint8) string {
	priorityStr := GetPriorityString(priority)
	switch priority {
	case 0: // Low
		return branding.GrayStyle.Render(priorityStr)
	case 1: // Medium
		return branding.PurpleStyle.Render(priorityStr)
	case 2: // High
		return branding.GreenStyle.Render(priorityStr)
	default:
		return priorityStr
	}
}

// FormatTransactionDetails returns a formatted string with transaction details
func FormatTransactionDetails(tx *TransactionData) string {
	var output string

	// Status
	statusLabel := branding.GrayStyle.Render("   Status:")
	statusValue := GetColoredStatus(tx.Status)
	output += fmt.Sprintf("%s %s\n", statusLabel, statusValue)

	// Priority
	priorityLabel := branding.GrayStyle.Render("   Priority:")
	priorityValue := GetColoredPriority(tx.Priority)
	output += fmt.Sprintf("%s %s\n", priorityLabel, priorityValue)

	// Execution Effort
	effortLabel := branding.GrayStyle.Render("   Execution Effort:")
	effortValue := branding.PurpleStyle.Render(fmt.Sprintf("%d", tx.ExecutionEffort))
	output += fmt.Sprintf("%s %s\n", effortLabel, effortValue)

	// Fees
	feesLabel := branding.GrayStyle.Render("   Fees:")
	feesValue := branding.PurpleStyle.Render(fmt.Sprintf("%s FLOW", tx.Fees))
	output += fmt.Sprintf("%s %s\n", feesLabel, feesValue)

	// Scheduled Timestamp
	timestampLabel := branding.GrayStyle.Render("   Scheduled Timestamp:")
	timestampValue := branding.PurpleStyle.Render(tx.ScheduledTimestamp)
	output += fmt.Sprintf("%s %s\n", timestampLabel, timestampValue)

	// Handler Type
	handlerTypeLabel := branding.GrayStyle.Render("   Handler Type:")
	handlerTypeValue := branding.PurpleStyle.Render(tx.HandlerTypeIdentifier)
	output += fmt.Sprintf("%s %s\n", handlerTypeLabel, handlerTypeValue)

	// Handler Address
	handlerAddrLabel := branding.GrayStyle.Render("   Handler Address:")
	handlerAddrValue := branding.PurpleStyle.Render(tx.HandlerAddress)
	output += fmt.Sprintf("%s %s\n", handlerAddrLabel, handlerAddrValue)

	return output
}

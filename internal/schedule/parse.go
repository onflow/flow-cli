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

	"github.com/onflow/cadence"
)

// TransactionData holds the parsed transaction data from the scheduler
type TransactionData struct {
	ID                    uint64
	Priority              uint8
	ExecutionEffort       uint64
	Status                uint8
	Fees                  string
	ScheduledTimestamp    string
	HandlerTypeIdentifier string
	HandlerAddress        string
}

// ParseTransactionData parses the cadence.Value returned from a script into TransactionData
func ParseTransactionData(value cadence.Value) (*TransactionData, error) {
	// Check if result is nil (optional return)
	optional, ok := value.(cadence.Optional)
	if !ok {
		return nil, fmt.Errorf("expected optional value, got %T", value)
	}

	if optional.Value == nil {
		return nil, nil // Transaction not found
	}

	// Cast to struct
	structValue, ok := optional.Value.(cadence.Struct)
	if !ok {
		return nil, fmt.Errorf("expected struct value, got %T", optional.Value)
	}

	// Get fields mapped by name
	fields := cadence.FieldsMappedByName(structValue)

	// Parse individual fields
	result := &TransactionData{}

	// ID (UInt64)
	if id, ok := fields["id"].(cadence.UInt64); ok {
		result.ID = uint64(id)
	}

	// Priority (Enum with rawValue)
	if priority, ok := fields["priority"].(cadence.Enum); ok {
		priorityFields := cadence.FieldsMappedByName(priority)
		if rawValue, ok := priorityFields["rawValue"].(cadence.UInt8); ok {
			result.Priority = uint8(rawValue)
		}
	}

	// Execution Effort (UInt64)
	if effort, ok := fields["executionEffort"].(cadence.UInt64); ok {
		result.ExecutionEffort = uint64(effort)
	}

	// Status (Enum with rawValue)
	if status, ok := fields["status"].(cadence.Enum); ok {
		statusFields := cadence.FieldsMappedByName(status)
		if rawValue, ok := statusFields["rawValue"].(cadence.UInt8); ok {
			result.Status = uint8(rawValue)
		}
	}

	// Fees (UFix64)
	if fees, ok := fields["fees"].(cadence.UFix64); ok {
		result.Fees = fees.String()
	}

	// Scheduled Timestamp (UFix64)
	if timestamp, ok := fields["scheduledTimestamp"].(cadence.UFix64); ok {
		result.ScheduledTimestamp = timestamp.String()
	}

	// Handler Type Identifier (String)
	if handlerType, ok := fields["handlerTypeIdentifier"].(cadence.String); ok {
		result.HandlerTypeIdentifier = string(handlerType)
	}

	// Handler Address (Address)
	if handlerAddr, ok := fields["handlerAddress"].(cadence.Address); ok {
		result.HandlerAddress = handlerAddr.String()
	}

	return result, nil
}

// GetStatusString converts status code to readable string
func GetStatusString(status uint8) string {
	switch status {
	case 0:
		return "Pending"
	case 1:
		return "Scheduled"
	case 2:
		return "Executing"
	case 3:
		return "Executed"
	case 4:
		return "Failed"
	case 5:
		return "Cancelled"
	default:
		return fmt.Sprintf("Unknown(%d)", status)
	}
}

// GetPriorityString converts priority code to readable string
func GetPriorityString(priority uint8) string {
	switch priority {
	case 0:
		return "Low"
	case 1:
		return "Medium"
	case 2:
		return "High"
	default:
		return fmt.Sprintf("Unknown(%d)", priority)
	}
}

/*
 * Flow CLI
 *
 * Copyright 2019 Dapper Labs, Inc.
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

package config

// Emulator defines the configuration for a Flow Emulator instance.
type Emulator struct {
	Name           string
	Port           int
	ServiceAccount string
}

type Emulators []Emulator

// DefaultEmulators gets all default emulators.
func DefaultEmulators() Emulators {
	return Emulators{DefaultEmulator()}
}

// DefaultEmulator gets default emulator.
func DefaultEmulator() Emulator {
	return Emulator{
		Name:           "default",
		ServiceAccount: DefaultEmulatorServiceAccountName,
		Port:           3569,
	}
}

// Default gets default emulator.
func (e Emulators) Default() *Emulator {
	for i := range e {
		if e[i].Name == DefaultEmulatorConfigName {
			return &e[i]
		}
	}

	return nil
}

// AddOrUpdate add new or update if already present.
func (e *Emulators) AddOrUpdate(name string, emulator Emulator) {
	for i, existingEmulator := range *e {
		if existingEmulator.Name == name {
			(*e)[i] = emulator
			return
		}
	}

	*e = append(*e, emulator)
}

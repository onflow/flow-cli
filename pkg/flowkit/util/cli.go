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

// Package cli defines constants, configurations, and utilities that are used across the Flow CLI.
package util

import (
	"crypto/rand"
	"fmt"
	"os"
)

const (
	EnvPrefix = "FLOW"
)

func Exit(code int, msg string) {
	fmt.Println(msg)
	os.Exit(code)
}

func RandomSeed(n int) ([]byte, error) {
	seed := make([]byte, n)

	_, err := rand.Read(seed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random seed: %v", err)
	}

	return seed, nil
}

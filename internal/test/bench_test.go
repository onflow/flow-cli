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

package test

import (
	"fmt"
	"testing"

	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/accounts"
	"github.com/onflow/flowkit/v2/tests"
)

func buildTestFiles(n int) map[string][]byte {
	script := tests.TestScriptSimple
	files := make(map[string][]byte, n)
	for i := range n {
		files[fmt.Sprintf("test_%02d_%s", i, script.Filename)] = script.Source
	}
	return files
}

func BenchmarkTestCode_NFiles(b *testing.B) {
	rw, _ := tests.ReaderWriter()
	state, err := flowkit.Init(rw)
	if err != nil {
		b.Fatal(err)
	}
	emulatorAccount, _ := accounts.NewEmulatorAccount(rw, crypto.ECDSA_P256, crypto.SHA3_256, "")
	state.Accounts().AddOrUpdate(emulatorAccount)
	testFiles := buildTestFiles(10)

	for b.Loop() {
		_, err := testCode(testFiles, state, flagsTests{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

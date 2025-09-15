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

package command

import (
	"errors"
	"strings"
	"testing"
)

func TestIsMissingSignerKeyFile(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"unrelated", errors.New("some other error"), false},
		{"missing pkey file from open", errors.New("open testnet.pkey: no such file or directory"), true},
		{"missing file with provided location", errors.New("could not load the key for the account from provided location testnet.pkey: open testnet.pkey: no such file or directory"), true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isMissingSignerKeyFile(tc.err)
			if got != tc.want {
				t.Fatalf("isMissingSignerKeyFile() = %v, want %v for %q", got, tc.want, tc.name)
			}
		})
	}
}

func TestBuildMissingSignerKeySuggestion(t *testing.T) {
	msg := buildMissingSignerKeySuggestion(errors.New("open testnet.pkey: no such file or directory"))
	if !strings.Contains(msg, "testnet.pkey") {
		t.Fatalf("expected suggestion to include file path, got: %s", msg)
	}

	msg2 := buildMissingSignerKeySuggestion(errors.New("some error without file"))
	if !strings.Contains(strings.ToLower(msg2), "missing signer private key file") {
		t.Fatalf("expected generic suggestion, got: %s", msg2)
	}
}

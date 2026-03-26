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

package mcp

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Severity represents the severity level of a code review finding.
type Severity string

const (
	SeverityWarning Severity = "warning"
	SeverityNote    Severity = "note"
	SeverityInfo    Severity = "info"
)

// Finding represents a single code review finding.
type Finding struct {
	Rule     string   `json:"rule"`
	Severity Severity `json:"severity"`
	Line     int      `json:"line"`
	Message  string   `json:"message"`
}

// ReviewResult holds all findings from a code review along with a summary.
type ReviewResult struct {
	Findings []Finding      `json:"findings"`
	Summary  map[string]int `json:"summary"`
}

// reviewRule defines a single regex-based code review rule.
type reviewRule struct {
	id       string
	severity Severity
	pattern  *regexp.Regexp
	message  func(match []string) string
}

var addressImportPattern = regexp.MustCompile(`^\s*import\s+\w[\w, ]*\s+from\s+0x`)

var reviewRules = []reviewRule{
	{
		id:       "overly-permissive-access",
		severity: SeverityWarning,
		pattern:  regexp.MustCompile(`access\(all\)\s+(var|let)\s+`),
		message: func(_ []string) string {
			return "State field with access(all) — consider restricting access with entitlements"
		},
	},
	{
		id:       "overly-permissive-function",
		severity: SeverityNote,
		pattern:  regexp.MustCompile(`access\(all\)\s+fun\s+(\w+)`),
		message: func(match []string) string {
			name := ""
			if len(match) > 1 {
				name = match[1]
			}
			return fmt.Sprintf("Function '%s' has access(all) — review if public access is intended", name)
		},
	},
	{
		id:       "deprecated-pub",
		severity: SeverityInfo,
		pattern:  regexp.MustCompile(`\bpub\s+(var|let|fun|resource|struct|event|contract|enum)\b`),
		message: func(_ []string) string {
			return "`pub` is deprecated in Cadence 1.0 — use `access(all)` or a more restrictive access modifier"
		},
	},
	{
		id:       "unsafe-force-unwrap",
		severity: SeverityNote,
		pattern:  regexp.MustCompile(`[)\w]\s*!`),
		message: func(_ []string) string {
			return "Force-unwrap (!) used — consider nil-coalescing (??) or optional binding for safer handling"
		},
	},
	{
		id:       "auth-account-exposure",
		severity: SeverityWarning,
		pattern:  regexp.MustCompile(`\bAuthAccount\b`),
		message: func(_ []string) string {
			return "AuthAccount reference found — passing AuthAccount gives full account access, use capabilities instead"
		},
	},
	{
		id:       "auth-reference-exposure",
		severity: SeverityWarning,
		pattern:  regexp.MustCompile(`\bauth\s*\(.*?\)\s*&Account\b`),
		message: func(_ []string) string {
			return "auth(…) &Account reference found — this grants broad account access, prefer scoped capabilities"
		},
	},
	{
		id:       "hardcoded-address",
		severity: SeverityInfo,
		pattern:  regexp.MustCompile(`0x[0-9a-fA-F]{8,16}\b`),
		message: func(_ []string) string {
			return "Hardcoded address detected — consider using named address imports for portability"
		},
	},
	{
		id:       "unguarded-capability",
		severity: SeverityWarning,
		pattern:  regexp.MustCompile(`\.publish\s*\(`),
		message: func(_ []string) string {
			return "Capability published — verify that proper entitlements guard this capability"
		},
	},
	{
		id:       "resource-loss-destroy",
		severity: SeverityWarning,
		pattern:  regexp.MustCompile(`destroy\s*\(`),
		message: func(_ []string) string {
			return "Explicit destroy call — ensure the resource is intentionally being destroyed and not lost"
		},
	},
}

// codeReview runs all rules against the provided Cadence source code and returns
// a ReviewResult with findings sorted by line number.
func codeReview(code string) ReviewResult {
	lines := strings.Split(code, "\n")
	var findings []Finding

	for lineIdx, line := range lines {
		lineNum := lineIdx + 1
		for _, rule := range reviewRules {
			// Special case: skip hardcoded-address on import-from-address lines.
			if rule.id == "hardcoded-address" && addressImportPattern.MatchString(line) {
				continue
			}

			match := rule.pattern.FindStringSubmatch(line)
			if match != nil {
				findings = append(findings, Finding{
					Rule:     rule.id,
					Severity: rule.severity,
					Line:     lineNum,
					Message:  rule.message(match),
				})
			}
		}
	}

	sort.Slice(findings, func(i, j int) bool {
		return findings[i].Line < findings[j].Line
	})

	summary := map[string]int{
		string(SeverityWarning): 0,
		string(SeverityNote):    0,
		string(SeverityInfo):    0,
	}
	for _, f := range findings {
		summary[string(f.Severity)]++
	}

	return ReviewResult{
		Findings: findings,
		Summary:  summary,
	}
}

// formatReviewResult formats a ReviewResult as human-readable text.
func formatReviewResult(result ReviewResult) string {
	if len(result.Findings) == 0 {
		return "No findings.\n"
	}

	var sb strings.Builder
	for _, f := range result.Findings {
		sb.WriteString(fmt.Sprintf("[%s] line %d (%s): %s\n", f.Severity, f.Line, f.Rule, f.Message))
	}

	sb.WriteString(fmt.Sprintf("\nSummary: %d warning(s), %d note(s), %d info(s)\n",
		result.Summary[string(SeverityWarning)],
		result.Summary[string(SeverityNote)],
		result.Summary[string(SeverityInfo)],
	))

	return sb.String()
}

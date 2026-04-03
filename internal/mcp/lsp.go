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
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/onflow/cadence-tools/languageserver/integration"
	"github.com/onflow/cadence-tools/languageserver/protocol"
	"github.com/onflow/cadence-tools/languageserver/server"
)

const scratchURI = protocol.DocumentURI("file:///mcp/scratch.cdc")

// diagConn implements protocol.Conn and captures diagnostics published by the LSP server.
type diagConn struct {
	mu          sync.Mutex
	diagnostics []protocol.Diagnostic
}

func (c *diagConn) Notify(method string, params any) error {
	if method == "textDocument/publishDiagnostics" {
		switch p := params.(type) {
		case *protocol.PublishDiagnosticsParams:
			c.captureDiagnostics(p.URI, p.Diagnostics)
		default:
			// Try JSON round-trip for map types
			data, err := json.Marshal(p)
			if err == nil {
				var pdp protocol.PublishDiagnosticsParams
				if json.Unmarshal(data, &pdp) == nil {
					c.captureDiagnostics(pdp.URI, pdp.Diagnostics)
				}
			}
		}
	}
	return nil
}

func (c *diagConn) ShowMessage(_ *protocol.ShowMessageParams) {}

func (c *diagConn) ShowMessageRequest(_ *protocol.ShowMessageRequestParams) (*protocol.MessageActionItem, error) {
	return nil, nil
}

func (c *diagConn) LogMessage(_ *protocol.LogMessageParams) {}

func (c *diagConn) PublishDiagnostics(params *protocol.PublishDiagnosticsParams) error {
	if params != nil {
		c.captureDiagnostics(params.URI, params.Diagnostics)
	}
	return nil
}

func (c *diagConn) RegisterCapability(_ *protocol.RegistrationParams) error {
	return nil
}

func (c *diagConn) captureDiagnostics(uri protocol.DocumentURI, diags []protocol.Diagnostic) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if uri != "" && uri != scratchURI {
		return // ignore diagnostics for unrelated documents
	}
	c.diagnostics = diags // replace, not append — one publish per check cycle
}

func (c *diagConn) reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.diagnostics = nil
}

func (c *diagConn) getDiagnostics() []protocol.Diagnostic {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]protocol.Diagnostic, len(c.diagnostics))
	copy(result, c.diagnostics)
	return result
}

// LSPWrapper manages an in-process cadence-tools LSP server,
// handling document lifecycle and diagnostic capture.
type LSPWrapper struct {
	server     *server.Server
	conn       *diagConn
	mu         sync.Mutex
	docVersion int32
	docOpen    bool
}

// NewLSPWrapper creates a new LSP wrapper with an in-process Cadence language server.
func NewLSPWrapper(enableFlowClient bool) (*LSPWrapper, error) {
	s, err := server.NewServer()
	if err != nil {
		return nil, fmt.Errorf("creating LSP server: %w", err)
	}

	_, err = integration.NewFlowIntegration(s, enableFlowClient)
	if err != nil {
		return nil, fmt.Errorf("creating flow integration: %w", err)
	}

	conn := &diagConn{}

	_, err = s.Initialize(conn, &protocol.InitializeParams{
		XInitializeParams: protocol.XInitializeParams{
			InitializationOptions: map[string]any{
				"accessCheckMode": "strict",
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("initializing LSP server: %w", err)
	}

	return &LSPWrapper{
		server: s,
		conn:   conn,
	}, nil
}

// updateDocument sends the code to the LSP server as a virtual document.
// Must be called with w.mu held.
func (w *LSPWrapper) updateDocument(code string) error {
	w.docVersion++
	version := w.docVersion

	if !w.docOpen {
		w.docOpen = true
		return w.server.DidOpenTextDocument(w.conn, &protocol.DidOpenTextDocumentParams{
			TextDocument: protocol.TextDocumentItem{
				URI:        scratchURI,
				LanguageID: "cadence",
				Version:    version,
				Text:       code,
			},
		})
	}

	return w.server.DidChangeTextDocument(w.conn, &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{
				URI: scratchURI,
			},
			Version: version,
		},
		ContentChanges: []protocol.TextDocumentContentChangeEvent{
			{Text: code},
		},
	})
}

// Check sends code to the LSP and returns any diagnostics.
func (w *LSPWrapper) Check(code string) ([]protocol.Diagnostic, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.conn.reset()

	if err := w.updateDocument(code); err != nil {
		return nil, fmt.Errorf("updating document: %w", err)
	}

	return w.conn.getDiagnostics(), nil
}

// Hover returns hover information at the given position.
func (w *LSPWrapper) Hover(code string, line, character int) (*protocol.Hover, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.conn.reset()

	if err := w.updateDocument(code); err != nil {
		return nil, fmt.Errorf("updating document: %w", err)
	}

	return w.server.Hover(w.conn, &protocol.TextDocumentPositionParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: scratchURI},
		Position:     protocol.Position{Line: uint32(line), Character: uint32(character)},
	})
}

// Definition returns the definition location for the symbol at the given position.
func (w *LSPWrapper) Definition(code string, line, character int) (*protocol.Location, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.conn.reset()

	if err := w.updateDocument(code); err != nil {
		return nil, fmt.Errorf("updating document: %w", err)
	}

	return w.server.Definition(w.conn, &protocol.TextDocumentPositionParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: scratchURI},
		Position:     protocol.Position{Line: uint32(line), Character: uint32(character)},
	})
}

// Symbols returns the document symbols for the given code.
func (w *LSPWrapper) Symbols(code string) ([]*protocol.DocumentSymbol, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.conn.reset()

	if err := w.updateDocument(code); err != nil {
		return nil, fmt.Errorf("updating document: %w", err)
	}

	return w.server.DocumentSymbol(w.conn, &protocol.DocumentSymbolParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: scratchURI},
	})
}

// Completion returns completion items at the given position.
func (w *LSPWrapper) Completion(code string, line, character int) ([]*protocol.CompletionItem, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.conn.reset()

	if err := w.updateDocument(code); err != nil {
		return nil, fmt.Errorf("updating document: %w", err)
	}

	return w.server.Completion(w.conn, &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: scratchURI},
			Position:     protocol.Position{Line: uint32(line), Character: uint32(character)},
		},
	})
}

// formatDiagnostics formats diagnostics as human-readable text.
func formatDiagnostics(diagnostics []protocol.Diagnostic) string {
	if len(diagnostics) == 0 {
		return "No errors found."
	}

	var b strings.Builder
	for _, d := range diagnostics {
		fmt.Fprintf(&b, "[%s] line %d:%d: %s\n",
			d.Severity.String(),
			d.Range.Start.Line+1,
			d.Range.Start.Character+1,
			d.Message,
		)
	}
	return b.String()
}

// formatHover formats a hover result as human-readable text.
func formatHover(result *protocol.Hover) string {
	if result == nil {
		return "No hover information available."
	}
	return result.Contents.Value
}

// formatSymbols formats document symbols as an indented tree.
// Accepts []*protocol.DocumentSymbol (from the server API).
func formatSymbols(symbols []*protocol.DocumentSymbol, indent int) string {
	var b strings.Builder
	prefix := strings.Repeat("  ", indent)
	for _, s := range symbols {
		fmt.Fprintf(&b, "%s%s %s", prefix, s.Kind.String(), s.Name)
		if s.Detail != "" {
			fmt.Fprintf(&b, " — %s", s.Detail)
		}
		b.WriteString("\n")
		if len(s.Children) > 0 {
			b.WriteString(formatSymbolValues(s.Children, indent+1))
		}
	}
	return b.String()
}

// formatSymbolValues formats []protocol.DocumentSymbol (value type, used for Children).
func formatSymbolValues(symbols []protocol.DocumentSymbol, indent int) string {
	var b strings.Builder
	prefix := strings.Repeat("  ", indent)
	for _, s := range symbols {
		fmt.Fprintf(&b, "%s%s %s", prefix, s.Kind.String(), s.Name)
		if s.Detail != "" {
			fmt.Fprintf(&b, " — %s", s.Detail)
		}
		b.WriteString("\n")
		if len(s.Children) > 0 {
			b.WriteString(formatSymbolValues(s.Children, indent+1))
		}
	}
	return b.String()
}

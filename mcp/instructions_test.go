package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

type buildInstructionsTest struct {
	name       string
	base       string
	clientInfo string // raw initialize params; empty = paramless connect
	wantQuirks []string
}

var buildInstructionsCases = []buildInstructionsTest{
	{
		name:       "OpencodeGetsQuirkSection",
		base:       "BASE",
		clientInfo: `{"clientInfo":{"name":"opencode-go","version":"1.0"}}`,
		wantQuirks: []string{"Known harness quirks (opencode)"},
	},
	{
		name:       "UnknownHarnessGetsBaseOnly",
		base:       "BASE",
		clientInfo: `{"clientInfo":{"name":"claude-code","version":"2.0"}}`,
		wantQuirks: []string{},
	},
	{
		name:       "NoClientInfoGetsBaseOnly",
		base:       "BASE",
		clientInfo: `{}`,
		wantQuirks: []string{},
	},
	{
		name:       "MatchIsCaseInsensitiveSubstring",
		base:       "BASE",
		clientInfo: `{"clientInfo":{"name":"OpenCode-Go-Upstream-Fork"}}`,
		wantQuirks: []string{"Known harness quirks (opencode)"},
	},
}

func TestBuildInstructions(t *testing.T) {
	for _, tt := range buildInstructionsCases {
		t.Run(tt.name, func(t *testing.T) {
			var req struct {
				ClientInfo struct {
					Name string `json:"name"`
				} `json:"clientInfo"`
			}
			json.Unmarshal([]byte(tt.clientInfo), &req)

			got := buildInstructions(tt.base, req.ClientInfo.Name)

			if !strings.HasPrefix(got, tt.base) {
				t.Errorf("base instructions must always be included, got %q", got)
			}
			if tt.wantQuirks == nil {
				return
			}
			for _, want := range tt.wantQuirks {
				if !strings.Contains(got, want) {
					t.Errorf("expected instructions to contain %q, got %q", want, got)
				}
			}
		})
	}
}

func TestBaseInstructionsContents(t *testing.T) {
	// the columns doc is load-bearing context for the model; guard it
	for _, want := range []string{"## Columns", "- UID:", "- SCHEDULED:", "single source of truth"} {
		if !strings.Contains(baseInstructions, want) {
			t.Errorf("baseInstructions missing %q", want)
		}
	}
}

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtraStore_SaveRuleKeepsTitle(t *testing.T) {
	dir, err := os.MkdirTemp("", "tcode_extra_rule_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	s := &ExtraStore{
		baseDir:   dir,
		mcpFile:   filepath.Join(dir, "mcp.json"),
		skillFile: filepath.Join(dir, "skills.json"),
		ruleFile:  filepath.Join(dir, "rules.json"),
		mcps:      []MCPServerConfig{},
		skills:    []SkillConfig{},
		rules:     []RuleConfig{},
	}
	if err := s.SaveRule(RuleConfig{ID: "rule_1", Title: "铁律", Content: "no fake", Enabled: true}); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveRule(RuleConfig{ID: "rule_1", Title: "铁律改", Content: "still no fake", Enabled: true}); err != nil {
		t.Fatal(err)
	}
	list := s.ListRules()
	if len(list) != 1 {
		t.Fatalf("got %d rules", len(list))
	}
	if list[0].Title != "铁律改" || list[0].Content != "still no fake" {
		t.Fatalf("%+v", list[0])
	}
}

func TestExtraStore_MCPUpsertAndDelete(t *testing.T) {
	dir, err := os.MkdirTemp("", "tcode_extra_mcp_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	s := &ExtraStore{
		baseDir:   dir,
		mcpFile:   filepath.Join(dir, "mcp.json"),
		skillFile: filepath.Join(dir, "skills.json"),
		ruleFile:  filepath.Join(dir, "rules.json"),
		mcps:      []MCPServerConfig{},
		skills:    []SkillConfig{},
		rules:     []RuleConfig{},
	}
	if err := s.SaveMCP(MCPServerConfig{ID: "mcp_1", Name: "fs", Command: "npx", Enabled: true}); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveMCP(MCPServerConfig{ID: "mcp_1", Name: "fs2", Command: "npx", Enabled: false}); err != nil {
		t.Fatal(err)
	}
	if len(s.ListMCPs()) != 1 || s.ListMCPs()[0].Name != "fs2" {
		t.Fatalf("%+v", s.ListMCPs())
	}
	if err := s.DeleteMCP("mcp_1"); err != nil {
		t.Fatal(err)
	}
	if len(s.ListMCPs()) != 0 {
		t.Fatalf("expected empty after delete")
	}
}

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtraStore_UIPrefsRoundTrip(t *testing.T) {
	dir, err := os.MkdirTemp("", "tcode_ui_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	s := &ExtraStore{baseDir: dir}
	if err := s.SaveUIPrefs(UIPrefs{Theme: "dark", MonacoFont: "Fira Code", MonacoSize: 16}); err != nil {
		t.Fatal(err)
	}
	got := s.GetUIPrefs()
	if got.Theme != "dark" || got.MonacoFont != "Fira Code" || got.MonacoSize != 16 {
		t.Fatalf("%+v", got)
	}
	if _, err := os.Stat(filepath.Join(dir, "ui_prefs.json")); err != nil {
		t.Fatal(err)
	}
}

func TestExtraStore_ImportWorkspaceRules(t *testing.T) {
	root, err := os.MkdirTemp("", "tcode_ws_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)
	dir, err := os.MkdirTemp("", "tcode_rules_*")
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
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("# agents"), 0644); err != nil {
		t.Fatal(err)
	}
	n, err := s.ImportWorkspaceRules(root)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("imported %d", n)
	}
	list := s.ListRules()
	if len(list) != 1 || list[0].Title != "AGENTS.md" {
		t.Fatalf("%+v", list)
	}
}

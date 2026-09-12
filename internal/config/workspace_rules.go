package config

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

var workspaceRuleFiles = []string{".cursorrules", "AGENTS.md", "CLAUDE.md"}

// ImportWorkspaceRules reads known rule files from a workspace and upserts them as RuleConfig.
func (s *ExtraStore) ImportWorkspaceRules(root string) (int, error) {
	if s == nil {
		return 0, os.ErrInvalid
	}
	n := 0
	for _, name := range workspaceRuleFiles {
		p := filepath.Join(root, name)
		b, err := os.ReadFile(p)
		if err != nil || len(strings.TrimSpace(string(b))) == 0 {
			continue
		}
		id := "ws_" + strings.ReplaceAll(strings.ToLower(name), ".", "_")
		if err := s.SaveRule(RuleConfig{
			ID:        id,
			Title:     name,
			Content:   string(b),
			Scope:     "workspace",
			Enabled:   true,
			UpdatedAt: time.Now().Unix(),
		}); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

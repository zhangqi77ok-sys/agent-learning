package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type UIPrefs struct {
	Theme      string `json:"theme"`
	MonacoFont string `json:"monaco_font"`
	MonacoSize int    `json:"monaco_size"`
}

func DefaultUIPrefs() UIPrefs {
	return UIPrefs{Theme: "warm", MonacoFont: "JetBrains Mono", MonacoSize: 14}
}

func (s *ExtraStore) prefsPath() string {
	if s == nil {
		return filepath.Join(UserDataDir(), "ui_prefs.json")
	}
	return filepath.Join(s.baseDir, "ui_prefs.json")
}

func (s *ExtraStore) GetUIPrefs() UIPrefs {
	p := DefaultUIPrefs()
	path := s.prefsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return p
	}
	_ = json.Unmarshal(data, &p)
	if p.Theme == "" {
		p.Theme = "warm"
	}
	if p.MonacoFont == "" {
		p.MonacoFont = "JetBrains Mono"
	}
	if p.MonacoSize < 10 {
		p.MonacoSize = 14
	}
	return p
}

func (s *ExtraStore) SaveUIPrefs(p UIPrefs) error {
	if p.Theme == "" {
		p.Theme = "warm"
	}
	if p.MonacoFont == "" {
		p.MonacoFont = "JetBrains Mono"
	}
	if p.MonacoSize < 10 {
		p.MonacoSize = 14
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return atomicWriteConfig(s.prefsPath(), data)
}

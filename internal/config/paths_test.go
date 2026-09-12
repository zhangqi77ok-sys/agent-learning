package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMigrateLegacyDir(t *testing.T) {
	root := t.TempDir()
	legacy := filepath.Join(root, ".tcode")
	dest := filepath.Join(root, ".tiancode")
	_ = os.MkdirAll(legacy, 0755)
	if err := os.WriteFile(filepath.Join(legacy, "channels.json"), []byte(`[]`), 0644); err != nil {
		t.Fatal(err)
	}
	migrateLegacyDir(legacy, dest)
	if _, err := os.Stat(filepath.Join(dest, "channels.json")); err != nil {
		t.Fatal(err)
	}
}

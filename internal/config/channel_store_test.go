package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChannelStore_ZeroDemo_CleanEmptyState(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tcode_test_channels_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store := &ChannelStore{
		filePath: filepath.Join(tempDir, "channels.json"),
		channels: make([]ChannelConfig, 0),
	}
	_ = store.load()

	list := store.List()
	if len(list) != 0 {
		t.Fatalf("expected 0 channels on clean init, got %d", len(list))
	}

	primary := store.GetPrimary()
	if primary != nil {
		t.Fatalf("expected nil primary channel on clean init, got %+v", primary)
	}
}

func TestChannelStore_SaveUpdatesSameID(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tcode_test_ch_upd_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	store := &ChannelStore{
		filePath: filepath.Join(tempDir, "channels.json"),
		channels: make([]ChannelConfig, 0),
	}
	if err := store.Save(ChannelConfig{ID: "ch_1", Name: "a", Endpoint: "https://x", APIKey: "k1", Model: "m1"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(ChannelConfig{ID: "ch_1", Name: "b", Endpoint: "https://y", APIKey: "k2", Model: "m2"}); err != nil {
		t.Fatal(err)
	}
	list := store.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 channel after update, got %d", len(list))
	}
	if list[0].Name != "b" || list[0].Endpoint != "https://y" || list[0].Model != "m2" {
		t.Fatalf("not updated: %+v", list[0])
	}
}

func TestAtomicWriteConfig_CreatesParentDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tcode_test_extra_cfg_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	subDir := filepath.Join(tempDir, "a", "b", "c")
	cfgFile := filepath.Join(subDir, "config.json")
	if err := atomicWriteConfig(cfgFile, []byte(`{"ok": true}`)); err != nil {
		t.Fatalf("atomicWriteConfig failed to create parent dir: %v", err)
	}

	data, err := os.ReadFile(cfgFile)
	if err != nil {
		t.Fatalf("failed to read back: %v", err)
	}
	if string(data) != `{"ok": true}` {
		t.Errorf("unexpected content: %s", string(data))
	}
}

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestADRStore_SaveLoad(t *testing.T) {
	dir, err := os.MkdirTemp("", "tcode_adr_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	s := NewADRStore(filepath.Join(dir, "adr.json"))
	if n := s.Get("missing"); n != "" {
		t.Fatalf("empty: %q", n)
	}
	if err := s.Save("pkg.Foo", "包级入口，禁止假数据"); err != nil {
		t.Fatal(err)
	}
	if s.Get("pkg.Foo") != "包级入口，禁止假数据" {
		t.Fatalf("got %q", s.Get("pkg.Foo"))
	}
	s2 := NewADRStore(filepath.Join(dir, "adr.json"))
	if s2.Get("pkg.Foo") != "包级入口，禁止假数据" {
		t.Fatalf("reload %q", s2.Get("pkg.Foo"))
	}
	if err := s2.Save("pkg.Foo", ""); err != nil {
		t.Fatal(err)
	}
	if s2.Get("pkg.Foo") != "" {
		t.Fatal("expected delete on empty note")
	}
}

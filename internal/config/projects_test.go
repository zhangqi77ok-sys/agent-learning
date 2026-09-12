package config

import (
	"path/filepath"
	"testing"
)

func TestProjectStore_AddListRemove(t *testing.T) {
	dir := t.TempDir()
	s := &ProjectStore{
		filePath: filepath.Join(dir, "projects.json"),
		projects: make([]Project, 0),
	}
	if err := s.Add(filepath.Join(dir, "alpha")); err != nil {
		t.Fatal(err)
	}
	if err := s.Add(filepath.Join(dir, "beta")); err != nil {
		t.Fatal(err)
	}
	list := s.List()
	if len(list) != 2 {
		t.Fatalf("got %d", len(list))
	}
	if err := s.Remove(filepath.Join(dir, "alpha")); err != nil {
		t.Fatal(err)
	}
	if len(s.List()) != 1 || s.List()[0].Name != "beta" {
		t.Fatalf("%+v", s.List())
	}
}

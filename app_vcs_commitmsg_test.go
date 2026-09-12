package main

import (
	"strings"
	"testing"
)

func TestConventionalCommitFromPorcelain(t *testing.T) {
	msg := conventionalCommitFromPorcelain(" M app.go\n?? new.go\n")
	if !strings.Contains(msg, "app.go") || !strings.Contains(msg, "new.go") {
		t.Fatalf("got %q", msg)
	}
	if !strings.HasPrefix(msg, "feat:") && !strings.HasPrefix(msg, "fix:") {
		t.Fatalf("expected conventional prefix, got %q", msg)
	}
	empty := conventionalCommitFromPorcelain("   \n")
	if empty != "chore: update workspace" {
		t.Fatalf("empty: %q", empty)
	}
}

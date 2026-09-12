package search

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"tiancode/internal/core/sandbox"
)

func TestSearchTool_GrepAndFind(t *testing.T) {
	tempDir := t.TempDir()
	sb, err := sandbox.NewSandbox(tempDir)
	if err != nil {
		t.Fatalf("failed to create sandbox: %v", err)
	}

	// 准备测试文件
	subDir := filepath.Join(tempDir, "src", "core")
	_ = os.MkdirAll(subDir, 0755)

	file1 := filepath.Join(tempDir, "main.go")
	_ = os.WriteFile(file1, []byte("package main\n\nfunc main() {\n\tprintln(\"Hello Tiancode\")\n}\n"), 0644)

	file2 := filepath.Join(subDir, "engine.go")
	_ = os.WriteFile(file2, []byte("package core\n\n// StartEngine initializes loop\nfunc StartEngine() {\n\tprintln(\"Engine running\")\n}\n"), 0644)

	// 准备应被忽略的目录
	ignoredDir := filepath.Join(tempDir, "node_modules", "pkg")
	_ = os.MkdirAll(ignoredDir, 0755)
	_ = os.WriteFile(filepath.Join(ignoredDir, "index.js"), []byte("Hello Tiancode in node_modules"), 0644)

	tool := NewTool(sb)

	// 1. 测试 Grep: 搜索 "Tiancode"
	grepArgs, _ := json.Marshal(map[string]any{
		"action": "grep",
		"query":  "Tiancode",
	})
	res, err := tool.Execute(context.Background(), grepArgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success, got error: %s", res.Content)
	}
	if !strings.Contains(res.Content, "main.go:4:") {
		t.Errorf("expected grep to find main.go:4, got:\n%s", res.Content)
	}
	if strings.Contains(res.Content, "node_modules") {
		t.Errorf("expected node_modules to be ignored, got:\n%s", res.Content)
	}

	// 2. 测试 Find: 搜索 "*.go" 文件
	findArgs, _ := json.Marshal(map[string]any{
		"action": "find",
		"query":  "*.go",
	})
	resFind, err := tool.Execute(context.Background(), findArgs)
	if err != nil || resFind.IsError {
		t.Fatalf("find failed: %v, content: %s", err, resFind.Content)
	}
	if !strings.Contains(resFind.Content, "main.go") || !strings.Contains(resFind.Content, "engine.go") {
		t.Errorf("expected find to return main.go and engine.go, got:\n%s", resFind.Content)
	}

	// 3. 测试空参数防御
	emptyArgs, _ := json.Marshal(map[string]any{
		"action": "grep",
		"query":  "",
	})
	resEmpty, _ := tool.Execute(context.Background(), emptyArgs)
	if !resEmpty.IsError {
		t.Errorf("expected error on empty query")
	}
}

package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunTDDValidation_NoTestSuite(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tcode_test_agent_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	report, err := RunTDDValidation(tempDir)
	if err != nil {
		t.Fatalf("RunTDDValidation failed: %v", err)
	}
	// 无测试套件时，根据铁律与达标标准，严禁返回假 PASS 假成功
	if report.Status != "FAIL" {
		t.Errorf("expected FAIL when no test suite is configured, got %s", report.Status)
	}
	if !strings.Contains(report.Output, "未检测到可执行的自动化测试套件") {
		t.Errorf("expected failure explanation in output, got: %s", report.Output)
	}
}

func TestRunTDDValidation_WithGoMod(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tcode_test_agent_gomod_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	_ = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module testmod\n\ngo 1.22\n"), 0644)
	report, err := RunTDDValidation(tempDir)
	if err != nil {
		t.Fatalf("RunTDDValidation failed: %v", err)
	}
	if report.Status != "PASS" && report.Status != "FAIL" {
		t.Errorf("expected PASS or FAIL, got: %s", report.Status)
	}
}

func TestRunTDDValidation_WithPackageJson(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tcode_test_agent_pkg_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	pkgContent := `{"name": "testpkg", "scripts": {"test": "echo test executed"}}`
	_ = os.WriteFile(filepath.Join(tempDir, "package.json"), []byte(pkgContent), 0644)
	report, err := RunTDDValidation(tempDir)
	if err != nil {
		t.Fatalf("RunTDDValidation failed: %v", err)
	}
	if report.Status != "PASS" && report.Status != "FAIL" {
		t.Errorf("expected PASS or FAIL, got: %s", report.Status)
	}
}

func TestRunTDDValidation_HybridGoAndFrontend(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tcode_test_agent_hybrid_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	_ = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module testhybrid\n\ngo 1.22\n"), 0644)
	frontendDir := filepath.Join(tempDir, "frontend")
	_ = os.MkdirAll(frontendDir, 0755)
	pkgContent := `{"name": "frontend-ui", "scripts": {"test": "echo frontend test passed"}}`
	_ = os.WriteFile(filepath.Join(frontendDir, "package.json"), []byte(pkgContent), 0644)

	report, err := RunTDDValidation(tempDir)
	if err != nil {
		t.Fatalf("RunTDDValidation failed: %v", err)
	}
	if report.Status != "PASS" && report.Status != "FAIL" {
		t.Errorf("expected PASS or FAIL, got: %s", report.Status)
	}
	if !strings.Contains(report.Output, "Go Test 套件") || !strings.Contains(report.Output, "Npm Test 套件") {
		t.Errorf("expected output to contain both Go and Npm test sections, got: %s", report.Output)
	}
}


func TestRunSecurityAudit_Clean(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	repoRoot := filepath.Dir(filepath.Dir(wd))

	report, err := RunSecurityAudit(repoRoot)
	if err != nil {
		t.Fatalf("RunSecurityAudit failed: %v", err)
	}
	if report.FilesScanned == 0 {
		t.Errorf("expected files scanned > 0, got %d", report.FilesScanned)
	}
}

func TestRunSecurityAudit_SkipLargeFiles(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tcode_test_audit_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建一个包含高危关键词但大于 5MB 的文件
	largePath := filepath.Join(tempDir, "big_script.py")
	f, err := os.Create(largePath)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	// 写入 6MB 数据并包含 high risk 关键字
	_ = f.Truncate(6 * 1024 * 1024)
	_, _ = f.WriteString("shutdown")
	_ = f.Close()

	report, err := RunSecurityAudit(tempDir)
	if err != nil {
		t.Fatalf("RunSecurityAudit failed: %v", err)
	}
	if len(report.Issues) > 0 {
		t.Errorf("expected 0 issues due to file size > 5MB, got %d", len(report.Issues))
	}
}

func TestRunSecurityAudit_DetectExecCommandKeyword(t *testing.T) {
	tempDir := t.TempDir()
	sampleFile := filepath.Join(tempDir, "danger.go")
	content := []byte(`package main
import "os/exec"
func dangerous() {
	exec.Command("cmd", "/c", "del /f /q *")
}
`)
	_ = os.WriteFile(sampleFile, content, 0644)

	report, err := RunSecurityAudit(tempDir)
	if err != nil {
		t.Fatalf("RunSecurityAudit failed: %v", err)
	}
	if len(report.Issues) == 0 {
		t.Fatalf("expected to detect exec.Command keyword, but 0 issues found")
	}
}

func TestRunSecurityAudit_WorkspaceNamedBuild(t *testing.T) {
	tempDir := t.TempDir()
	wsNamedBuild := filepath.Join(tempDir, "build")
	if err := os.MkdirAll(wsNamedBuild, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	sampleFile := filepath.Join(wsNamedBuild, "app.go")
	_ = os.WriteFile(sampleFile, []byte("package main\nfunc main() {}\n"), 0644)

	report, err := RunSecurityAudit(wsNamedBuild)
	if err != nil {
		t.Fatalf("RunSecurityAudit failed: %v", err)
	}
	if report.FilesScanned == 0 {
		t.Fatalf("expected files scanned > 0 even when workspace is named 'build', got 0")
	}
}

func TestExtractFailedTests(t *testing.T) {
	goOut := `=== RUN   TestEngine_Pass
--- PASS: TestEngine_Pass (0.01s)
=== RUN   TestEngine_FailOne
--- FAIL: TestEngine_FailOne (0.02s)
=== RUN   TestEngine_FailTwo
FAIL:	TestEngine_FailTwo
FAIL
FAIL	github.com/tiancode/tcode/internal/loop	0.05s
`
	npmOut := `
FAIL src/components/Chat.test.ts
  ● Chat > should send message
    expect(received).toBe(expected)
  ✕ should handle error properly (25ms)
  ✖ should auto scroll
`
	failed := ExtractFailedTests(goOut, npmOut)
	if len(failed) != 6 {
		t.Fatalf("expected 6 failed tests extracted, got %d: %v", len(failed), failed)
	}

	expected := map[string]bool{
		"TestEngine_FailOne":           true,
		"TestEngine_FailTwo":           true,
		"src/components/Chat.test.ts":  true,
		"Chat > should send message":   true,
		"should handle error properly": true,
		"should auto scroll":           true,
	}

	for _, name := range failed {
		if !expected[name] {
			t.Errorf("unexpected failed test name: %s", name)
		}
	}
}



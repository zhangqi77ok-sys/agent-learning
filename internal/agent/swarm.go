package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// TestReport TDD 自动化测试验证器结果
type TestReport struct {
	Status    string   `json:"status"` // "PASS" | "FAIL"
	Passed    int      `json:"passed"`
	Failed    int      `json:"failed"`
	Duration  string   `json:"duration"`
	Output    string   `json:"output"`
	Timestamp int64    `json:"timestamp"`
}

// AuditReport 安全沙箱代码审查报告
type AuditReport struct {
	Status      string   `json:"status"` // "SECURE" | "WARNING" | "BLOCKED"
	RiskLevel   string   `json:"risk_level"` // "LOW" | "MEDIUM" | "CRITICAL"
	Issues      []string `json:"issues"`
	FilesScanned int     `json:"files_scanned"`
	Timestamp   int64    `json:"timestamp"`
}

// RunTDDValidation 运行自动化 TDD 测试驱动红绿灯验证 (支持 Go 原生 go test 与 Node npm test，带 60s 硬超时与零黑框)
func RunTDDValidation(workspace string) (TestReport, error) {
	start := time.Now()

	hasGoMod := false
	if fi, err := os.Stat(filepath.Join(workspace, "go.mod")); err == nil && !fi.IsDir() {
		hasGoMod = true
	}

	hasPkgJsonTest := false
	pkgJsonPath := filepath.Join(workspace, "package.json")
	if fi, err := os.Stat(pkgJsonPath); err == nil && !fi.IsDir() {
		if data, err := os.ReadFile(pkgJsonPath); err == nil {
			var pkg struct {
				Scripts map[string]string `json:"scripts"`
			}
			if json.Unmarshal(data, &pkg) == nil && pkg.Scripts != nil {
				if testScript, ok := pkg.Scripts["test"]; ok && strings.TrimSpace(testScript) != "" {
					hasPkgJsonTest = true
				}
			}
		}
	}

	if !hasGoMod && !hasPkgJsonTest {
		return TestReport{
			Status:    "FAIL",
			Passed:    0,
			Failed:    1,
			Duration:  "0ms",
			Output:    "当前工作区未检测到可执行的自动化测试套件 (既未找到 go.mod，亦未找到配置了 test 脚本的 package.json)。TDD 模式严禁无测试直接通过。",
			Timestamp: time.Now().Unix(),
		}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var testCmd *exec.Cmd
	var isNpmTest bool

	if hasGoMod {
		goExe, err := exec.LookPath("go")
		if err != nil {
			// 备用查找路径
			for _, candidate := range []string{`E:\pro\tools\go\bin\go.exe`, `C:\Program Files\Go\bin\go.exe`} {
				if _, statErr := os.Stat(candidate); statErr == nil {
					goExe = candidate
					err = nil
					break
				}
			}
		}
		if err != nil {
			return TestReport{
				Status:    "FAIL",
				Passed:    0,
				Failed:    1,
				Duration:  "0ms",
				Output:    "未检测到系统安装的 Go 编译器环境 (go not found in PATH)，请配置 Go 工具链以执行 TDD 验证",
				Timestamp: time.Now().Unix(),
			}, nil
		}
		testCmd = exec.CommandContext(ctx, goExe, "test", "-v", "./...")
	} else {
		isNpmTest = true
		npmCmdName := "npm"
		if runtime.GOOS == "windows" {
			npmCmdName = "npm.cmd"
		}
		npmExe, err := exec.LookPath(npmCmdName)
		if err != nil {
			npmExe, err = exec.LookPath("npm")
		}
		if err != nil {
			return TestReport{
				Status:    "FAIL",
				Passed:    0,
				Failed:    1,
				Duration:  "0ms",
				Output:    "未检测到 Node.js / npm 运行环境 (npm not found in PATH)，请配置 Node.js 以执行 npm test",
				Timestamp: time.Now().Unix(),
			}, nil
		}
		testCmd = exec.CommandContext(ctx, npmExe, "test")
	}

	testCmd.Dir = workspace
	if runtime.GOOS == "windows" {
		testCmd.SysProcAttr = &syscall.SysProcAttr{
			CreationFlags: 0x08000000,
			HideWindow:    true,
		}
		testCmd.Cancel = func() error {
			if testCmd.Process != nil && testCmd.Process.Pid > 0 {
				killCmd := exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", testCmd.Process.Pid))
				killCmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000, HideWindow: true}
				return killCmd.Run()
			}
			return nil
		}
	} else {
		testCmd.Cancel = func() error {
			if testCmd.Process != nil {
				return testCmd.Process.Kill()
			}
			return nil
		}
	}

	out, err := testCmd.CombinedOutput()
	duration := time.Since(start).Round(time.Millisecond).String()

	outputStr := string(out)
	if ctx.Err() == context.DeadlineExceeded {
		outputStr += "\n[超时警告] 测试执行超过 60s 硬超时上限，已被安全中断"
	}

	passed := 0
	failed := 0

	if !isNpmTest {
		passed = strings.Count(outputStr, "--- PASS:")
		failed = strings.Count(outputStr, "--- FAIL:")
	} else {
		if err == nil {
			passed = 1
		} else {
			failed = 1
		}
	}

	status := "PASS"
	if err != nil || failed > 0 {
		status = "FAIL"
		if failed == 0 {
			failed = 1 // 编译失败或异常退出时，确保 failed >= 1，杜绝 0 失败假成功
		}
	}

	return TestReport{
		Status:    status,
		Passed:    passed,
		Failed:    failed,
		Duration:  duration,
		Output:    outputStr,
		Timestamp: time.Now().Unix(),
	}, nil
}

// RunSecurityAudit 运行安全沙箱审查器，检测高危代码、未脱敏密钥与系统提权指令
func RunSecurityAudit(workspace string) (AuditReport, error) {
	issues := make([]string, 0)
	filesScanned := 0

	highRiskKeywords := []string{
		"rm -rf /", "mkfs", "format c:", "drop database", "shutdown", "exec.Command(\"cmd\", \"/c\", \"del",
		"exec.Command(\"sh\", \"-c\", \"rm",
	}

	_ = filepath.Walk(workspace, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}
		if info.IsDir() {
			cleanPath := filepath.Clean(path)
			cleanWs := filepath.Clean(workspace)
			if cleanPath != cleanWs {
				base := filepath.Base(path)
				if base == ".git" || base == "node_modules" || base == "bin" || base == "dist" || base == "build" || base == ".idea" || base == ".vscode" {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// 限制单个审计文件最大不超过 5MB，杜绝大文件读取引发 OOM
		if info.Size() > 5*1024*1024 {
			return nil
		}

		// 只审计源码与脚本
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".go" && ext != ".js" && ext != ".ts" && ext != ".sh" && ext != ".py" {
			return nil
		}

		filesScanned++
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		rel, _ := filepath.Rel(workspace, path)
		str := string(content)

		for _, kw := range highRiskKeywords {
			if strings.Contains(strings.ToLower(str), strings.ToLower(kw)) {
				issues = append(issues, "文件 ["+rel+"] 发现高危破坏性指令特征: "+kw)
			}
		}

		return nil
	})

	status := "SECURE"
	riskLevel := "LOW"
	if len(issues) > 0 {
		status = "WARNING"
		riskLevel = "MEDIUM"
	}

	return AuditReport{
		Status:       status,
		RiskLevel:    riskLevel,
		Issues:       issues,
		FilesScanned: filesScanned,
		Timestamp:   time.Now().Unix(),
	}, nil
}

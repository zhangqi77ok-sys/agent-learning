package loop

import (
	"encoding/json"
	"strings"
	"testing"

	"tiancode/internal/llm"
)

func TestNormalizeStrategy(t *testing.T) {
	if NormalizeStrategy("") != StrategyImplement {
		t.Fatal("empty should be implement")
	}
	if NormalizeStrategy("ANALYZE") != StrategyAnalyze {
		t.Fatal("analyze")
	}
	if NormalizeStrategy("tdd") != StrategyTDD {
		t.Fatal("tdd")
	}
}

func TestApplyStrategy_AnalyzeDropsExec(t *testing.T) {
	tools := []llm.ToolDef{
		{Type: "function", Function: llm.ToolFunctionDef{Name: "fs_control"}},
		{Type: "function", Function: llm.ToolFunctionDef{Name: "exec_command"}},
	}
	out, sys := ApplyStrategy("analyze", "不要动配置", tools, "base")
	if len(out) != 1 || out[0].Function.Name != "fs_control" {
		t.Fatalf("tools=%v", out)
	}
	if !strings.Contains(sys, "analyze") || !strings.Contains(sys, "不要动配置") {
		t.Fatalf("sys=%s", sys)
	}
}

func TestShouldVerifyAfterWrite(t *testing.T) {
	if !ShouldVerifyAfterWrite(StrategyTDD) {
		t.Fatal("tdd must verify after write")
	}
	if ShouldVerifyAfterWrite(StrategyAnalyze) || ShouldVerifyAfterWrite(StrategyImplement) {
		t.Fatal("analyze/implement must not auto-run tests")
	}
}

func TestFormatVerifyFollowup(t *testing.T) {
	ok := FormatVerifyFollowup("a.go", "PASS", true)
	if !strings.Contains(ok, "a.go") || !strings.Contains(ok, "通过") {
		t.Fatalf("%s", ok)
	}
	fail := FormatVerifyFollowup("a.go", "FAIL x", false)
	if !strings.Contains(fail, "失败") || !strings.Contains(fail, "FAIL x") {
		t.Fatalf("%s", fail)
	}
}

func TestApplyStrategy_TDDKeepsExecAndPrompt(t *testing.T) {
	tools := []llm.ToolDef{
		{Type: "function", Function: llm.ToolFunctionDef{Name: "fs_control"}},
		{Type: "function", Function: llm.ToolFunctionDef{Name: "exec_command"}},
	}
	out, sys := ApplyStrategy("tdd", "", tools, "base")
	if len(out) != 2 {
		t.Fatalf("tdd should keep tools, got %d", len(out))
	}
	if !strings.Contains(sys, "tdd") || !strings.Contains(sys, "测试") {
		t.Fatalf("sys=%s", sys)
	}
}

func TestDenyByStrategy_BlocksWrite(t *testing.T) {
	deny, _ := DenyByStrategy("analyze", "exec_command", nil)
	if !deny {
		t.Fatal("expected deny exec")
	}
	raw, _ := json.Marshal(map[string]string{"action": "write", "path": "a.go"})
	deny, _ = DenyByStrategy("analyze", "fs_control", raw)
	if !deny {
		t.Fatal("expected deny write")
	}
	raw, _ = json.Marshal(map[string]string{"action": "read", "path": "a.go"})
	deny, _ = DenyByStrategy("analyze", "fs_control", raw)
	if deny {
		t.Fatal("read should pass")
	}
	deny, _ = DenyByStrategy("implement", "exec_command", nil)
	if deny {
		t.Fatal("implement should allow exec")
	}
}

func TestApplyStrategy_MapFirstAndTDDCompletionBlock(t *testing.T) {
	_, sysAnalyze := ApplyStrategy("analyze", "", nil, "base")
	if !strings.Contains(sysAnalyze, "先地图再下钻") || !strings.Contains(sysAnalyze, "禁止盲目") {
		t.Errorf("expected map-first review rule in analyze prompt, got: %s", sysAnalyze)
	}

	_, sysTDD := ApplyStrategy("tdd", "", nil, "base")
	if !strings.Contains(sysTDD, "测试失败则任务状态绝对不是完成") {
		t.Errorf("expected TDD completion blockage in prompt, got: %s", sysTDD)
	}
}

func TestDenyByStrategy_Turn1MapEnforcement(t *testing.T) {
	// turn 1: 读根目录文件 (如 go.mod, README.md, a.go) 允许
	rawRoot, _ := json.Marshal(map[string]string{"action": "read", "path": "go.mod"})
	deny, _ := DenyByStrategy("analyze", "fs_control", rawRoot, 1)
	if deny {
		t.Errorf("turn 1 reading go.mod should be allowed")
	}

	// turn 1: 读深层业务文件 (如 internal/core/loop/llm_path.go) 应被硬闸阻断，要求先看地图
	rawDeep, _ := json.Marshal(map[string]string{"action": "read", "path": "internal/core/loop/llm_path.go"})
	deny, reason := DenyByStrategy("analyze", "fs_control", rawDeep, 1)
	if !deny {
		t.Errorf("turn 1 reading deep internal file should be denied by map-first rule")
	}
	if !strings.Contains(reason, "先地图后下钻") {
		t.Errorf("expected map-first reason, got: %s", reason)
	}

	// turn 2+: 读深层业务文件放行
	denyTurn2, _ := DenyByStrategy("analyze", "fs_control", rawDeep, 2)
	if denyTurn2 {
		t.Errorf("turn 2+ reading deep file should be allowed for drill-down")
	}
}

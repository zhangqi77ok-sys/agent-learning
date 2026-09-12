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

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

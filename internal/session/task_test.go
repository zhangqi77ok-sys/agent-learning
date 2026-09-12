package session

import (
	"os"
	"testing"
)

func TestTaskModel_PersistenceAndStatus(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tcode_test_task_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	s := &Store{baseDir: tempDir}
	sess := ChatSession{
		ID:        "sess_task_1",
		Title:     "重构登录模块",
		Model:     "gpt-5.6-sol",
		CreatedAt: 1788480000,
		UpdatedAt: 1788480000,
		Messages:  []SessionMessage{{ID: "m1", Role: "user", Content: "重构登录模块"}},
		Task: &TaskModel{
			Goal:             "重构登录模块并增加测试",
			Status:           TaskStatusRunning,
			ToolBudget:       24,
			ToolsUsed:        5,
			Summary:          "已完成接口定义，正在实现具体逻辑",
			PendingDiffFiles: []string{"internal/auth/login.go"},
		},
	}

	if err := s.Save(sess); err != nil {
		t.Fatalf("save session with task failed: %v", err)
	}

	loaded, err := s.Get("sess_task_1")
	if err != nil {
		t.Fatalf("get session failed: %v", err)
	}
	if loaded.Task == nil {
		t.Fatalf("expected non-nil Task on loaded session")
	}
	if loaded.Task.Goal != "重构登录模块并增加测试" {
		t.Errorf("expected goal '重构登录模块并增加测试', got '%s'", loaded.Task.Goal)
	}
	if loaded.Task.Status != TaskStatusRunning {
		t.Errorf("expected status 'running', got '%s'", loaded.Task.Status)
	}
	if len(loaded.Task.PendingDiffFiles) != 1 || loaded.Task.PendingDiffFiles[0] != "internal/auth/login.go" {
		t.Errorf("expected pending diff file 'internal/auth/login.go', got %v", loaded.Task.PendingDiffFiles)
	}
}

func TestCleanGoalPrompt(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"[执行策略 implement: 功能实现]\n[附加约束 不要改已有测试]\n\n重构登录模块", "重构登录模块"},
		{"[执行策略 analyze: 审查与分析]\n审查 internal/core", "审查 internal/core"},
		{"普通用户需求", "普通用户需求"},
		{"[单标签] 内容", "内容"},
	}
	for _, c := range cases {
		got := CleanGoalPrompt(c.input)
		if got != c.expected {
			t.Errorf("CleanGoalPrompt(%q) = %q, want %q", c.input, got, c.expected)
		}
	}
}

func TestIsContinuationPrompt(t *testing.T) {
	cases := []struct {
		input    string
		expected bool
	}{
		{"继续", true},
		{"继续执行", true},
		{"continue", true},
		{"Continue", true},
		{" go on ", true},
		{"接着做", true},
		{"接着把审查写完", true},
		{"请继续优化 internal/core", true},
		{"继续完善测试用例", true},
		{"[执行策略 implement: 功能实现]\n\n接着把未完成项写完", true},
		{"[执行策略 analyze: 审查]\n继续", true},
		{"请帮我写一个快速排序", false},
		{"重新开始", false},
	}

	for _, c := range cases {
		got := IsContinuationPrompt(c.input)
		if got != c.expected {
			t.Errorf("IsContinuationPrompt(%q) = %v, want %v", c.input, got, c.expected)
		}
	}
}

func TestBuildContinuationContext(t *testing.T) {
	task := &TaskModel{
		Goal:      "审查前端性能瓶颈",
		Status:    TaskStatusCapped,
		Summary:   "已分析完 App.vue，还需分析 LeftDrawer.vue 与 DiffWorkspace.vue",
		ToolsUsed: 15,
	}

	ctxPrompt := BuildContinuationContext(task, "接着把审查写完")
	if ctxPrompt == "" {
		t.Fatalf("expected non-empty continuation context")
	}
	if !containsStr(ctxPrompt, "审查前端性能瓶颈") {
		t.Errorf("expected goal in prompt, got: %s", ctxPrompt)
	}
	if !containsStr(ctxPrompt, "LeftDrawer.vue") {
		t.Errorf("expected summary in prompt, got: %s", ctxPrompt)
	}
	if !containsStr(ctxPrompt, "接着把审查写完") {
		t.Errorf("expected follow up in prompt, got: %s", ctxPrompt)
	}
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

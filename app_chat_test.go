package main

import (
	"strings"
	"testing"

	"tiancode/internal/config"
)

func TestResolveChatCredentials_FailClosed(t *testing.T) {
	_, _, _, err := resolveChatCredentials(nil, "hi")
	if err == nil || !strings.Contains(err.Error(), "未配置") {
		t.Fatalf("expected missing channel, got %v", err)
	}
	_, _, _, err = resolveChatCredentials(&config.ChannelConfig{Endpoint: "https://x", APIKey: "k"}, "")
	if err == nil || !strings.Contains(err.Error(), "未指定模型") {
		t.Fatalf("expected missing model, got %v", err)
	}
}

func TestResolveChatCredentials_UsesPrimary(t *testing.T) {
	primary := &config.ChannelConfig{
		Endpoint: "https://example.invalid/v1",
		APIKey:   "fake-api-key-0123456789abcdef",
		Model:    "local-model",
	}
	ep, key, model, err := resolveChatCredentials(primary, "")
	if err != nil {
		t.Fatal(err)
	}
	if ep != "https://example.invalid/v1" || key == "" || model != "local-model" {
		t.Fatalf("got %s %s %s", ep, key, model)
	}
	_, _, model, err = resolveChatCredentials(primary, "override")
	if err != nil || model != "override" {
		t.Fatalf("override model: %s %v", model, err)
	}
}

func TestAppendEnabledPolicies_SkipsDisabledAndEmpty(t *testing.T) {
	out := appendEnabledPolicies("base", []config.SkillConfig{
		{Name: "off", Prompt: "x", Enabled: false},
		{Name: "go-style", Prompt: "prefer go test", Enabled: true},
		{Name: "empty", Prompt: "  ", Enabled: true},
	}, []config.RuleConfig{
		{Title: "r1", Content: "no fake data", Enabled: true},
	})
	if !strings.Contains(out, "[技能 go-style] prefer go test") {
		t.Fatalf("missing skill: %s", out)
	}
	if strings.Contains(out, "off") || strings.Contains(out, "[技能 empty]") {
		t.Fatalf("disabled/empty skill leaked: %s", out)
	}
	if !strings.Contains(out, "[规则规约] no fake data") {
		t.Fatalf("missing rule: %s", out)
	}
}

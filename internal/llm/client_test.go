package llm

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMessage_ReasoningContentPassedBack(t *testing.T) {
	msg := Message{
		Role:               "assistant",
		Content:            "ok",
		ReasoningContent:   "先看目录再读文件",
		ToolCalls:          []ToolCall{{ID: "c1", Type: "function"}},
	}
	msg.ToolCalls[0].Function.Name = "fs_control"
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, `"reasoning_content":"先看目录再读文件"`) {
		t.Fatalf("thinking must be serialized for AgentRouter: %s", s)
	}
	empty := Message{Role: "assistant", Content: "hi"}
	data, _ = json.Marshal(empty)
	if strings.Contains(string(data), "reasoning_content") {
		t.Fatalf("empty thinking should omit field: %s", string(data))
	}
}
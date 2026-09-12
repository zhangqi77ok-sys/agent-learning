package loop

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"tiancode/internal/llm"
)

func (e *ExecutionEngine) executeDirectLLM(ctx context.Context, req *EngineRequest, eventChan chan<- EngineEvent) error {
	conversation := req.Messages
	if len(conversation) == 0 {
		sys := req.SystemPrompt
		if sys == "" {
			sys = "你是 湉码 / tiancode 纯原生桌面智能体。你有权调用工具来审查、读取、修改工程代码及运行测试命令。"
		}
		conversation = []llm.Message{
			{Role: "system", Content: sys},
			{Role: "user", Content: req.Prompt},
		}
	}
	tools := req.LLMTools

	const maxWatchdogTurns = 12
	for turn := 1; turn <= maxWatchdogTurns; turn++ {
		if ctx.Err() != nil {
			eventChan <- EngineEvent{Type: EventError, ErrorMessage: "task canceled by client"}
			return ctx.Err()
		}

		llmReq := llm.Request{
			Endpoint: req.Endpoint,
			APIKey:   req.APIKey,
			Model:    req.Model,
			Messages: conversation,
			Tools:    tools,
		}

		var roundContent strings.Builder
		toolCalls, err := llm.StreamChat(ctx, llmReq, llm.StreamHandlers{
			OnThinking: func(text string) {
				eventChan <- EngineEvent{Type: EventChunk, Thinking: text}
			},
			OnContent: func(delta string) {
				roundContent.WriteString(delta)
				eventChan <- EngineEvent{Type: EventChunk, DeltaContent: delta}
			},
			OnError: func(streamErr error) {
				eventChan <- EngineEvent{Type: EventError, ErrorMessage: streamErr.Error()}
			},
		})
		if err != nil || len(toolCalls) == 0 {
			break
		}

		for i := range toolCalls {
			if toolCalls[i].ID == "" {
				toolCalls[i].ID = fmt.Sprintf("call_%d_%d", i, time.Now().UnixNano())
			}
			if toolCalls[i].Type == "" {
				toolCalls[i].Type = "function"
			}
		}

		conversation = append(conversation, llm.Message{
			Role:      "assistant",
			Content:   roundContent.String(),
			ToolCalls: toolCalls,
		})

		for _, tc := range toolCalls {
			toolName := tc.Function.Name
			toolArgs := tc.Function.Arguments
			rawToolArgs := json.RawMessage(toolArgs)

			eventChan <- EngineEvent{
				Type:       EventToolStart,
				ToolCallID: tc.ID,
				ToolName:   toolName,
				ToolArgs:   rawToolArgs,
			}

			output, isErr, written := e.runTool(ctx, req.SessionID, toolName, rawToolArgs, nil)

			eventChan <- EngineEvent{
				Type:       EventToolEnd,
				ToolCallID: tc.ID,
				ToolName:   toolName,
				ToolOutput: output,
				IsError:    isErr,
			}
			if written != "" {
				eventChan <- EngineEvent{Type: EventFilesChanged, ToolName: written}
			}

			toolOutput := strings.TrimSpace(output)
			if toolOutput == "" {
				toolOutput = fmt.Sprintf("tool [%s] executed successfully with empty output", toolName)
			}
			conversation = append(conversation, llm.Message{
				Role:       "tool",
				ToolCallID: tc.ID,
				Name:       toolName,
				Content:    toolOutput,
			})
		}
	}

	eventChan <- EngineEvent{Type: EventDone}
	return nil
}

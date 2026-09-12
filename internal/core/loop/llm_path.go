package loop

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"tiancode/internal/llm"
	v1 "tiancode/pkg/plugin/v1"
)

func (e *ExecutionEngine) executeDirectLLM(ctx context.Context, req *EngineRequest, eventChan chan<- EngineEvent) error {
	provs := e.registry.GetProviders()
	if len(provs) == 0 {
		eventChan <- EngineEvent{Type: EventError, ErrorMessage: "no provider plugin registered"}
		return fmt.Errorf("no provider registered")
	}
	prov := provs[0]
	cfg, _ := json.Marshal(map[string]string{
		"api_key":  req.APIKey,
		"base_url": strings.TrimRight(req.Endpoint, "/"),
	})
	if err := prov.Init(ctx, cfg); err != nil {
		eventChan <- EngineEvent{Type: EventError, ErrorMessage: err.Error()}
		return err
	}

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

	toolDefs := make([]v1.ToolDefinition, 0, len(req.LLMTools))
	for _, t := range req.LLMTools {
		params, _ := json.Marshal(t.Function.Parameters)
		toolDefs = append(toolDefs, v1.ToolDefinition{
			Name:        t.Function.Name,
			Description: t.Function.Description,
			Parameters:  params,
		})
	}
	if len(toolDefs) == 0 {
		for _, t := range e.registry.GetTools() {
			toolDefs = append(toolDefs, t.Definition())
		}
	}

	maxTurns := e.maxLLMTurns
	if maxTurns < 1 {
		maxTurns = 24
	}
	hitCap := false
	for turn := 1; turn <= maxTurns; turn++ {
		if ctx.Err() != nil {
			eventChan <- EngineEvent{
				Type:         EventChunk,
				DeltaContent: FormatInterruptedNotice(),
			}
			eventChan <- EngineEvent{Type: EventError, ErrorMessage: "task canceled by client"}
			return ctx.Err()
		}

		msgsBytes, err := json.Marshal(conversation)
		if err != nil {
			humanErr := FormatUpstreamError(err)
			eventChan <- EngineEvent{Type: EventChunk, DeltaContent: humanErr}
			eventChan <- EngineEvent{Type: EventError, ErrorMessage: humanErr}
			return err
		}
		chatReq := &v1.ChatRequest{
			Model:    req.Model,
			Messages: msgsBytes,
			Tools:    toolDefs,
			Stream:   true,
		}
		chunkChan, err := prov.StreamChat(ctx, chatReq)
		if err != nil {
			humanErr := FormatUpstreamError(err)
			eventChan <- EngineEvent{Type: EventChunk, DeltaContent: humanErr}
			eventChan <- EngineEvent{Type: EventError, ErrorMessage: humanErr}
			return err
		}

		var asstContent strings.Builder
		var asstThinking strings.Builder
		toolReassembler := make(map[int]*AssembledToolCall)
		for chunk := range chunkChan {
			if chunk.Error != nil {
				humanErr := FormatUpstreamError(chunk.Error)
				eventChan <- EngineEvent{Type: EventChunk, DeltaContent: humanErr}
				eventChan <- EngineEvent{Type: EventError, ErrorMessage: humanErr}
				return chunk.Error
			}
			if chunk.DeltaContent != "" || chunk.Thinking != "" {
				asstContent.WriteString(chunk.DeltaContent)
				asstThinking.WriteString(chunk.Thinking)
				eventChan <- EngineEvent{
					Type:         EventChunk,
					DeltaContent: chunk.DeltaContent,
					Thinking:     chunk.Thinking,
				}
			}
			for _, tc := range chunk.ToolCalls {
				entry, exists := toolReassembler[tc.Index]
				if !exists {
					entry = &AssembledToolCall{ID: tc.ID, Name: tc.Name}
					toolReassembler[tc.Index] = entry
				}
				if tc.ID != "" {
					entry.ID = tc.ID
				}
				if tc.Name != "" {
					entry.Name = tc.Name
				}
				entry.Arguments.WriteString(tc.ArgumentsDelta)
			}
		}

		if len(toolReassembler) == 0 {
			if turn == 1 && asstContent.Len() == 0 && asstThinking.Len() == 0 {
				eventChan <- EngineEvent{
					Type:         EventChunk,
					DeltaContent: FormatEmptyOutputNotice(),
				}
			}
			break
		}
		if turn == maxTurns {
			hitCap = true
		}

		tcIndices := make([]int, 0, len(toolReassembler))
		for idx := range toolReassembler {
			tcIndices = append(tcIndices, idx)
		}
		sort.Ints(tcIndices)
		rawToolCalls := make([]llm.ToolCall, 0, len(tcIndices))
		for _, idx := range tcIndices {
			atc := toolReassembler[idx]
			if atc == nil {
				continue
			}
			if atc.ID == "" {
				atc.ID = fmt.Sprintf("call_%d_%d", turn, idx)
			}
			tc := llm.ToolCall{ID: atc.ID, Type: "function"}
			tc.Function.Name = atc.Name
			tc.Function.Arguments = atc.Arguments.String()
			rawToolCalls = append(rawToolCalls, tc)
		}
		conversation = append(conversation, llm.Message{
			Role:             "assistant",
			Content:          asstContent.String(),
			ReasoningContent: asstThinking.String(),
			ToolCalls:        rawToolCalls,
		})

		for _, tc := range rawToolCalls {
			rawArgs := json.RawMessage(tc.Function.Arguments)
			eventChan <- EngineEvent{
				Type:       EventToolStart,
				ToolCallID: tc.ID,
				ToolName:   tc.Function.Name,
				ToolArgs:   rawArgs,
			}
			output, isErr, written, tddPass := e.runTool(ctx, req.SessionID, tc.Function.Name, rawArgs, nil, req.Strategy, turn)
			eventChan <- EngineEvent{
				Type:       EventToolEnd,
				ToolCallID: tc.ID,
				ToolName:   tc.Function.Name,
				ToolOutput: output,
				IsError:    isErr,
			}
			if tddPass != nil {
				eventChan <- EngineEvent{
					Type:      EventTDDResult,
					TDDPassed: tddPass,
				}
			}
			if written != "" {
				eventChan <- EngineEvent{Type: EventFilesChanged, ToolName: written}
			}
			toolOutput := strings.TrimSpace(output)
			if toolOutput == "" {
				toolOutput = fmt.Sprintf("tool [%s] executed successfully with empty output", tc.Function.Name)
			}
			conversation = append(conversation, llm.Message{
				Role:       "tool",
				ToolCallID: tc.ID,
				Name:       tc.Function.Name,
				Content:    toolOutput,
			})
		}
	}

	if hitCap {
		eventChan <- EngineEvent{Type: EventHitCap}
		notice := FormatHitCapNotice(maxTurns)
		eventChan <- EngineEvent{Type: EventChunk, DeltaContent: notice}
		conversation = append(conversation, llm.Message{
			Role:    "user",
			Content: "工具轮次已达上限。请不要再调用任何工具，用已经拿到的结果给出当前结论、未完成项和下一步建议。",
		})
		msgsBytes, err := json.Marshal(conversation)
		if err == nil && ctx.Err() == nil {
			chatReq := &v1.ChatRequest{Model: req.Model, Messages: msgsBytes, Stream: true}
			chunkChan, err := prov.StreamChat(ctx, chatReq)
			if err == nil {
				for chunk := range chunkChan {
					if chunk.Error != nil {
						eventChan <- EngineEvent{Type: EventError, ErrorMessage: chunk.Error.Error()}
						break
					}
					if chunk.DeltaContent != "" || chunk.Thinking != "" {
						eventChan <- EngineEvent{
							Type:         EventChunk,
							DeltaContent: chunk.DeltaContent,
							Thinking:     chunk.Thinking,
						}
					}
				}
			}
		}
	}

	eventChan <- EngineEvent{Type: EventDone}
	return nil
}

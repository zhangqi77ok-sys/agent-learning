package loop

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"tiancode/internal/host"
	"tiancode/internal/llm"
	v1 "tiancode/pkg/plugin/v1"
)

// EventType 引擎向前端派发的事件类型
type EventType string

const (
	EventChunk        EventType = "chunk"
	EventToolStart    EventType = "tool_start"
	EventToolEnd      EventType = "tool_end"
	EventDone         EventType = "done"
	EventError        EventType = "error"
	EventFilesChanged EventType = "files_changed"
)

// EngineEvent 引擎事件
type EngineEvent struct {
	Type         EventType       `json:"type"`
	DeltaContent string          `json:"delta_content,omitempty"`
	Thinking     string          `json:"thinking,omitempty"`
	ToolCallID   string          `json:"tool_call_id,omitempty"`
	ToolName     string          `json:"tool_name,omitempty"`
	ToolArgs     json.RawMessage `json:"tool_args,omitempty"`
	ToolOutput   string          `json:"tool_output,omitempty"`
	IsError      bool            `json:"is_error,omitempty"`
	ErrorMessage string          `json:"error_message,omitempty"`
}

// EngineRequest 用户推理请求
type EngineRequest struct {
	Model        string `json:"model"`
	Prompt       string `json:"prompt"`
	Provider     string `json:"provider,omitempty"`
	SessionID    string
	Endpoint     string
	APIKey       string
	SystemPrompt string
	Messages     []llm.Message
	LLMTools     []llm.ToolDef
	Strategy     string
	StrategyNote string
}

// AssembledToolCall 组装后的工具调用
type AssembledToolCall struct {
	ID        string
	Name      string
	Arguments strings.Builder
}

// ExecutionEngine ReAct 双环自主执行引擎
type ExecutionEngine struct {
	registry *host.Registry
	maxSteps int
	MCPCall  func(ctx context.Context, name string, args map[string]any) (string, error)
	// Verify 在 TDD 策略写盘成功后运行工作区测试，结果会拼进工具输出。
	Verify func(writtenFile string) (output string, pass bool)
}

// NewExecutionEngine 构造执行引擎
func NewExecutionEngine(reg *host.Registry) *ExecutionEngine {
	return &ExecutionEngine{
		registry: reg,
		maxSteps: 15, // 生产硬防死循环上限
	}
}

// Execute 驱动完整的 ReAct 自主思考与工具调用闭环
func (e *ExecutionEngine) Execute(ctx context.Context, req *EngineRequest, eventChan chan<- EngineEvent) error {
	defer close(eventChan)

	if e == nil || e.registry == nil {
		eventChan <- EngineEvent{Type: EventError, ErrorMessage: "engine or registry is not initialized"}
		return fmt.Errorf("engine or registry is not initialized")
	}
	if req == nil {
		eventChan <- EngineEvent{Type: EventError, ErrorMessage: "engine request cannot be nil"}
		return fmt.Errorf("engine request cannot be nil")
	}

	if req.APIKey != "" && req.Endpoint != "" {
		return e.executeDirectLLM(ctx, req, eventChan)
	}

	// 1. 查找对应的 Provider
	provList := e.registry.GetProviders()
	if len(provList) == 0 {
		eventChan <- EngineEvent{Type: EventError, ErrorMessage: "no provider plugin registered in micro-kernel"}
		return fmt.Errorf("no provider registered")
	}
	prov := provList[0]

	// 2. 收集已注册的 ToolPlugin 声明
	tools := e.registry.GetTools()
	toolDefs := make([]v1.ToolDefinition, 0, len(tools))
	toolMap := make(map[string]v1.ToolPlugin)
	for _, t := range tools {
		def := t.Definition()
		toolDefs = append(toolDefs, def)
		toolMap[def.Name] = t
	}

	// 3. 构建多轮上下文队列
	systemPrompt := req.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = "你是 湉码 / tiancode 纯原生桌面智能体。你有权调用工具来审查、读取、修改工程代码及运行测试命令。"
	}
	messages := []map[string]any{
		{"role": "system", "content": systemPrompt},
		{"role": "user", "content": req.Prompt},
	}

	step := 0
	for step < e.maxSteps {
		step++
		select {
		case <-ctx.Done():
			eventChan <- EngineEvent{Type: EventError, ErrorMessage: "task canceled by client"}
			return ctx.Err()
		default:
		}

		msgsBytes, _ := json.Marshal(messages)
		chatReq := &v1.ChatRequest{
			Model:    req.Model,
			Messages: msgsBytes,
			Tools:    toolDefs,
			Stream:   true,
		}

		chunkChan, err := prov.StreamChat(ctx, chatReq)
		if err != nil {
			eventChan <- EngineEvent{Type: EventError, ErrorMessage: err.Error()}
			return err
		}

		// 收集本轮模型生成的增量文本、思考与工具调用分片
		var asstContent strings.Builder
		var asstThinking strings.Builder
		toolReassembler := make(map[int]*AssembledToolCall)

		for chunk := range chunkChan {
			if chunk.Error != nil {
				eventChan <- EngineEvent{Type: EventError, ErrorMessage: chunk.Error.Error()}
				return chunk.Error
			}

			// 派发流式文本与思考链
			if chunk.DeltaContent != "" || chunk.Thinking != "" {
				asstContent.WriteString(chunk.DeltaContent)
				asstThinking.WriteString(chunk.Thinking)
				eventChan <- EngineEvent{
					Type:         EventChunk,
					DeltaContent: chunk.DeltaContent,
					Thinking:     chunk.Thinking,
				}
			}

			// 收集组装工具调用
			for _, tc := range chunk.ToolCalls {
				entry, exists := toolReassembler[tc.Index]
				if !exists {
					entry = &AssembledToolCall{
						ID:   tc.ID,
						Name: tc.Name,
					}
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

		// 若本轮无工具调用，说明已得出最终解答，平滑退出循环
		if len(toolReassembler) == 0 {
			break
		}

		// 模型发起了工具调用：记录 Assistant 历史消息
		assistantMsg := map[string]any{
			"role":    "assistant",
			"content": asstContent.String(),
		}
		if think := asstThinking.String(); think != "" {
			assistantMsg["reasoning_content"] = think
		}
		// 收集排序后的所有 tool call 索引，兼容非 0 开始与稀疏索引
		tcIndices := make([]int, 0, len(toolReassembler))
		for idx := range toolReassembler {
			tcIndices = append(tcIndices, idx)
		}
		sort.Ints(tcIndices)

		rawToolCalls := make([]map[string]any, 0, len(tcIndices))
		for _, idx := range tcIndices {
			atc := toolReassembler[idx]
			if atc == nil {
				continue
			}
			if atc.ID == "" {
				atc.ID = fmt.Sprintf("call_%d_%d", step, idx)
			}
			rawToolCalls = append(rawToolCalls, map[string]any{
				"id":   atc.ID,
				"type": "function",
				"function": map[string]any{
					"name":      atc.Name,
					"arguments": atc.Arguments.String(),
				},
			})
		}
		assistantMsg["tool_calls"] = rawToolCalls
		messages = append(messages, assistantMsg)

		// 依次安全调度物理算子
		for _, idx := range tcIndices {
			atc := toolReassembler[idx]
			if atc == nil {
				continue
			}

			argsStr := atc.Arguments.String()
			rawArgs := json.RawMessage(argsStr)

			// 通知前端工具启动
			eventChan <- EngineEvent{
				Type:       EventToolStart,
				ToolCallID: atc.ID,
				ToolName:   atc.Name,
				ToolArgs:   rawArgs,
			}

			toolOutput, isErr, written := e.runTool(ctx, req.SessionID, atc.Name, rawArgs, toolMap, req.Strategy)

			// 通知前端工具完成
			eventChan <- EngineEvent{
				Type:       EventToolEnd,
				ToolCallID: atc.ID,
				ToolName:   atc.Name,
				ToolOutput: toolOutput,
				IsError:    isErr,
			}
			if written != "" {
				eventChan <- EngineEvent{Type: EventFilesChanged, ToolName: written}
			}

			// 将工具结果回填进会话上下文，确保 content 绝不为空（防御上游网关 400 校验）
			contentForContext := toolOutput
			if strings.TrimSpace(contentForContext) == "" {
				contentForContext = fmt.Sprintf("tool [%s] executed successfully with empty output", atc.Name)
			}
			messages = append(messages, map[string]any{
				"role":         "tool",
				"tool_call_id": atc.ID,
				"content":      contentForContext,
			})
		}
	}

	if step >= e.maxSteps {
		eventChan <- EngineEvent{
			Type:         EventChunk,
			DeltaContent: "\n\n⚠️ 【系统提示】已达到智能体自主推理最大步数上限（15 步），执行已安全收敛终止。如需继续，请补充新指令。",
		}
	}

	eventChan <- EngineEvent{Type: EventDone}
	return nil
}

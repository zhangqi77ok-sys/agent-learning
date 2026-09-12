package main

import (
	"tiancode/internal/lsp"
	"tiancode/internal/telemetry"

	"context"
	"fmt"
	"strings"
	"time"

	"tiancode/internal/config"
	"tiancode/internal/core/loop"
	"tiancode/internal/core/sandbox"
	"tiancode/internal/session"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type ChatRequest struct {
	SessionID  string `json:"session_id"`
	Prompt     string `json:"prompt"`
	Model      string `json:"model"`
	IsFullAuto bool   `json:"is_full_auto"`
}

func resolveChatCredentials(primary *config.ChannelConfig, reqModel string) (endpoint, apiKey, model string, err error) {
	if primary == nil {
		return "", "", "", fmt.Errorf("未配置任何模型渠道。请在设置中添加真实 endpoint 与 API Key，禁止使用内置假地址")
	}
	endpoint = strings.TrimSpace(primary.Endpoint)
	apiKey = strings.TrimSpace(primary.APIKey)
	model = strings.TrimSpace(reqModel)
	if model == "" {
		model = strings.TrimSpace(primary.Model)
	}
	if endpoint == "" {
		return "", "", "", fmt.Errorf("主渠道未填写 endpoint")
	}
	if apiKey == "" {
		return "", "", "", fmt.Errorf("主渠道未填写 API Key")
	}
	if model == "" {
		return "", "", "", fmt.Errorf("未指定模型：请在对话顶栏选择，或在渠道中填写 model")
	}
	return endpoint, apiKey, model, nil
}

func (a *App) SendMessage(req ChatRequest) error {
	if a.ctx == nil {
		return fmt.Errorf("context not initialized")
	}

	a.agentMu.Lock()
	if a.agentCancel != nil {
		a.agentCancel()
	}
	agentCtx, cancel := context.WithCancel(context.Background())
	a.agentTaskID++
	currentTaskID := a.agentTaskID
	a.agentCancel = cancel
	a.currentSessionID = req.SessionID
	a.agentMu.Unlock()

	go func() {
		defer func() {
			a.agentMu.Lock()
			if a.agentTaskID == currentTaskID {
				a.agentCancel = nil
			}
			a.agentMu.Unlock()
		}()

		var primary *config.ChannelConfig
		if a.channelStore != nil {
			primary = a.channelStore.GetPrimary()
		}
		endpoint, apiKey, model, credErr := resolveChatCredentials(primary, req.Model)
		if credErr != nil {
			errMsg := "\n\n[配置错误] " + credErr.Error()
			runtime.EventsEmit(a.ctx, "agent:start", map[string]any{
				"session_id": req.SessionID,
				"model":      model,
			})
			runtime.EventsEmit(a.ctx, "agent:chunk", map[string]any{
				"session_id": req.SessionID,
				"delta":      errMsg,
				"turn":       1,
			})
			runtime.EventsEmit(a.ctx, "agent:complete", map[string]any{
				"session_id": req.SessionID,
				"error":      "missing_api_key",
			})
			return
		}

		// 2. 加载已有会话历史，若不存在则新建
		var currentSession session.ChatSession
		existing, err := a.sessionStore.Get(req.SessionID)
		if err == nil && existing != nil {
			currentSession = *existing
		} else {
			currentSession = session.ChatSession{
				ID:        req.SessionID,
				Title:     req.Prompt,
				Model:     model,
				Tag:       "默认",
				CreatedAt: time.Now().Unix(),
				UpdatedAt: time.Now().Unix(),
				Messages:  make([]session.SessionMessage, 0),
			}
			r := []rune(req.Prompt)
			if len(r) > 16 {
				currentSession.Title = string(r[:16]) + "..."
			}
		}

		// 追加用户消息
		userMsg := session.SessionMessage{
			ID:      fmt.Sprintf("msg_%d", time.Now().UnixNano()),
			Role:    "user",
			Content: req.Prompt,
			Time:    time.Now().Format("15:04"),
		}
		currentSession.Messages = append(currentSession.Messages, userMsg)
		_ = a.sessionStore.Save(currentSession)

		// 3. 发送开始事件
		runtime.EventsEmit(a.ctx, "agent:start", map[string]any{
			"session_id": req.SessionID,
			"model":      model,
		})

		// 4. 构建提示词体系 (注入规则 + 工作区技术栈感知 + 最近多轮历史)
		systemPrompt := "你是 湉码 / tiancode 纯原生桌面智能体。你有权调用工具来审查、读取、修改工程代码及运行测试命令。请优先利用工具解决问题，并在每次调用后解释原因。"
		rules := a.extraStore.ListRules()
		for _, r := range rules {
			if r.Enabled {
				systemPrompt += "\n[规则规约] " + r.Content
			}
		}
		// 动态侦测工作区项目技术栈并注入环境上下文
		stackInfo := sandbox.DetectProjectStack(a.workspace)
		if stackPrompt := sandbox.FormatStackPrompt(stackInfo); stackPrompt != "" {
			systemPrompt += "\n" + stackPrompt
		}

		// 动态上下文窗口：基于预算自适应选择多轮历史，避免截断关键上下文或超出 Token 上限
		conversation := buildConversationWindow(systemPrompt, currentSession.Messages, 32000)

		var assistantThinking strings.Builder
		var assistantContent strings.Builder
		var lastToolExec *session.ToolExecution
		allToolExecs := make([]session.ToolExecution, 0)

		workspaceTools := a.buildLLMToolsFromRegistry(agentCtx)
		roundStart := time.Now()
		eventChan := make(chan loop.EngineEvent, 64)
		go func() {
			_ = a.engine.Execute(agentCtx, &loop.EngineRequest{
				Model:        model,
				Prompt:       req.Prompt,
				SessionID:    req.SessionID,
				Endpoint:     endpoint,
				APIKey:       apiKey,
				SystemPrompt: systemPrompt,
				Messages:     conversation,
				LLMTools:     workspaceTools,
			}, eventChan)
		}()

		for ev := range eventChan {
			switch ev.Type {
			case loop.EventChunk:
				if ev.Thinking != "" {
					assistantThinking.WriteString(ev.Thinking)
					runtime.EventsEmit(a.ctx, "agent:thinking", map[string]any{
						"session_id": req.SessionID,
						"thinking":   ev.Thinking,
					})
				}
				if ev.DeltaContent != "" {
					assistantContent.WriteString(ev.DeltaContent)
					runtime.EventsEmit(a.ctx, "agent:chunk", map[string]any{
						"session_id": req.SessionID,
						"delta":      ev.DeltaContent,
					})
				}
			case loop.EventToolStart:
				runtime.EventsEmit(a.ctx, "agent:tool_start", map[string]any{
					"session_id": req.SessionID,
					"id":         ev.ToolCallID,
					"tool":       ev.ToolName,
					"args":       string(ev.ToolArgs),
				})
			case loop.EventToolEnd:
				tExec := session.ToolExecution{Name: ev.ToolName, Args: string(ev.ToolArgs), Output: ev.ToolOutput}
				allToolExecs = append(allToolExecs, tExec)
				cp := tExec
				lastToolExec = &cp
				runtime.EventsEmit(a.ctx, "agent:tool_end", map[string]any{
					"session_id": req.SessionID,
					"id":         ev.ToolCallID,
					"tool":       ev.ToolName,
					"output":     ev.ToolOutput,
				})
			case loop.EventFilesChanged:
				runtime.EventsEmit(a.ctx, "agent:files_changed", map[string]any{
					"session_id": req.SessionID,
					"file":       ev.ToolName,
				})
				if diagReport, diagErr := lsp.DiagnoseFile(a.workspace, ev.ToolName); diagErr == nil && diagReport != nil && diagReport.HasErrors {
					runtime.EventsEmit(a.ctx, "lsp:diagnostic", map[string]any{
						"session_id": req.SessionID,
						"file":       ev.ToolName,
						"has_errors": true,
						"errors":     diagReport.Errors,
					})
				}
			case loop.EventError:
				errMsg := fmt.Sprintf("\n\n[系统错误: %s]", ev.ErrorMessage)
				assistantContent.WriteString(errMsg)
				runtime.EventsEmit(a.ctx, "agent:chunk", map[string]any{
					"session_id": req.SessionID,
					"delta":      errMsg,
				})
			}
		}

		roundDuration := time.Since(roundStart).Milliseconds()
		promptTok := (len(req.Prompt) + 300) / 3
		compTok := (assistantContent.Len() + assistantThinking.Len()) / 3
		if compTok < 1 && assistantContent.Len() > 0 {
			compTok = 1
		}
		telemetry.GetTracker().Record(model, promptTok, compTok, roundDuration)

		// 7. 持久化 Assistant 回复至磁盘
		if agentCtx.Err() != nil {
			// 若已被用户中断且没有任何有效产出，跳过写入空 assistant 消息，避免污染历史
			if assistantContent.Len() == 0 && assistantThinking.Len() == 0 && len(allToolExecs) == 0 {
				return
			}
		}

		asstMsg := session.SessionMessage{
			ID:       fmt.Sprintf("msg_%d", time.Now().UnixNano()),
			Role:     "assistant",
			Content:  assistantContent.String(),
			Thinking: assistantThinking.String(),
			Tool:     lastToolExec,
			Tools:    allToolExecs,
			Time:     time.Now().Format("15:04"),
		}
		currentSession.Messages = append(currentSession.Messages, asstMsg)
		currentSession.UpdatedAt = time.Now().Unix()
		_ = a.sessionStore.Save(currentSession)

		if agentCtx.Err() == nil {
			runtime.EventsEmit(a.ctx, "agent:done", map[string]any{"session_id": req.SessionID})
		}
	}()

	return nil
}

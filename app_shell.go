package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tiancode/internal/ast"
	terminaltool "tiancode/plugins/tool/terminal"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) GetProjectASTGraph() ([]ast.GraphNode, error) {
	return ast.ScanWorkspaceAST(a.workspace)
}

func (a *App) GetGitStatus() (map[string]any, error) {
	gitTool, ok := a.registry.GetTool("tool.git")
	if !ok {
		return map[string]any{"error": "git tool not registered"}, nil
	}
	rawArgs, _ := json.Marshal(map[string]any{})
	res, err := gitTool.Execute(a.ctx, rawArgs)
	if err != nil {
		return map[string]any{"branch": "", "staged": []any{}, "working": []any{}, "untracked": []any{}, "error": err.Error()}, nil
	}
	if res != nil && res.IsError {
		return map[string]any{"branch": "", "staged": []any{}, "working": []any{}, "untracked": []any{}, "error": res.Content}, nil
	}
	// git_status tool 返回 JSON 格式的状态报告，直接解析
	var report map[string]any
	if jsonErr := json.Unmarshal([]byte(res.Content), &report); jsonErr != nil {
		return map[string]any{"raw": res.Content}, nil
	}
	return report, nil
}

func (a *App) GetFileTree(dir string) ([]FileNode, error) {
	targetDir := a.workspace
	if dir != "" {
		if a.sandbox != nil {
			validated, err := a.sandbox.ValidatePath(dir)
			if err != nil {
				return nil, fmt.Errorf("invalid path access: %w", err)
			}
			targetDir = validated
		} else {
			targetDir = filepath.Join(a.workspace, dir)
		}
	}
	return a.buildFileTree(targetDir, 0, 4)
}

func (a *App) buildFileTree(currentDir string, currentDepth, maxDepth int) ([]FileNode, error) {
	nodeCount := 0
	visited := make(map[string]bool)
	return a.buildFileTreeInternal(currentDir, currentDepth, maxDepth, &nodeCount, visited)
}

func (a *App) buildFileTreeInternal(currentDir string, currentDepth, maxDepth int, nodeCount *int, visited map[string]bool) ([]FileNode, error) {
	if *nodeCount >= 500 {
		return nil, nil
	}

	realDir, err := filepath.EvalSymlinks(currentDir)
	if err != nil {
		realDir = currentDir
	}
	if visited[realDir] {
		return nil, nil // 避免软链接循环递归
	}
	visited[realDir] = true

	// 确保没有越出工作区
	cleanReal, err := filepath.Abs(realDir)
	if err == nil {
		cleanWs, err := filepath.Abs(a.workspace)
		if err == nil {
			normWs := normalizeWindowsPath(cleanWs)
			normReal := normalizeWindowsPath(cleanReal)
			relWs, err := filepath.Rel(normWs, normReal)
			if err != nil || relWs == ".." || strings.HasPrefix(relWs, ".."+string(filepath.Separator)) || strings.HasPrefix(relWs, "../") {
				return nil, nil // 软链接指向工作区外部，阻断
			}
		}
	}

	entries, err := os.ReadDir(currentDir)
	if err != nil {
		return nil, err
	}

	nodes := make([]FileNode, 0, len(entries))
	for _, entry := range entries {
		if *nodeCount >= 500 {
			break
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") || name == "node_modules" || name == "bin" || name == "dist" || name == "build" {
			continue
		}
		rel, _ := filepath.Rel(a.workspace, filepath.Join(currentDir, name))
		rel = filepath.ToSlash(rel)

		*nodeCount++
		node := FileNode{
			Name:  name,
			Path:  rel,
			IsDir: entry.IsDir(),
		}
		if entry.IsDir() && currentDepth < maxDepth {
			subNodes, _ := a.buildFileTreeInternal(filepath.Join(currentDir, name), currentDepth+1, maxDepth, nodeCount, visited)
			node.Children = subNodes
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (a *App) ReadFile(relPath string) (string, error) {
	if a.sandbox == nil {
		return "", fmt.Errorf("sandbox not initialized")
	}
	data, err := a.sandbox.SafeReadFile(relPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (a *App) WriteFile(relPath string, content string) error {
	if a.sandbox == nil {
		return fmt.Errorf("sandbox not initialized")
	}
	return a.sandbox.AtomicWriteFile(relPath, []byte(content))
}

func (a *App) ExecCommand(command string) (string, error) {
	termTool, ok := a.registry.GetTool("tool.terminal")
	if !ok {
		return "", fmt.Errorf("terminal tool not registered in registry")
	}
	rawArgs, _ := json.Marshal(map[string]string{"command": command})
	res, err := termTool.Execute(a.ctx, rawArgs)
	if err != nil {
		return "", err
	}
	return res.Content, nil
}

// ExecTerminalStream 异步执行终端命令，通过 Wails 事件实时推送流式输出与退出码
func (a *App) ExecTerminalStream(command string) error {
	trimmed := strings.TrimSpace(command)
	if trimmed == "" {
		return fmt.Errorf("command cannot be empty")
	}
	if a.ctx == nil {
		return fmt.Errorf("context not initialized")
	}
	termPlugin, ok := a.registry.GetTool("tool.terminal")
	if !ok {
		return fmt.Errorf("terminal tool not registered in registry")
	}
	termTool, ok := termPlugin.(*terminaltool.Tool)
	if !ok {
		return fmt.Errorf("terminal tool type assertion failed")
	}

	a.terminalMu.Lock()
	if a.terminalCancel != nil {
		a.terminalCancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.termTaskID++
	currentTaskID := a.termTaskID
	a.terminalCancel = cancel
	a.terminalMu.Unlock()

	go func() {
		startTime := time.Now()
		runtime.EventsEmit(a.ctx, "terminal:start", map[string]any{
			"command":    command,
			"start_time": startTime.UnixMilli(),
		})

		exitCode, err := termTool.ExecuteStream(ctx, trimmed, func(chunk string) {
			runtime.EventsEmit(a.ctx, "terminal:data", chunk)
		})

		elapsed := time.Since(startTime).Milliseconds()
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}

		runtime.EventsEmit(a.ctx, "terminal:exit", map[string]any{
			"command":     command,
			"exit_code":   exitCode,
			"duration_ms": elapsed,
			"error":       errMsg,
		})

		a.terminalMu.Lock()
		if a.termTaskID == currentTaskID {
			a.terminalCancel = nil
		}
		a.terminalMu.Unlock()
	}()

	return nil
}

// CancelTerminalCommand 中断当前正在执行的终端命令
func (a *App) CancelTerminalCommand() {
	a.terminalMu.Lock()
	defer a.terminalMu.Unlock()
	if a.terminalCancel != nil {
		a.terminalCancel()
		a.terminalCancel = nil
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "terminal:data", "\n[Process interrupted by user]\n")
		}
	}
}

// CancelAgentStream 中断当前正在进行的大模型流式推理与自主工具循环
func (a *App) CancelAgentStream() {
	a.agentMu.Lock()
	defer a.agentMu.Unlock()
	if a.agentCancel != nil {
		a.agentCancel()
		a.agentCancel = nil
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "agent:interrupted", map[string]any{
				"session_id": a.currentSessionID,
				"message":    "用户手动中断了本次推理",
			})
		}
	}
}

package loop

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	v1 "tiancode/pkg/plugin/v1"
)

func trimToolOutput(output string, maxChars int) string {
	runes := []rune(output)
	if len(runes) <= maxChars {
		return output
	}
	half := maxChars / 2
	head := string(runes[:half])
	tail := string(runes[len(runes)-half:])
	return head + fmt.Sprintf("\n\n...[输出过长，中间 %d 字符已截断]...\n\n", len(runes)-maxChars) + tail
}

func (e *ExecutionEngine) runTool(ctx context.Context, sessionID, toolName string, rawArgs json.RawMessage, toolMap map[string]v1.ToolPlugin) (output string, isErr bool, written string) {
	rails := e.registry.ListRails()
	for _, rail := range rails {
		decision, railErr := rail.OnBeforeAct(ctx, sessionID, toolName, rawArgs)
		if railErr != nil || (decision != nil && !decision.Allow) {
			reason := fmt.Sprintf("%v", railErr)
			if decision != nil && decision.Reason != "" {
				reason = decision.Reason
			}
			return fmt.Sprintf("[安全拦截] 工具 [%s] 被 Rail 阻断: %s", toolName, reason), true, ""
		}
	}

	var result *v1.ToolResult
	if toolMap != nil {
		if impl, ok := toolMap[toolName]; ok {
			res, err := impl.Execute(ctx, rawArgs)
			if err != nil {
				return fmt.Sprintf("execution failure: %v", err), true, ""
			}
			if res == nil {
				return fmt.Sprintf("tool [%s] returned nil result", toolName), true, ""
			}
			result = res
			output = trimToolOutput(res.Content, 3000)
			isErr = res.IsError
		}
	}
	if result == nil {
		if tool, ok := e.registry.GetToolByName(toolName); ok {
			res, err := tool.Execute(ctx, rawArgs)
			if err != nil {
				return fmt.Sprintf("工具 [%s] 执行失败: %v", toolName, err), true, ""
			}
			if res == nil {
				return fmt.Sprintf("工具 [%s] 返回空结果", toolName), true, ""
			}
			result = res
			output = trimToolOutput(res.Content, 3000)
			isErr = res.IsError
		} else if e.MCPCall != nil {
			var mcpArgs map[string]any
			if len(rawArgs) > 0 {
				_ = json.Unmarshal(rawArgs, &mcpArgs)
			}
			if mcpArgs == nil {
				mcpArgs = map[string]any{}
			}
			mcpRes, err := e.MCPCall(ctx, toolName, mcpArgs)
			if err != nil {
				return fmt.Sprintf("MCP 算子 [%s] 执行失败: %v", toolName, err), true, ""
			}
			output = trimToolOutput(mcpRes, 3000)
		} else {
			return fmt.Sprintf("[未知工具] %s 未在 Registry 或 MCP 中注册", toolName), true, ""
		}
	}

	if result != nil {
		for _, rail := range rails {
			_ = rail.OnAfterAct(ctx, sessionID, toolName, result)
		}
		isWrite := toolName == "write_file"
		var argsObj struct {
			Action   string `json:"action"`
			RelPath  string `json:"rel_path"`
			Path     string `json:"path"`
			FilePath string `json:"file_path"`
		}
		_ = json.Unmarshal(rawArgs, &argsObj)
		if strings.ToLower(argsObj.Action) == "write" {
			isWrite = true
		}
		if isWrite && result != nil && !result.IsError {
			written = argsObj.RelPath
			if written == "" {
				written = argsObj.Path
			}
			if written == "" {
				written = argsObj.FilePath
			}
		}
	}
	return output, isErr, written
}

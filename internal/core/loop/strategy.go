package loop

import (
	"encoding/json"
	"strings"

	"tiancode/internal/llm"
)

const (
	StrategyAnalyze   = "analyze"
	StrategyImplement = "implement"
	StrategyTDD       = "tdd"
)

// NormalizeStrategy 只接受内核认识的三种策略，其余一律当成 implement。
func NormalizeStrategy(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case StrategyAnalyze, "readonly", "read_only":
		return StrategyAnalyze
	case StrategyTDD, "test":
		return StrategyTDD
	default:
		return StrategyImplement
	}
}

// ApplyStrategy 改写系统提示，并在只读策略下从模型可见工具表里拿掉会改磁盘的算子。
func ApplyStrategy(strategy, note string, tools []llm.ToolDef, system string) ([]llm.ToolDef, string) {
	s := NormalizeStrategy(strategy)
	note = strings.TrimSpace(note)
	switch s {
	case StrategyAnalyze:
		system += "\n[执行策略 analyze] 只允许读取与解释。禁止写入文件、禁止执行会改动工作区的命令、禁止 Git 写操作。"
		tools = filterTools(tools, func(name string) bool {
			n := strings.ToLower(name)
			if n == "exec_command" || n == "write_file" {
				return false
			}
			return true
		})
	case StrategyTDD:
		system += "\n[执行策略 tdd] 先运行或补齐测试，再改实现，直到测试通过。不要在测试失败时宣称完成。"
	default:
		system += "\n[执行策略 implement] 允许读写文件并执行必要命令完成任务。先说明要改什么，再调用工具。"
	}
	if note != "" {
		system += "\n[用户附加约束] " + note
	}
	return tools, system
}

// DenyByStrategy 在真正执行工具前再拦一层，避免模型无视提示词去写盘。
func DenyByStrategy(strategy, toolName string, rawArgs json.RawMessage) (deny bool, reason string) {
	s := NormalizeStrategy(strategy)
	if s != StrategyAnalyze {
		return false, ""
	}
	name := strings.ToLower(strings.TrimSpace(toolName))
	if name == "exec_command" {
		return true, "当前策略为只读分析，已拦截 exec_command"
	}
	if name == "write_file" {
		return true, "当前策略为只读分析，已拦截写文件"
	}
	if name == "fs_control" {
		var args struct {
			Action string `json:"action"`
		}
		_ = json.Unmarshal(rawArgs, &args)
		if strings.EqualFold(strings.TrimSpace(args.Action), "write") {
			return true, "当前策略为只读分析，已拦截 fs_control write"
		}
	}
	return false, ""
}

// ShouldVerifyAfterWrite TDD 策略在写盘后必须跑工作区测试，其它策略不自动跑。
func ShouldVerifyAfterWrite(strategy string) bool {
	return NormalizeStrategy(strategy) == StrategyTDD
}

// FormatVerifyFollowup 把测试结果缝进工具输出，下一轮模型能看见。
func FormatVerifyFollowup(file, output string, pass bool) string {
	if pass {
		return "[TDD 验证] 写入 " + file + " 后测试通过\n" + output
	}
	return "[TDD 验证失败] 写入 " + file + " 后测试未通过，必须继续修复，不得宣称完成\n" + output
}

func filterTools(tools []llm.ToolDef, keep func(name string) bool) []llm.ToolDef {
	out := make([]llm.ToolDef, 0, len(tools))
	for _, t := range tools {
		name := t.Function.Name
		if keep(name) {
			out = append(out, t)
		}
	}
	return out
}

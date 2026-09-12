package loop

import (
	"fmt"
	"strings"
)

// FormatHitCapNotice 格式化工具触顶人话收尾
func FormatHitCapNotice(turns int) string {
	return fmt.Sprintf("\n\n⚠️ 【系统提示】本轮工具自主调用已达安全上限（%d 轮），已停止继续调工具。当前阶段进展总结如下；若需接续完成剩余任务，请发送「继续」。\n", turns)
}

// FormatInterruptedNotice 格式化用户中止人话收尾
func FormatInterruptedNotice() string {
	return "\n\n🛑 【系统提示】任务已由用户手动中止。当前已保留已完成的操作与上下文。若需恢复执行，请发送「继续」或补充新指令。\n"
}

// FormatEmptyOutputNotice 格式化模型空输出人话收尾
func FormatEmptyOutputNotice() string {
	return "\n\n⚠️ 【系统提示】模型未返回任何有效文本或工具调用（响应为空内容）。这通常是上游模型内容风控拦截、或服务偶发空响应，建议重新发送或更换模型。\n"
}

// FormatUpstreamError 将底层网络或上游网关报错转换为结构化人话指引
func FormatUpstreamError(err error) string {
	if err == nil {
		return ""
	}
	raw := err.Error()
	lower := strings.ToLower(raw)

	switch {
	case strings.Contains(raw, "HTTP 400") || strings.Contains(lower, "bad request") || strings.Contains(lower, "invalid request"):
		return fmt.Sprintf("\n\n❌ 【上游请求失败 (HTTP 400)】\n- 根本原因: 上游网关判定请求格式或参数不合法，常见原因包括单条消息过长、上下文击穿模型限制、或不支持某些 Function Calling 参数。\n- 建议方案: 请尝试在顶栏选择新会话、或清空过多冗余历史后重试。\n- 原始错误: %s\n", raw)

	case strings.Contains(raw, "HTTP 401") || strings.Contains(raw, "HTTP 403") || strings.Contains(lower, "unauthorized") || strings.Contains(lower, "forbidden") || strings.Contains(lower, "invalid api key"):
		return fmt.Sprintf("\n\n❌ 【鉴权未通过 (HTTP 401/403)】\n- 根本原因: 上游端点拒绝了当前的 API Key，可能由于密钥无效、过期或当前渠道无权访问该模型。\n- 建议方案: 请点击侧边栏「设置 -> 模型与网关渠道」，检查并更正该渠道的 API Key 与 Base URL。\n- 原始错误: %s\n", raw)

	case strings.Contains(raw, "HTTP 429") || strings.Contains(lower, "rate limit") || strings.Contains(lower, "quota exceeded") || strings.Contains(lower, "insufficient quota"):
		return fmt.Sprintf("\n\n❌ 【接口频率或配额受限 (HTTP 429)】\n- 根本原因: 上游模型渠道请求频率超限、并发数触顶，或当前 API Key 账户配额已耗尽。\n- 建议方案: 请稍候 10~30 秒重试，或在设置中切换到备用模型渠道。\n- 原始错误: %s\n", raw)

	case strings.Contains(raw, "HTTP 500") || strings.Contains(raw, "HTTP 502") || strings.Contains(raw, "HTTP 503") || strings.Contains(raw, "HTTP 504"):
		return fmt.Sprintf("\n\n❌ 【上游服务内部故障 (HTTP 500/5xx)】\n- 根本原因: 上游模型服务发生内部故障或网关超时，非本地内核代码错误。\n- 建议方案: 上游节点当前处于不稳定状态，建议等待 1 分钟后重试或切换其他模型。\n- 原始错误: %s\n", raw)

	case strings.Contains(lower, "connection refused") || strings.Contains(lower, "dial tcp") || strings.Contains(lower, "timeout") || strings.Contains(lower, "no such host"):
		return fmt.Sprintf("\n\n❌ 【网络连接异常】\n- 根本原因: 无法与上游 API 端点建立 TCP/TLS 连接，可能本地网络断开、DNS 无法解析或目标端点地址有误。\n- 建议方案: 请检查网络连接、代理状态，并在设置中验证渠道 Base URL 是否可达。\n- 原始错误: %s\n", raw)

	default:
		return fmt.Sprintf("\n\n❌ 【执行异常: %s】\n请检查模型设置与执行参数。\n", raw)
	}
}

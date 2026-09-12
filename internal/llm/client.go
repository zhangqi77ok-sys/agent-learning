package llm

// ToolFunctionDef 算子定义
type ToolFunctionDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// ToolDef 顶层算子定义
type ToolDef struct {
	Type     string          `json:"type"`
	Function ToolFunctionDef `json:"function"`
	Mutating bool            `json:"-"`
}

// StreamHandlers 流式回调处理器
type StreamHandlers struct {
	OnThinking  func(text string)
	OnContent   func(delta string)
	OnToolStart func(id string, name string)
	OnToolArg   func(id string, argDelta string)
	OnDone      func()
	OnError     func(err error)
}

// ToolCall 算子调用
type ToolCall struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Function  struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// Request 请求结构
type Request struct {
	Endpoint string        `json:"endpoint"`
	APIKey   string        `json:"api_key"`
	Model    string        `json:"model"`
	Messages []Message     `json:"messages"`
	Tools    []ToolDef     `json:"tools,omitempty"`
}

// Message 对话消息
type Message struct {
	Role             string     `json:"role"`
	Content          string     `json:"content"`
	ReasoningContent string     `json:"reasoning_content,omitempty"`
	ToolCallID       string     `json:"tool_call_id,omitempty"`
	Name             string     `json:"name,omitempty"`
	ToolCalls        []ToolCall `json:"tool_calls,omitempty"`
}

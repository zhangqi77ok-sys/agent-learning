# Tcode 插件热插拔架构与自执行约束规范 (PLUGIN_ARCH.md)

## 1. 架构愿景与微内核哲学

Tcode 从第一天起坚持**微内核 + 热插拔插件**的架构范式：
- **微内核 (Microkernel / Host)**：负责生命周期管理、窗口宿主与注册调度，不持有任何具体算子业务。
- **插件体系 (Plugins)**：作为第一公民独立存在。支持大模型 Provider 插件、Tool 执行插件、Rail 治理插件与 Storage 存储插件。
- **自执行架构约束 (Self-Enforcing Architecture)**：通过“编译期阻断 + 静态检测拦截 + 文件契约上下文注入 + 架构文档兜底”四道防线，杜绝开发过程中 AI 因上下文丢失而造成的架构漂移与硬编码耦合。

---

## 2. 依赖方向与单向流转法则

```
┌─────────────────────────────────────────────────────────┐
│                      Plugins Layer                      │
│   (plugins/tool/*, plugins/provider/*, plugins/rail/*)  │
└────────────────────────────┬────────────────────────────┘
                             │ implements
                             ▼
┌─────────────────────────────────────────────────────────┐
│                    pkg/plugin/v1                        │
│   (Plugin, ToolPlugin, ProviderPlugin, RailPlugin...)   │
└────────────────────────────▲────────────────────────────┘
                             │ uses contract interfaces
┌────────────────────────────┴────────────────────────────┐
│                    internal/core                        │
│            (loop/engine.go, sandbox, ...)               │
└────────────────────────────▲────────────────────────────┘
                             │ drives
┌────────────────────────────┴────────────────────────────┐
│                      app.go / Host                      │
│             (Wails Desktop / HTTP Gateway)              │
└─────────────────────────────────────────────────────────┘
```

### 铁律规约：
1. **禁止反向依赖**：`internal/core` 严禁直接 `import "tiancode/plugins/..."` 具体包。
2. **禁止宿主耦合**：`plugins/` 严禁 `import "tiancode/app"` 或 `"tiancode/main"`。
3. **统一入口**：所有工具调度必须且只能经由 `host.Registry.GetTool(name).Execute(...)`。

---

## 3. 四道架构守卫防线

| 防线层级 | 机制 | 触发场景 | 阻断效果 |
|---|---|---|---|
| **防线一：编译期约束** | 移除 `App` 结构体中的 `*gittool.Tool`、`*fstool.Tool`、`*terminaltool.Tool` 等字段 | 任何尝试写 `a.termTool.Execute()` | 编译直接报错，不可绕过 |
| **防线二：提交时拦截** | `scripts/arch_check.ps1` 静态分析脚本 | Git pre-commit / CI 阶段 | 检测 switch/case 路由、非法 import、字段违规持有 |
| **防线三：上下文注入** | 关键文件头部契约注释 + `AGENTS.md` 铁律 7 + `.cursorrules` | AI 每次读取代码与任务触发时 | 保证即使上下文缩减，第一行也是架构硬约束 |
| **防线四：标准文档兜底** | `docs/architecture/PLUGIN_ARCH.md` | 开发指南与重构参考 | 提供正反例模板与标准扩展步骤 |

---

## 4. 新增插件标准开发步骤 (Step-by-Step)

若需新增算子（例如 `db_query` 数据库算子）：

### 步骤 1：创建插件目录
在 `plugins/tool/db/` 下建立 `db_tool.go`。

### 步骤 2：实现 `pkg/plugin/v1.ToolPlugin` 完整接口
```go
package db

import (
    "context"
    "encoding/json"
    v1 "tcode/pkg/plugin/v1"
)

type Tool struct {
    id string
}

func NewTool() *Tool {
    return &Tool{id: "tool.db"}
}

func (t *Tool) ID() string             { return t.id }
func (t *Tool) Name() string           { return "Database Query Tool" }
func (t *Tool) Version() string        { return "1.0.0" }
func (t *Tool) Type() v1.PluginType    { return v1.TypeTool }
func (t *Tool) Init(ctx context.Context, cfg json.RawMessage) error { return nil }
func (t *Tool) Start(ctx context.Context) error { return nil }
func (t *Tool) Stop(ctx context.Context) error  { return nil }
func (t *Tool) Health(ctx context.Context) v1.HealthStatus {
    return v1.HealthStatus{Healthy: true, Message: "DB ready"}
}
func (t *Tool) Definition() v1.ToolDefinition {
    // 声明大模型 Schema
    return v1.ToolDefinition{
        Name: "db_query",
        Description: "Query workspace SQLite/PostgreSQL database",
        Parameters: json.RawMessage(`{"type":"object","properties":{"sql":{"type":"string"}}}`),
    }
}
func (t *Tool) Execute(ctx context.Context, rawArgs json.RawMessage) (*v1.ToolResult, error) {
    // 执行业务逻辑
    return &v1.ToolResult{Content: "query ok", IsError: false}, nil
}
```

### 步骤 3：在 `NewApp()` 中注册进 Registry
```go
// app.go
_ = reg.Register(db.NewTool())
```

### 步骤 4：零侵入验证
无需修改 `SendMessage`，无需修改任何 switch/case，大模型与 ReAct 引擎自动具备调度 `db_query` 算子的能力，且自动受到 Rail 安全与审计拦截！

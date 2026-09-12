# 湉码 / tiancode 现行架构设计规范 (ARCHITECTURE.md)

> **现行唯一发货技术栈**：**Wails v2 + Go 微内核 + Vue 3 / Vite + TypeScript + Pinia**  
> ⚠️ **历史架构作废与归档声明**：所有涉及 Tauri v2 / Rust Core / React 19 / Python Daemon / HTML 原型的材料已全部归档至 `archive/` 目录，明确标记为历史探索产物，**严禁作为主路径代码或发货基准**。

---

## 🏛️ 一、现行分层解耦微内核全景图 (Current Microkernel Topology)

```mermaid
graph TB
    subgraph UI ["1. 表现层 (Presentation Layer · Vue 3 + TS + Pinia)"]
        ActivityBar["ActivityBar (侧边极简导航 · 工作区模式切换)"]
        LeftDrawer["LeftDrawer (工程文件树 · Git 状态 · 会话列表)"]
        ChatCockpit["ChatCockpit (智能体驾驶舱 · 顶栏项目宪法 · 流式对话 · 任务目标卡片)"]
        DiffWorkspace["DiffWorkspace (Monaco 差异审查 · Hunk 细粒度采纳/放弃 · 待确认阻断)"]
        TerminalDrawer["TerminalDrawer (原生沉浸式终端抽屉)"]
    end

    subgraph StoreLayer ["2. 表现层领域状态中枢 (Domain Stores)"]
        ChatStore["chatStore (多会话管理 · 消息历史 · 最小任务模型 TaskModel · 流式追踪)"]
        GitStore["gitStore (Git 状态 · 分支 · 差异报告 · PendingDiffFiles 追踪)"]
        SettingsStore["settingsStore (渠道管理 · 探活测速 · MCP stdio 服务 · 规则与技能库)"]
        Workbench["workbench.ts (总成调度中枢 · 跨领域事件协调)"]
    end

    subgraph IPCBridge ["3. Wails v2 原生桥接层 (app.go / app_chat.go / app_shell.go)"]
        WailsRuntime["Wails Runtime (EventsOn / EventsEmit / WindowControl)"]
        AppController["App 宿主对象 (会话持久化 · 目录对话框 · 命令中断控制)"]
    end

    subgraph Kernel ["4. Go 微内核自主调度引擎 (internal/core/loop)"]
        ExecutionEngine["ExecutionEngine (单一执行内核 · ReAct 双环调度)"]
        TaskModelEngine["Session Task 模型 (Goal / Status / ToolBudget / Summary)"]
        HumanEnding["Human Ending 解释器 (触顶收敛 · 用户中止 · 上游4xx/5xx转人话 · 空输出兜底)"]
        StrategyGuard["Strategy 策略拦截器 (Implement / Analyze 地图先验 / TDD 完成阻断)"]
    end

    subgraph SafetyRail ["5. 运行时防线 (plugins/rail)"]
        SafetyRailImpl["Safety Rail (OnBeforeAct 阻断高危命令 / OnAfterAct 审计)"]
        SecretMasker["Secret Masker (敏感凭据脱敏防泄漏)"]
    end

    subgraph PluginRegistry ["6. 热插拔插件生态 (pkg/plugin/v1 & host.Registry)"]
        Registry["host.Registry (插件注册与工具寻址中枢)"]
        subgraph ToolPlugins ["受控工具插件 (Tool Plugins)"]
            FSTool["tool.fs (沙箱文件读写/列表 · 写前影子快照)"]
            GitTool["tool.git (真实 Git 状态/暂存/检出/还原)"]
            TermTool["tool.terminal (无窗口静默外部进程执行)"]
            MCPStdio["MCP Manager (标准 JSON-RPC 2.0 Stdio 外部工具集成)"]
        end
        subgraph Providers ["模型驱动插件 (Provider Plugins)"]
            OpenAIProvider["provider.openai (OpenAI 兼容协议族 / AgentRouter 穿透)"]
        end
    end

    UI --> StoreLayer --> IPCBridge --> Kernel
    Kernel --> StrategyGuard
    Kernel --> TaskModelEngine
    Kernel --> HumanEnding
    Kernel --> SafetyRailImpl
    SafetyRailImpl --> Registry
    Registry --> ToolPlugins
    Registry --> Providers
```

---

## 🔄 二、单一执行内核与任务生命周期流转 (Execution Flow & Task Lifecycle)

```mermaid
sequenceDiagram
    autonumber
    actor Dev as 开发者 (User)
    participant UI as Vue 3 前端 (ChatCockpit)
    participant Bridge as Wails IPC (app_chat.go)
    participant Kernel as Go 内核 (loop.ExecutionEngine)
    participant Rail as Safety Rail (安全前置拦截)
    participant Tool as ToolPlugin / MCP Stdio
    participant Upstream as 大模型上游网关 (OpenAI / DeepSeek / Claude)

    Dev->>UI: 输入任务目标 (或发送「继续」)
    UI->>Bridge: SendMessage(ChatRequest)
    Note over Bridge: 检查任务模型：若为「继续」，注入上轮未完成目标与上下文，禁止推翻重勘
    Bridge->>Kernel: Execute(&EngineRequest)
    
    loop 自主推理与工具调用 (最多 24 轮)
        Kernel->>Upstream: StreamChat(messages, tools)
        alt 上游报错 (HTTP 400/401/429/500)
            Kernel->>UI: FormatUpstreamError 转为人话指引入库
            Note over Bridge: 标记 Task.Status = failed
        else 模型空输出 (0字符且无工具)
            Kernel->>UI: FormatEmptyOutputNotice 注入友好提示
            Note over Bridge: 标记 Task.Status = failed
        else 正常流式返回
            Kernel-->>UI: agent:chunk / agent:thinking
        end

        opt 模型触发工具调用
            Kernel->>Rail: OnBeforeAct(ctx, tool, args)
            alt 被安全策略或只读策略阻断
                Rail-->>Kernel: 阻断原因
            else 允许调用
                Kernel->>Tool: Execute(args)
                Tool-->>Kernel: ToolResult
                Kernel->>Rail: OnAfterAct(tool, result)
                opt 文件发生写入 (fs_control write)
                    Kernel-->>UI: agent:files_changed
                    Note over UI: 自动调出 Diff 面板，进入 PendingDiff 审查态
                end
            end
        end
    end

    alt 达到工具调用上限 (24轮)
        Kernel->>UI: FormatHitCapNotice 人话收敛总结并提示发「继续」
        Note over Bridge: 标记 Task.Status = capped
    else 用户主动点击中止 (Esc / Stop)
        Bridge->>UI: FormatInterruptedNotice 人话提示已中止
        Note over Bridge: 标记 Task.Status = interrupted
    else TDD 策略下测试失败
        Note over Bridge: 保持 Task.Status = tdd_failed，阻断宣称完成
    else 存在未采纳的 Diff 文件
        Note over Bridge: 保持 Task.Status = pending_diff，等待人工采纳
    else 全部验证通过且变更采纳
        Note over Bridge: 标记 Task.Status = completed
    end
    Bridge-->>UI: agent:done (携带 task_status 终态)
```

---

## 📐 三、核心架构设计与工程约束

### 1. 一条执行内核 (Single Execution Engine)
- 废除任何两套执行循环。所有大模型推理统一收敛至 `internal/core/loop/llm_path.go` 驱动的直接执行循环；
- 统一工具调用轮次上限为 24 轮，统一错误捕获、统一取消处理、统一事件管道派发。

### 2. Session 最小任务模型 (TaskModel)
- 每个会话实体 `session.ChatSession` 均强类型挂载 `TaskModel`：
  - `Goal`：用户原始目标（或提炼目标）；
  - `Status`：`idle` | `running` | `completed` | `capped` | `interrupted` | `failed` | `pending_diff` | `tdd_failed`；
  - `ToolBudget` / `ToolsUsed`：工具配额与已用次数计数器；
  - `Summary`：当前阶段探明成果与未完成项摘要；
  - `PendingDiffFiles`：智能体已修改但尚未经人工点击采纳的文件列表。
- **「继续」语义接续**：当用户输入「继续」时，底层自动唤醒未完成任务模型，注入接续提示词，绝对不重新扫描全局目录结构。

### 3. 审查类任务先地图再下钻规约 (Map-First Strategy)
- 在 `StrategyAnalyze`（及所有代码审查任务）中注入硬性约束：
  - **第一步（看地图）**：先读取项目顶层结构与关键配置文件（`go.mod`、`package.json`、`Cargo.toml`、`README.md`）；
  - **第二步（定靶向）**：定位核心模块入口与调用拓扑；
  - **第三步（精准下钻）**：仅读取靶向具体文件，严禁全库盲目递归扫描。

### 4. 人话收尾系统 (Human-Readable Endings)
- 严禁用户面对「突然没了」或空白气泡：
  - **触顶 (Capped)**：输出友好阶段性成果说明并提示发送「继续」；
  - **取消 (Interrupted)**：输出中止说明并保存已生成内容与上下文；
  - **上游错误 (HTTP 4xx/5xx)**：解析网络与协议错误，以人话指出原因（上下文超长/鉴权失败/限流/服务挂死）及解决建议；
  - **空回复 (Empty)**：输出空响应防御提示，指导重新提问或更换模型。

### 5. 文件变更 Diff 确认闭环
- 工具写入文件后，自动派发 `agent:files_changed`；
- 前端自动打开 Monaco Diff 工作区，展示行级差异与变更块（Hunk）；
- 用户点击「✓ 采纳变更」（`git add / stage`）或「✕ 放弃修改」（`git checkout`）后，任务才进入下一阶段或终态。

### 6. 项目宪法顶栏显式可见
- 将当前工作区生效的 Rules 规约与 Skills 技能库统计在对话顶栏（`📜 项目宪法`）；
- 开发者可随时展开抽屉/弹层，核对当前全流程注入大模型上下文的完整条文。

### 7. MCP 传输规范
- 微内核与 UI 仅承诺并实现标准 JSON-RPC 2.0 `stdio` 通信；
- 拒绝任何未实现的 SSE 伪宣传，确保工业级生产稳定性。

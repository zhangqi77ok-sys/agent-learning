# 湉码 / tiancode 研发守则 (AGENTS.md)

## 🚨 项目级绝对强制铁律 (Always-On Mandatory Iron Rules)

### 【铁律 0: 需求-原型-开发三位一体执行法则】
1. **双向强同步原则**：
   - 需求变更同步更新 `docs/`；视觉参考用 `archive/web_prototype.html`，**发货 UI 以 `frontend/src` 为准**。
2. **严禁越界开发禁令**：
   - 活路径是 Wails + Go + Vue。`archive/` 内 React/Python/Tauri 材料禁止当主路径施工。
   - 推荐：`需求 (docs/) -> frontend/src 实现 -> go test + vite build`。
3. **功能原型必须具备真实交互**：
   - 凡是涉及到功能特性的原型，**必须具备完整、真实、可点击、可感知的交互页面与状态流转**；
   - 杜绝静态图片拼接或不可交互的空壳假原型，必须让用户能够实际体验真实交互细节。

---

### 【铁律 0.5: 严禁假数据与 Demo 占位铁律 (Zero Demo & Fake Data Policy)】
1. **绝对禁止展示 Demo 数据与 Demo 伪功能**：
   - 严禁在前端界面、组件状态、后端存储或控制台输出中注入任何硬编码的 Demo 假数据（如假会话、预填的假对话气泡、假思考内容、硬编码的假插件列表、虚假日志等）；
   - 严禁为了“界面好看或看起来丰富”而添加伪造的模拟数据；
2. **真实状态与干净空状态原则 (Pure & Clean Empty State)**：
   - 新建项目、初次启动或无数据时，必须展示纯净、真实的空状态提示（例如：“当前暂无会话，请点击新建会话”或“未打开文件”），严禁静默预置假消息；
   - 插件列表必须读取 Go `host.Registry` 真实挂载的插件与真实连接的 MCP Server；
3. **严格 Fail-Closed 与零假成功 (No Fake Fallback)**：
   - 连通性测试与网络请求未配置或失败时，必须直接暴露真实错误，严禁返回假 200 OK、假延迟或假回复；
   - 严禁任何形式的静默降级与伪造成功。

---

### 【铁律 0.8: 活路径技能前置（Wails + Go + Vue）】
1. **强制先验查阅**：
   - 活路径是 **Wails v2 + Go 微内核 + Vue 3**，不是 Tauri / React / Rust。动手前优先：
     - **`ui-ux`**：暖色极简（`#FAF8F5` / `#D96B27`）、16:9、空状态干净；
     - **`tcode-studio-architect`**：Go Registry + Rail + 禁止在 `SendMessage` 里写死工具路由；
     - 聊天循环只允许改 `internal/core/loop`，宿主 `app.go` 只转发事件。
2. **归档栈**：`archive/`（含旧 `prototype/`、`src-desktop/`、`web_prototype.html`）仅历史参考。
3. **审查违规打回制**：把桌面主循环写回 `app.go`、或重新引入第二条 LLM 调度，一律打回。

### 【铁律 0.9: 后续施工只认一份合同】
1. 原型缺口与工作包顺序以 **`docs/AI_IMPLEMENTATION_CONTRACT.md`** 为准；旧 PRD / `docs/knowledge` / Tauri-Python 计划不是施工依据。
2. 按 WP-1 → WP-10 顺序；禁止跳号做 Swarm、OAuth、SSE、图谱主功能。
3. 未满足该文档「完成定义」不得声称完成，不得把半成品标成达标。

---

### 【铁律 1.5: 验证闭环（按改动范围，不必每次打安装包）】
- Go 逻辑：`go test ./...`；架构：`go run ./tools/archcheck`。
- 前端：`cd frontend && npm run build`。
- 安装包：改安装器或发版时执行 `powershell -File scripts/build-windows.ps1`。
- **真实凭据纪律**：API Key 只进 `~/.tiancode`（加密），**严禁写入源码或 skill**；未配置凭据必须 fail-closed。

---

### 【铁律 1: SDD + TDD 强制工程工作流】
- **触发条件**：任何涉及代码生成、功能实现、接口定义、Bug 修复或架构重构的请求；
- **执行规约**：必须自动激活并在首要环节调用 `.agents/skills/sdd-tdd-workflow/`；
- **严禁事项**：
  - 严禁在未输出 Spec 规范与契约前直接写业务逻辑；
  - 严禁在未编写前置测试前直接写实现代码；
  - 严禁先写代码后补测试用例；
  - 严禁过度封装，必须坚持积木解耦思想、Harness 治具思想与 ReAct 自愈循环。

---

### 【铁律 2: 暖色极简与人机工程学设计规范】
- 保持单一架构主轴（Wails v2 原生桌面端 + Go 微内核 + Vue 3 / TS）；
- 严格遵循暖米白（`#FAF8F5`）、工作台（`#F4EFEA`）、低饱和陶土暖橙（`#D96B27`）与代码暖炭黑（`#1E1C1A`）界面视觉规范；
- 模块间彻底解耦，依赖倒置，保持极简无冗余；
- 空间布局严格遵循 16:9 原生工作台人体工程学。

---

### 【铁律 3: 目录物理隔离与代码纯净性】
- `docs/`：需求与架构；
- `archive/`：停用的 React/Python/HTML 原型，禁止当发货代码；
- 根目录严禁堆砌散落临时代码与构建产物。

---

### 【铁律 4: 架构思路与技术亮点强同步法则 (README.md 强制更新)】
- **触发条件**：任何代码修改、功能开发、模块演进或架构重构完成后；
- **执行规约**：**每次写完代码后，必须无条件同步更新 `README.md`**，把最新的**架构设计思路（Architecture Rationale）+ 技术亮点（Technical Highlights）+ 核心数据流与关键设计**在 `README.md` 中详实表述；
- **严禁事项**：严禁只改代码不更新 `README.md`，严禁 `README.md` 仅包含空洞说明，必须让文档始终与真实工程演进保持高度同构与先进性。

---

### 【铁律 5: 弹窗与交互设计三维铁律 (Modals, Native Folder Dialog & Icon Tooltips)】
1. **弹窗视觉与交互规范**：
   - **严禁使用浏览器原生 `alert()` / `confirm()` / `prompt()`**；
   - 所有的弹窗必须符合项目 Warm Cream 暖米白（`#FAF8F5`）、工作台米灰（`#F4EFEA`）与陶土暖橙（`#D96B27`）样式；
   - 所有的弹窗必须在屏幕**严格水平垂直居中**；
   - 所有的弹窗必须支持 **Esc 快捷键退出**；
   - 所有的弹窗右上角必须拥有 **显式的关闭按钮 `[X]`**；
2. **系统原生文件夹选择**：
   - 选择项目或打开工作区文件夹时，**必须调用系统原生资源管理器文件夹选择框**（经由 Rust `rfd` 原生文件对话框），严禁弹出文本框让用户手输路径；
3. **图标优先与悬停提示**：
   - 界面上**能以图标显示的按钮/操作一律采用图标展示**，保持紧凑极简；
   - 所有图标和功能按钮**必须配置鼠标悬停说明（Tooltip / title 属性）**，清晰告知用户其用途。

---

### 【铁律 6: 知识点与解决方案强制沉淀归档法则 (Knowledge Base & Solution Vault)】
1. **知识点沉淀无遗漏**：
   - 凡在开发、调试、编译、构建、测试或运维过程中遇到的**每一个关键知识点、环境依赖陷阱、架构选型决策、高频编译/运行时报错**，**必须无条件在 `docs/knowledge/` 目录下创建对应的知识文档**；
   - 严禁“解决了问题但不写沉淀文档”或“仅在聊天记录中说明却未落盘仓库”。
2. **标准文档结构（四段论）**：
   每个知识点文档必须统一存放于 `docs/knowledge/` 对应分类子目录下，并严格包含以下核心章节：
   - **① 知识点与问题背景 (Context & Problem Statement)**：清晰阐述背景、遇到的问题现象、报错堆栈或需要解决的核心目标；
   - **② 核心原理与知识内容 (Knowledge Content & Root Cause)**：深入剖析底层机制、相关技术规范、协议细节或导致问题的根本原因；
   - **③ 标准解决方案与实操步骤 (Actionable Solutions & Step-by-Step Guide)**：给出详实、可立即执行、可验证的解决方案（包含命令、配置示例与修复代码）；
   - **④ 避坑指南与最佳实践 (Troubleshooting & Best Practices)**：总结防患于未然的经验与工程规范。
3. **全局目录索引同步更新**：
   - 每次新增或更新知识点文档时，必须同步在 `docs/knowledge/README.md` 中更新索引导航与分类标签，确保项目团队任何人均可按图索骥快速查阅。

---

### 【铁律 7: 插件热插拔架构强制执行（Plugin Hotplug Architecture Iron Rule）】

**背景**：Tcode 从第一天起的核心架构就是微内核 + 热插拔插件体系。任何破坏此架构的代码都是
技术债务，必须被打回重构。守卫脚本 `scripts/arch_check.ps1` 在每次提交时自动执行。

**强制规则（违反即打回重构，且 arch_check.ps1 会自动阻断提交）**：

1. **工具调用唯一路径**：
   - 所有工具执行必须通过 `registry.GetTool(name).Execute(ctx, args)` 调用；
   - 严禁在 `App` 结构体或任何业务层直接持有和调用具体 Tool 实例（如 `*gittool.Tool`）；
   - 若需调用插件专有方法（如 `StageFile`），通过 `registry.GetTool(name)` 获取后做类型断言，禁止字段持有。

2. **禁止 switch hardcode 路由**：
   - 严禁出现 `switch toolName { case "exec_command": ... }` 模式；
   - 工具路由由 Registry 自动完成，新增工具无需修改任何路由代码。

3. **新增工具的唯一合法方式**：
   ```go
   // Step 1: 在 plugins/tool/<name>/ 创建新目录
   // Step 2: 实现 pkg/plugin/v1.ToolPlugin 接口（ID/Name/Version/Type/Init/Start/Stop/Health/Definition/Execute）
   // Step 3: 在 NewApp() 中注册
   _ = reg.Register(mytool.NewTool(...))
   // Step 4: 完成。不需要也不允许修改 SendMessage 或任何业务代码
   ```

4. **Rail 拦截链必须生效**：
   - 工具执行前必须调用 `rail.OnBeforeAct()`，执行后调用 `rail.OnAfterAct()`；
   - 跳过 Rail 的工具调用视为安全漏洞。

5. **依赖方向约束（单向，禁止反转）**：
   ```
   plugins/ → pkg/plugin/v1 ← internal/ ← app.go
   ```
   - `internal/core/` 层禁止直接 import `plugins/tool/` 具体包；
   - `plugins/` 禁止 import `app` 或 `main` 包；
   - 违反依赖方向 = 严重架构违规。

**正确写法参考**：
```go
// ✅ 正确：通过 registry 统一调度
tool, ok := a.registry.GetTool(toolName)
if !ok {
    output = fmt.Sprintf("[未知工具] %s 未在 Registry 或 MCP 中注册", toolName)
} else {
    res, err := tool.Execute(ctx, rawArgs)
}

// ✅ 正确：需要专有方法时类型断言（仅限 transport/ 层）
gt, ok := s.registry.GetTool("tool.git")
if impl, ok2 := gt.(*gittool.Tool); ok2 {
    impl.StageFile(path)
}

// ❌ 错误：直接持有字段并调用
type App struct { termTool *terminaltool.Tool }  // 禁止
a.termTool.Execute(ctx, rawArgs)                  // 禁止
switch toolName { case "exec_command": ... }       // 禁止
```

**验证**：每次提交前运行 `powershell -File scripts/arch_check.ps1`，6 条规则全通过才允许推送。

# 44. 工作区动态重绑定、Tab 内存持久化、TDD 双栈闭环与文件树按需懒加载 (Workspace Rebind, Tab Persistence, Dual-Stack TDD & Lazy File Tree)

## ① 知识点与问题背景 (Context & Problem Statement)

在湉码桌面端从“功能型 Agent 壳”向“专业级 Coding 工作台”演进的过程中，开发者在真实重度使用时暴露了五处关键工程缺陷：
1. **工作区热切换漏绑定检索算子**：在 `app.go` 的 `SetWorkspace` 中重新实例化了 `gittool`、`fstool`、`terminaltool`，但遗漏了 `searchtool`。导致用户切换项目后，智能体的 `search_workspace` 算子仍死锁在旧项目根目录下跨目录搜索；
2. **多文件标签页（Tabs）切页丢草稿**：切换 Monaco 标签页时无脑调用 `loadEditor()` 强行从物理磁盘重读覆盖 `editorContent`，导致用户正在输入且尚未保存的代码被静默丢弃；
3. **关闭脏文件原生弹窗违规与静默关闭**：关闭未保存标签页时缺少拦截，或使用原生 `confirm()` 违背铁律 5；
4. **TDD 自动化双栈二选一互斥偏见**：`RunTDDValidation` 使用 `if hasGoMod ... else ...`，在全栈项目（如 湉码 自身：Go 后端微内核 + `frontend/package.json` 前端界面）中，一旦存在 `go.mod` 就彻底忽略 `npm test`，前端单测全挂也能谎称 PASS；
5. **文件树启动全量扫描 12 层导致 DOM 爆炸与 2000 节点截断**：大项目中文件众多，一次性扫描 12 层撑爆 Vue 虚拟 DOM，并提前触发 2000 节点全局截断导致深层目录丢失。

---

## ② 核心原理与知识内容 (Knowledge Content & Root Cause)

### 1. 微内核插件注册表的动态覆盖与生命周期绑定
在微内核 + 插件架构中，每个工具实例（如 `searchtool.Tool`）都绑定了一个沙箱受控根目录。工作区切换属于物理上下文重载，必须全量调用 `RegisterOrReplace` 刷新全部工程算子；此外，`engine.Verify` 闭包若在应用启动时捕获了旧的工作区字符串，将无法随着工作区变更而动态调整 TDD 执行靶心。

### 2. 编辑器多标签页内存缓冲协议 (Buffer Cache Protocol)
IDE（如 VS Code）的标签页模型必须维持单向数据流与内存缓存：
- 每个 Tab 实体拥有 `{ path, title, dirty, content }`；
- 当活动 Tab 切换前，活动编辑器的 `editorContent.value` 必须持久化同步至当前 Tab 的 `content` 缓冲；
- 激活目标 Tab 时，若目标 Tab 内存中已存在 `content`，则直接载入内存缓冲与 `dirty` 状态，严禁无脑重读物理磁盘。

### 3. 混合全栈 TDD 验证级联汇流机制
真实工程常为混合语言栈（Go + Vue/React，Rust + Node）。TDD 驱动验证引擎必须是级联复合型：
- 探测 Go 模块：`os.Stat("go.mod")`；
- 探测 Npm 模块：`package.json`（工作区根目录或 `frontend/package.json`）；
- 若两者同时存在，串行执行 `go test -v ./...` 与 `npm test`，分别解析各自的通过与失败数，聚合日志并输出分块报告；
- 任一套件失败即判定总状态为 `FAIL`，彻底杜绝局部失败被掩盖。

### 4. 虚拟文件树的受控懒加载 (Lazy Load on Expand)
大型仓库（如包含数十个微服务或数百个组件目录）禁止全量深度展开。初始仅加载深度 1（根目录与直接子项）；用户在界面上点击折叠箭头 `▶` 展开目录时，若未加载过子级，异步触发 `wailsBridge.getFileTree(node.path)` 请求该子目录的一级内容并挂载，将 DOM 节点数从数万降低至数百。

---

## ③ 标准解决方案与实操步骤 (Actionable Solutions & Step-by-Step Guide)

### 1. `app.go` 工作区重绑定与验证闭环动态化
```go
// app.go: SetWorkspace
if a.registry != nil {
    _ = a.registry.RegisterOrReplace(gittool.NewTool(absDir))
    _ = a.registry.RegisterOrReplace(fstool.NewTool(sb, a.snapshotMgr))
    _ = a.registry.RegisterOrReplace(terminaltool.NewTool(absDir))
    _ = a.registry.RegisterOrReplace(searchtool.NewTool(sb)) // 补齐检索算子重载
}

// 动态绑定 a.workspace 确保 TDD 验证靶向新工作区
ws := absDir
a.engine.Verify = func(writtenFile string) (string, bool) {
    report, err := agent.RunTDDValidation(ws)
    if err != nil {
        return err.Error(), false
    }
    out := report.Output
    if writtenFile != "" {
        out = writtenFile + "\n" + out
    }
    return out, report.Status == "PASS"
}
```

### 2. `workbench.ts` 标签页草稿持久化与脏状态拦截
```typescript
// workbench.ts: switchEditorTab & openEditorTab
function switchEditorTab(filePath: string) {
  if (!filePath || filePath === activeDiffFile.value) return
  // 切换前持久化暂存当前活动 tab 的编辑状态与内容
  if (activeDiffFile.value) {
    const currentTab = openEditorTabs.value.find(t => t.path === activeDiffFile.value)
    if (currentTab) {
      currentTab.content = editorContent.value
      currentTab.dirty = editorDirty.value
    }
  }

  activeDiffFile.value = filePath
  const targetTab = openEditorTabs.value.find(t => t.path === filePath)
  if (targetTab && targetTab.content !== undefined) {
    editorContent.value = targetTab.content
    editorDirty.value = !!targetTab.dirty
    void Promise.all([loadDiff(), refreshDiagnostics(filePath)])
  } else {
    void Promise.all([loadDiff(), loadEditor()])
  }
}
```

### 3. `DiffWorkspace.vue` 暖色模态窗安全关闭脏文件
```vue
<!-- 未保存文件关闭确认弹窗 (严格遵循铁律 2 与铁律 5) -->
<div
  v-if="s.pendingCloseTab"
  class="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 backdrop-blur-xs"
  @click.self="s.pendingCloseTab = null"
  @keydown.esc="s.pendingCloseTab = null"
>
  <div class="w-[420px] bg-[#FAF8F5] border border-black/[0.12] rounded-xl shadow-2xl p-5 flex flex-col gap-3">
    <!-- 取消 / 放弃并关闭 / 保存并关闭 -->
  </div>
</div>
```

### 4. `internal/agent/swarm.go` TDD 双栈级联执行
```go
// 1. Go 测试
if hasGoMod {
    goOut, goErr := runCmdWithTimeout(ctx, workspace, goExe, "test", "-v", "./...")
    ...
}
// 2. Npm 测试 (智能探测根目录与 frontend/ 目录)
if hasNpmTest {
    npmOut, npmErr := runCmdWithTimeout(ctx, npmDir, npmExe, "test")
    ...
}
// 汇总总通过与总失败
status := "PASS"
if totalFailed > 0 {
    status = "FAIL"
}
```

### 5. `app_shell.go` 与 `FileTreeNode.vue` 按需异步懒加载
```go
func (a *App) GetFileTree(dir string) ([]FileNode, error) {
    ...
    // 按需懒加载：初次加载与按目录加载均采用受控深度 (1 层)，由前端点击目录时异步下钻加载
    return a.buildFileTree(targetDir, 0, 1)
}
```

---

## ④ 避坑指南与最佳实践 (Troubleshooting & Best Practices)

1. **避免闭包中捕获旧工作区字符串**：
   - 在构造 `Verify` 或外部工具调用时，务必动态访问 `app.workspace`，禁止在初始化函数中用局部变量捕获后成为死引用的陈旧路径。
2. **Tab 切换时慎用直接写盘**：
   - 用户切 Tab 只是浏览其他代码，不等于用户希望保存；因此必须使用内存缓冲 `tab.content`，仅在用户按下 `Ctrl+S` 或点击「保存」时才写盘。
3. **禁止原生弹窗与静默丢弃**：
   - 严禁调用 `window.confirm`，弹窗必须遵循统一 Warm Cream 视觉设计规范，具备 Esc 快捷键与显式 `[X]` 按钮。
4. **混合语言栈项目单测必须串行执行并汇总**：
   - 在拥有前后端分离或多微服务的项目中，不能假定只有单一测试工具。`findNpmTestDir` 需兼容常见子目录布局，杜绝遗漏。

# 45 - 开发者全局检索打通、策略跳过守卫与 V1 边界矩阵对齐 (Knowledge Base #45)

## ① 知识点与问题背景 (Context & Problem Statement)

在 湉码 (tiancode) 的研发迭代中，虽然内核已经实现了大模型专用的 `search_workspace` (grep/find) 算子、单执行回路收敛以及双栈 TDD 验证，但在真实的开发者日常使用中暴露出了关键断点：
1. **开发者视角下的“盲人摸象”**：大模型拥有 `search_workspace` 工具可以随时在受控沙箱内搜索任意代码片段或文件，但左侧边栏的文件树是单层懒加载的，人类开发者在界面上无法搜寻尚未点开的深层文件，只能被动依赖智能体帮查；
2. **策略选择模态窗的“跳过”逻辑自相矛盾**：在策略推荐弹窗中，开发者点击“跳过”按钮，原本代码却设置为了直接改代码（`implement`），从而把本该是安全只读的会话意外提升为可能改文件的会话，导致用户不得不继续点取消或打回；
3. **模型在 implement 模式下可能盲写盲改**：模型在获得写盘权限时若未强制要求其先行阅读或检索上下文，可能出现根据记忆或猜测直接写入代码的“盲改”现象；
4. **V1 特性边界矩阵虚高**：此前矩阵中大量打上 🟢“达标/可用/交付”，掩盖了 LSP 不全、无跨文件重构、Diff 仅为写后审查、Git 冲突未闭环等真实存在的半成品状态；
5. **AST 图谱与用量看板占领活动栏造成主次不分**：仅针对 Go 单一语言的 AST 分析和纯统计展示被放在核心活动栏，给用户造成“全语言成熟 IDE”的错觉。

---

## ② 核心原理与知识内容 (Knowledge Content & Root Cause)

### 1. Wails 原生微内核算子复用与解耦调用 (Rule 7 铁律遵从)
在架构铁律第 7 条中，要求所有工具调度必须统一经过 `registry.GetTool(name).Execute(ctx, args)`。
为了向前端 UI 暴露开发者全局检索，不能在 `app.go` 或前端硬编码文件扫描逻辑，而是应当在应用壳层 (`app_shell.go`) 中实现 `SearchWorkspace(action, query, maxResults)`，通过 `a.registry.GetTool("tool.search")` 获取已有的安全沙箱检索算子，并传入序列化 JSON 参数。这样不仅复用了已有的多重安全沙箱防御（路径 canonicalize、忽略 .git/node_modules 等），还完全避免了层级破坏。

### 2. 状态机的默认安全底线 (Fail-Safe Defaults)
根据“最小特权”与“渐进授权”原则，只读审查 (`analyze`) 是系统唯一合法的默认态。
当用户在策略建议弹窗中点击“跳过”时，意图是“不采纳智能体推断的新策略”，此时系统的正确响应是保持底线的只读审查状态，而非擅自升级至写盘（`implement`）。

### 3. 双栈 TDD 的真跑动闭环
针对 Go + 前端混合架构，`RunTDDValidation` 探测到 `frontend/package.json` 时，要求其必须具备真实的 `npm test` 脚本。如果 package.json 缺少 test 脚本，或者测试超时阈值过低（例如 60s），在 Windows 平台冷启动下就容易出现挂起或失败。

---

## ③ 标准解决方案与实操步骤 (Actionable Solutions & Step-by-Step Guide)

### 1. 落地人类开发者全局检索视窗
在 `app_shell.go` 暴露统一搜索接口：
```go
func (a *App) SearchWorkspace(action, query string, maxResults int) ([]SearchMatch, error) {
    tool, ok := a.registry.GetTool("tool.search")
    if !ok {
        return nil, fmt.Errorf("search tool not registered")
    }
    // 组装 JSON 参数并调用 tool.Execute
    res, err := tool.Execute(ctx, rawArgs)
    ...
}
```
在前端 `workbench.ts` 中管理检索状态，并在 `DiffWorkspace.vue` 侧边栏提供 `📁 目录` 与 `🔍 检索` 双页签。

### 2. 修正策略选择器跳过逻辑
在 `frontend/src/App.vue` 中，将跳过逻辑由 `setExecutionStrategy('implement')` 修正为：
```ts
const skipStrategy = () => {
  workbenchStore.setExecutionStrategy('analyze')
  workbenchStore.dismissPendingStrategy()
}
```

### 3. 在 implement 策略提示词中注入前置检索要求
在 `internal/core/loop/strategy.go` 中，向 `StrategyImplement` 注入明确的防盲写要求：
```text
修改代码前必须先检索定位：优先使用 search_workspace (grep/find) 或 read_file 查明现有上下文与代码定义，严禁在未检索或未阅读目标文件的情况下盲改盲写！
```

### 4. 调整边界矩阵与活动栏实验标记
- 在 `ActivityBar.vue` 中，为用量监控与 AST 代码拓扑追加 `[实验特性]` 提示与 `β` 徽标；
- 在 `docs/V1_FEATURE_BOUNDARY_MATRIX.md` 中，严格落实验收口径：仅保留单文件安装器与沙箱安全为 🟢 可用，其余全部据实降级为 🟡 半成品或 🔴 实验特性。

---

## ④ 避坑指南与最佳实践 (Troubleshooting & Best Practices)

1. **避免在前端写死文件遍历**：在桌面应用中，文件系统扫描永远应当下沉到后端/内核执行，前端仅接收分页/限流后的匹配结果，防止几万个文件卡死浏览器主渲染线程；
2. **遵守单一路由铁律**：UI 调用功能（如检索）时，必须走统一注册表，不要重新搞一套独立的搜索函数，否则会导致配置（如忽略目录、沙箱隔离）在不同入口表现不一致；
3. **诚实面对技术成熟度**：对于探索性或仅支持单一语言的特性，必须在 UI 上明确标出实验性，避免误导开发者，影响主力开发流程的可靠度。

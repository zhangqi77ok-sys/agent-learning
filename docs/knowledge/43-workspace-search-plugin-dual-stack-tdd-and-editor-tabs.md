# 知识点沉淀 43: 工作区搜索算子架构、双栈 TDD 真实探测、递归文件树过滤与编辑器多页签

---

## ① 知识点与问题背景 (Context & Problem Statement)

在 Coding 桌面工作台日常开发中，前期草案主要暴露了三类严重影响可用性的结构性痛点：
1. **Agent 盲扫与上下文击穿**：缺少一等搜索算子，智能体面对任务时只能递归反复调用 `fs_control list` 枚举工作区全部目录，导致长上下文窗口被目录列表迅速撑爆触顶；
2. **TDD 虚假红绿灯与单栈绑定**：旧版 `RunTDDValidation` 仅识别 `go.mod`，且在缺少测试文件或非 Go 工程时静默跳过并返回假 `PASS`，误导模型和用户以为测试通过；
3. **前端资源管理器与编辑器简陋**：文件树原先写死只能展开 2 层且缺乏过滤，多层工程文件无法访问；Monaco 编辑器只能单文件打开，无多标签页管理，且在存在未确认代码变更时允许直接发送新轮次导致代码被冲刷覆盖。

---

## ② 核心原理与知识内容 (Knowledge Content & Root Cause)

1. **热插拔搜索算子设计 (`plugins/tool/search`)**：
   - 依据铁律 7 插件规范，实现 `pkg/plugin/v1.ToolPlugin` 接口，定义 `search_workspace` 算子；
   - 支持 `grep`（文本模式匹配）与 `find`（文件名通配）；
   - 在受控沙箱根目录内使用 `filepath.WalkDir` 高效扫描，利用 `shouldSkipDir` 与 `isBinaryExtension` 物理隔离无关中间产物（`.git`、`node_modules`、`bin`、`dist`、二进制静态资源）；
   - 对单行超长文本进行 200 字符紧凑截断，单次返回行数严格受控（默认 30，上限 100），极大降低模型 Token 消耗。

2. **双栈真实 TDD 红绿灯引擎 (`internal/agent/swarm.go`)**：
   - 彻底废止无测试时的“假成功”；
   - 优先探测 `go.mod`：调用 `go test -v ./...` 解析 `--- PASS:` 与 `--- FAIL:`；
   - 若无 `go.mod` 则探测 `package.json`：解析 `scripts.test` 脚本，存在则调用 `npm test`；
   - 若两者均不存在，直接判定 `Status: "FAIL"`，明确报错并要求前置测试用例，并在顶栏亮起 `❌ 测试未通过` 徽章。

3. **递归树状组件与多文件标签页管理**：
   - 提取 `FileTreeNode.vue` 递归组件，解除写死 2 层限制，支持 12 层任意深度展开；
   - 增加文件名实时模糊匹配过滤与 Git 状态角标映射（`[M]` / `[A]` / `[?]`）；
   - 在 `workbench.ts` 中管理 `openEditorTabs` 状态机，联动 Monaco 编辑器、未保存脏标记（●）与 Diff 工作区；
   - 引入 `isPendingDiffPromptOpen` 发送守卫，工作区存在未确认变更时强制阻断发送，杜绝代码被模型新一轮推理覆盖。

---

## ③ 标准解决方案与实操步骤 (Actionable Solutions & Step-by-Step Guide)

### 1. 注册工作区检索工具
```go
// app.go & backend/cmd/tcode-daemon/main.go
import searchtool "tiancode/plugins/tool/search"

_ = reg.Register(searchtool.NewTool(sb))
```

### 2. 双栈 TDD 验证核心逻辑
```go
// internal/agent/swarm.go
if !hasGoMod && !hasPkgJsonTest {
    return TestReport{
        Status: "FAIL",
        Failed: 1,
        Output: "当前工作区未检测到可执行的自动化测试套件。TDD 模式严禁无测试直接通过。",
    }, nil
}
```

### 3. 前端发送守卫拦截
```ts
// frontend/src/stores/workbench.ts
if (pendingDiffFiles.value.length > 0 && !forceSendWithPendingDiff.value) {
    isPendingDiffPromptOpen.value = true
    return
}
```

---

## ④ 避坑指南与最佳实践 (Troubleshooting & Best Practices)

1. **Pinia Store 内部禁止使用 `export` 关键字**：
   - 在 `defineStore('workbench', () => { ... })` 内部直接写 `export interface Foo` 会触发 esbuild 编译错误；接口与类型声明必须置于文件顶层导出。
2. **测试进程 Windows 僵尸进程防护**：
   - Windows 下 `exec.CommandContext` 仅杀死顶层父进程，必须配置 `taskkill /F /T /PID` 才能彻底递归终结 `.test.exe` 派生测试进程树。
3. **软链接防环防护**：
   - 遍历工作区目录树时必须调用 `filepath.EvalSymlinks` 与 `visited` 哈希表，防止循环软链接导致递归死循环爆栈。

# 湉码工程任务板

原则：**不做 Demo**。没有真实渠道就不推理；没有工作区就不假装有仓库；没有技能就不注入。每完成一项独立功能就 `git push`。

活路径：Wails v2 + Go 微内核 + Vue 3（`app_*.go` / `internal/` / `frontend/src`）。

**后续 AI 只认这份施工合同（原型缺口工作包 WP-1…WP-10）**：[`docs/AI_IMPLEMENTATION_CONTRACT.md`](./AI_IMPLEMENTATION_CONTRACT.md)

**安全/质量整改（合同补充卷 WP-R1…R6）**：[`docs/REVIEW_REMEDIATION_HANDOFF.md`](./REVIEW_REMEDIATION_HANDOFF.md)

**人机协同（对话中选择题 / 危险命令允许一次 WP-H1…H2）**：[`docs/HITL_CONTRACT.md`](./HITL_CONTRACT.md)

## 正在做 / 下一波（按序）

| ID | 功能 | 解决什么 | 验收 |
|----|------|----------|------|
| F1 | Fail-closed 零假数据 | 去掉硬编码模型列表、假仓库名 `agent-learning`、无渠道时仍填 `api.openai.com` | 未配渠道时下拉为空并提示；抽屉显示真实工作区名；无 Key 不发网 |
| F2 | Skill 注入推理 | 设置里保存的技能目前不进 prompt（字段还写错成 `content`） | 启用的 skill.prompt 进入 system prompt；落盘字段为 `prompt` |
| F3 | `/` 斜杠指令走真工具 | 输入 `/test` `/diff` 目前只是菜单文案 | `/test` 调 TDD；`/diff` 打开真实 Git diff |
| F4 | 渠道模型即唯一模型源 | 模型下拉必须来自主渠道或「拉取上游」结果 | 禁止写死 gpt-4o 等 |
| F5 | Provider 插件接聊天 | 桌面仍走 `llm.StreamChat`，注册的 openai Provider 闲置 | Execute 用渠道凭据 Init Provider |
| F6 | 流式列表性能 | 长会话整页重绘 | 只更新最后一条；超长列表窗口化 |
| F7 | 发版流水线 | 安装包只在本地脚本 | tag `v*` 打 `Tiancode_Setup`（可选） |

## 明确不做 / 暂缓项（聚焦核心编码主旅程）

- **暂缓（P2 Could）**：Swarm 复杂多 Agent 运行时、技能市场、OAuth 矩阵、知识图谱力导向图。主旅程稳定前不做，避免稀释产品焦点。
- **严禁事项**：
  - 严禁在没有多 Agent 运行时之前画 Swarm 空算子；
  - 严禁把 `archive/` 里的 Tauri/React/Python/HTML 接回主路径；
  - 严禁预置假会话、假插件、假指标等任何 Demo 数据；
  - 严禁堆砌 PRD 文档而不更新现行代码与规格。

## 最近完成记录 (P0 / P1 核心修复与微内核收敛)

- **P0-1 任务必须有结尾**：触顶（24轮上限阶段总结并提示发「继续」）、取消（保留内容并显式人话通知）、上游报错（解析HTTP 400/401/429/500为人话排查指引）、空结束（输出防御人话提示），彻底消除「突然没了」；
- **P0-2 审查任务先地图再下钻**：在 `StrategyAnalyze` 注入「先看地图（顶层结构+关键入口配置），再定靶向下钻」铁律，严禁盲目扫描全库；
- **P0-3 「继续」精准接续**：引入 Session 挂载的最小任务模型 `TaskModel`，当用户发送「继续」时无损承接既定目标与未完成项，禁止推翻重来；
- **P0-4 现行架构图收敛**：对外只承认 Wails + Go + Vue 3，彻底重写 `docs/ARCHITECTURE.md`，历史栈一律标归档；
- **P1-1 改文件默认出 Diff**：工具写盘后自动调出 Monaco Diff 工作区，跟踪 `PendingDiffFiles`，开发者点击采纳或放弃才算完成；
- **P1-2 TDD 失败状态流转**：TDD 测试未通过时，任务模型状态保持为 `tdd_failed`，严禁宣称任务完成；
- **P1-3 项目宪法顶栏显式可见**：在对话顶栏常驻展示 `📜 项目宪法`（启用的规约与技能统计），点击可展开完整条文；
- **P1-4 MCP 仅承诺 stdio**：内核与界面统一标准 JSON-RPC 2.0 stdio 通信，无假 SSE 承诺；
- **架构收敛**：执行内核合并为单一直接执行回路，清理冗余第二套循环；表现层收敛单一真实源 `workbench.ts`，清理未接线孤儿 Store。

## 完成记录

- 仓库改名 tiancode、单内核、SafetyRail、密钥加密、`~/.tiancode`
- Vue 壳拆分、`app.go` 拆文件、归档死栈、Windows 构建脚本
- F1 Fail-closed：无渠道不填假 OpenAI 地址；模型列表只来自渠道；抽屉用真实工作区名
- F2 Skill：启用技能的 `prompt` 注入 system prompt；保存不再误写 `content`
- F3 `/test` `/tdd` 跑真实 TDD；`/diff` 打开真实 Git 抽屉（未连接内核则报错，不假装 main 分支）
- F4 ListModels / 拉取上游：无 Key 报错，只返回网关真实 `/models`，去掉内置 gpt-4o 目录
- F5 桌面聊天 Init 已注册 Provider 再 StreamChat，不再绕过插件走裸 `llm.StreamChat`
- F6 对话默认只渲染最近 80 条，可一键展开全文（全量仍落盘）
- 会话模型：按工作区隔离；空草稿不落盘；标题取首条用户消息；列表不出现空会话
- 原型树：多项目折叠 + 项目下会话分支；打开项目；点会话切换工作区
- 修复桌面白屏：Vite `base: './'`，避免 Wails 加载 `/assets` 404
- 修复启动报错：workbench 误调用未导入的 `watch`/`store`，启动失败会整页替换成红字；Git 非仓库不再当致命错误

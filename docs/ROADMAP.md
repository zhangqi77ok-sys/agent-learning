# 湉码工程任务板

原则：**不做 Demo**。没有真实渠道就不推理；没有工作区就不假装有仓库；没有技能就不注入。每完成一项独立功能就 `git push`。

活路径：Wails v2 + Go 微内核 + Vue 3（`app_*.go` / `internal/` / `frontend/src`）。

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

## 明确不做（避免假功能）

- Swarm `budget()/parallel()` 在没有真实多 agent 运行时之前，不画空算子。
- 不预置假会话、假 MCP、假插件列表。
- 不把 `archive/` 里的 Tauri/React/Python 接回主路径。

## 完成记录

- 仓库改名 tiancode、单内核、SafetyRail、密钥加密、`~/.tiancode`
- Vue 壳拆分、`app.go` 拆文件、归档死栈、Windows 构建脚本
- F1 Fail-closed：无渠道不填假 OpenAI 地址；模型列表只来自渠道；抽屉用真实工作区名
- F2 Skill：启用技能的 `prompt` 注入 system prompt；保存不再误写 `content`
- F3 `/test` `/tdd` 跑真实 TDD；`/diff` 打开真实 Git 抽屉（未连接内核则报错，不假装 main 分支）
- F4 ListModels / 拉取上游：无 Key 报错，只返回网关真实 `/models`，去掉内置 gpt-4o 目录

# 湉码 / tiancode — 人机协同（HITL）施工合同

> **读者**：后续 AI 开发者。  
> **目的**：补上 Grok/Cursor 式「对话中途停下来让用户选」——不是再加快捷按钮。  
> **与其它文档关系**：  
> - 原型缺口 WP-1…WP-10：`docs/AI_IMPLEMENTATION_CONTRACT.md`（已完成，不要重做）  
> - 安全整改 WP-R1…R6：`docs/REVIEW_REMEDIATION_HANDOFF.md`（可并行，HITL 不要挡住 TLS 探活）  
> - 本文：`WP-H1` 选择题、`WP-H2` 危险工具一次授权。  
> **栈**：Wails v2 + Go 微内核 + Vue 3。发货只改 `internal/core/loop`、`app_chat.go`、`frontend/src/**`、`pkg/plugin/v1`（仅当 H2 扩展 RailDecision）。

视觉参考（只读）：`archive/web_prototype.html` 中选项卡（`pickChoice` / 确定提交）与危险命令「拒绝 / 显式授权」。**禁止把该 HTML 的 toast Demo 拷进产品。**

---

## 0. 产品要什么

用户在**同一轮对话未结束时**必须能：

1. 看到模型给出的 2～5 个互斥方案，点选一个（可加一句补充），点确定后循环继续。  
2. 看到「这条危险命令被拦住了」，选 **拒绝（默认）** 或 **允许这一次**；Esc / 超时 = 拒绝。

发送前的策略弹窗、写盘后的 Diff 横幅**保留**，不能代替上述两条。

**非目标**：点赞点踩、追问芯片墙、完整 Plan Mode、Swarm 投票、原生 `confirm()`。

---

## 1. 架构铁律

| ID | 规则 |
|----|------|
| H-A1 | 循环必须能 **暂停**：发出 choice/confirm 后本 goroutine **阻塞等待**，不得假装结束 `agent:done`。 |
| H-A2 | 恢复必须走显式 IPC：`ResumeAgentChoice` / `ResumeAgentConfirm`。禁止前端把选项当新的 `SendMessage` 另开一轮（会取消当前 `agentCancel`）。 |
| H-A3 | 默认 **fail-closed**：无人应答、Esc、会话切换、`CancelAgentStream` → 视为跳过推荐项（H1）或拒绝执行（H2）。 |
| H-A4 | 事件进 Wails：`agent:choice`、`agent:confirm`。`wailsBridge.sendMessage` 必须订阅这两类，且 `EventsOff` 列表要包含它们。 |
| H-A5 | 工具仍只经 Registry。H1 用模型工具 `ask_user`（内核实现，不经过 MCP）。H2 挂在现有 `OnBeforeAct` 拒绝路径上，不要第二套 Rail。 |
| H-A6 | 禁止 toast 冒充完成。卡片必须可点、结果必须写进会话历史。 |
| H-A7 | 一次只做一个 WP。H1 可先于 H2。 |

暂停实现建议（任选一种，测试能测即可）：

- 引擎内 `pendingHuman chan humanReply`，`App` 持有 `map[sessionID]chan`；或  
- `context` + 引擎方法 `DeliverHumanReply(sessionID, payload)`。  

**禁止**用轮询文件、禁止新开 HTTP 服务。

超时：默认 **5 分钟**；超时走 H-A3。超时时间可做常量，不要做成无超时挂死。

---

## WP-H1　对话中选择题（Options Card）

### 用户故事

Agent 遇到分叉（实现方案、窗口行为、测还是先提交等）时，对话流出现卡片：单选、推荐标记、可选补充输入、「跳过并采用推荐」「确定提交」。点完后同一轮继续，用户消息里能看到「已选择：…」。

### 协议

内核向模型暴露工具（名称固定）：

```text
ask_user
  question: string     // 必填，展示标题
  options: array       // 2～5 项 { id, label, description, recommended? }
  allow_custom: bool   // 是否显示补充输入，默认 true
```

行为：

1. `runTool` 识别 `ask_user`：**不**当普通工具执行完就下一轮 LLM。  
2. 发 `EngineEvent{Type: EventChoice, ...}` → `app_chat` → `runtime.EventsEmit("agent:choice", payload)`。  
3. payload 至少：`session_id`, `request_id`, `question`, `options`, `allow_custom`。  
4. 阻塞直到 `ResumeAgentChoice(sessionID, requestID, optionID, customNote)`。  
5. 把工具结果写回 conversation，例如：  
   `用户选择了 option_id=<id> label=<label>。补充：<note或无>`  
   再继续 `StreamChat`。  
6. 跳过 / 超时：采用 `recommended==true` 的第一项；都没有则采用 `options[0]`，结果里注明 `skipped_default=true`。

非法：`options` 少于 2 或多于 5、缺 question → 工具错误返回，不暂停。

### 前端

- `ChatCockpit`：当前 assistant 消息下渲染卡片（暖色：选中项 `border-[#D96B27]`）。  
- 流式中 `isStreaming` 保持 true（表示本轮未结束），发送按钮保持中断态可用。  
- 确定 → 只调 `ResumeAgentChoice`，**禁止**再调 `SendMessage`。  
- Esc：若 choice 卡片打开，视为跳过采用推荐（写入历史），不要直接 `CancelAgentStream`（那是中断整轮）。  
- 与 WP-10 对齐：Esc 优先关 choice 卡片，再关 mention / 子弹窗 / 设置。

### 测试（先红）

`internal/core/loop`：

- `TestEngine_AskUserPausesUntilResume`：mock 第一轮 tool `ask_user`，断言在 Resume 前 **没有** 第二次 `StreamChat`；Resume 后有第二轮且 tool 结果含所选 id。  
- `TestEngine_AskUserSkipUsesRecommended`：不 Resume、关 channel 或超时 → 结果含推荐项。  
- `TestAskUserRejectsFewerThanTwoOptions`：不暂停。

前端至少有一处手工验收（合同不强制 Vitest）：卡片出现、点确定后流式继续。

### 文件（预计）

- `internal/core/loop/engine.go` 事件类型 `EventChoice`  
- `internal/core/loop/llm_path.go` / `tools.go`  
- `app_chat.go`：`ResumeAgentChoice`、等待表  
- `frontend/src/core/wailsBridge.ts`、`stores/workbench.ts`、`ChatCockpit.vue`

### 完成定义

- [ ] 无前端 Resume、仅 SendMessage 新句子，会取消旧任务（现行为）——H1 完成后天 **choice 等待中发新句** 的策略：**取消等待并视为整轮中断**（fail-closed），toast 说明。不要两轮并行。  
- [ ] 上述 Go 测试绿。  
- [ ] `go test ./internal/core/loop`；`npm.cmd run build`。

---

## WP-H2　危险工具「允许这一次」

### 用户故事

Agent 调用将被 SafetyRail 判定危险的 `exec_command`（或同等）时，对话出现琥珀色卡片：命令摘要、风险说明、**拒绝并让模型改方案（默认）**、**允许这一次**。不允许永久白名单。

### 与现有 Rail 的关系

- **不要**把所有 Rail 拒绝都变成询问。仅当：工具为 `exec_command`（及未来显式标 `NeedsConfirm` 的算子），且 Rail 因 **dangerous 正则** 拒绝（Reason 含 `dangerous command` 或专用码）。  
- 路径穿越、analyze 写盘、策略拦截：**仍直接拒绝**，不弹确认（否则只读策略被点穿）。  
- 用户允许后：本 `request_id` 执行 **一次**；下一次同类命令再问。用 `sync.Map`/`session+toolCallID` 一次性 ticket，禁止 session 级永久放行。

建议：`RailDecision` 增加 `NeedsConfirm bool`（可选）。SafetyRail 危险命令：`Allow=false, NeedsConfirm=true`。`runTool`：若 `NeedsConfirm` 且非策略拦 → 发 `EventConfirm` 并暂停。

### 协议

事件 `agent:confirm`：`session_id`, `request_id`, `tool`, `args_preview`（命令字符串，**不得**带密钥；过 StripSecrets）, `reason`。

IPC：`ResumeAgentConfirm(sessionID, requestID, allow bool)`。

- `allow=false` 或超时/Esc：工具结果 `[安全拦截] 用户拒绝：...`，模型继续，**不执行**。  
- `allow=true`：跳过**这一次** Rail 危险检查后 `Execute`（仍要过路径沙箱与 DenyByStrategy）。

### 测试（先红）

- `TestRunTool_DangerousCommandPausesForConfirm`：`rm -rf` 在 confirm 前不执行（可用 mock terminal）。  
- `TestRunTool_ConfirmDenyReturnsIntercept`：拒绝后无副作用。  
- `TestRunTool_AnalyzeWriteStillNoConfirm`：analyze + `fs_control write` **不**发 confirm，直接策略拦截。

### 前端

原型暖色条：拒绝 / 允许这一次。Esc = 拒绝。  
流式保持；确定只调 Resume。

### 完成定义

- [ ] analyze 写盘仍无确认框。  
- [ ] 允许一次后，下一句同样命令仍要再确认。  
- [ ] `go test ./internal/core/loop ./plugins/rail/safety` 绿。

---

## 2. 明确不要做

- 把 H1 做成快捷短语芯片（「运行测试」等）冒充选择题。  
- 危险命令默认允许。  
- 用 `window.confirm`。  
- 为 HITL 引入第二条 `llm.StreamChat` 或 HTTP 回调服务。  
- 扩大询问面到所有工具（读文件、search、git status 不要问）。

---

## 3. 建议顺序

```
WP-H1 引擎暂停 + ask_user + UI
WP-H2 NeedsConfirm + 一次票 + UI
```

H1 的暂停通道应设计成 H2 可复用（同一 `pendingHuman`）。

若与 WP-R1（探活 TLS）并行：R1 不改 loop 暂停；H1/H2 不改 pinger。

---

## 4. 交卷

```
feat(wp-hN): <用户可感知：对话中可选方案 / 危险命令可点允许一次>

- 改了：...
- 红测：...
- 绿证：go test ./internal/core/loop
- 未做：永久白名单 / 追问芯片
```

手工：发一条会让模型调用 `ask_user` 的提示（可在测试里强制 mock）；或临时在 analyze 下让模型「给出两种重构方案并调用 ask_user」。产品上可在 system 提示加一句：**遇到互斥实现方案时必须调用 ask_user，禁止擅自选一个继续改盘。** 该句放 `ApplyStrategy` 末尾，所有策略生效。

---

## 5. 完成总检

- [ ] 选择题：暂停 → 点选 → 同轮继续，会话里有选择记录。  
- [ ] 新 SendMessage 在等待中会中断等待（fail-closed），不双开。  
- [ ] 危险命令卡片：拒绝则不执行；允许一次则执行且不永久放行。  
- [ ] 策略拦截与路径穿越无卡片。  
- [ ] Esc 分层：choice/confirm 优先于设置。  
- [ ] 无 toast 假完成。

# 湉码 / tiancode — 审查整改交接书（Review Remediation Handoff）

> **读者**：接手的 AI 开发代理 + 人类 Reviewer。  
> **性质**：`docs/AI_IMPLEMENTATION_CONTRACT.md`（原型缺口 WP-1…WP-10）的**补充卷**。  
> 只装源码审查发现的**安全 / 质量缺口**。合同铁律与 `AGENTS.md` 冲突时，**以合同与 AGENTS.md 为准**。  
> **不要重做**已完成的 WP-1…WP-10（诚实命名、人搜索、首轮硬闸、Diff 横幅、策略弹窗、SKILL.md、MCP env、TDD 卡片、折叠、Esc）。  
> **栈**：Wails v2 + Go 微内核 + Vue 3。工作包编号：`WP-R1` … `WP-R6`（R = Remediation）。

行号是交接时快照。若与 HEAD 不符，**以语义为准**，并顺手改正文行号。

---

## 0. 给 AI 的铁律

1. **TDD**：每个 WP 先提交当前必失败的测试 → 确认红 → 再改实现 → 绿。未见红灯禁止改实现。  
2. **一次一个 WP**。不夹带重构、不抬 `maxLLMTurns`、不加活动栏图标。  
3. **零 Demo / fail-closed**。禁止假数据、假在线、测试里硬编码 `true`。  
4. **禁止第二套 LLM 回路**（合同 A1）。生产聊天只走 `internal/core/loop` + `prov.StreamChat`。  
5. **交卷前必跑** §7。任一步失败即未完成。  
6. **Windows 发货**：Go 路径可用 `E:\pro\tools\go\bin\go.exe`；前端 `npm.cmd run build`。

---

## 1. 速览（避免按错栈）

- 活路径：`app.go` / `app_*.go` / `internal/**` / `plugins/**` / `frontend/src/**`。  
- 执行：`internal/core/loop`（`executeDirectLLM`）。  
- 闸门：`DenyByStrategy` → `SafetyRail.OnBeforeAct` → `Registry` / `MCPCall`（`internal/core/loop/tools.go`）。  
- 配置：`~/.tiancode`。  
- 禁止当施工依据：`archive/**`、`docs/knowledge/**`、Tauri/Python/React 旧 PRD。  
- **双 module**：根 `go.mod` 与 `backend/go.mod`。根目录 `go test ./...` **覆盖不到 `backend/`**。

```powershell
# 根 module
& E:\pro\tools\go\bin\go.exe test ./...
& E:\pro\tools\go\bin\go.exe run ./tools/archcheck
cd frontend; npm.cmd run build; cd ..
```

---

## 2. 工作包总览

| 包 | 标题 | 等级 | 证据 |
|----|------|------|------|
| WP-R1 | 探活全局关闭 TLS 校验 | **P0 活路径** | 源码已证 |
| WP-R2 | SafetyRail `rm` 变体漏拦 | **P1 先红测** | 正则推演，测试未覆盖 |
| WP-R3 | analyze 按工具名硬编码；MCP 无拦 | **P1** | 源码已证 |
| WP-R4 | `llm.getHTTPClient` 子串 localhost（死路径） | P2 清理 | 仅测试引用 |
| WP-R5 | 非 Windows 明文密钥 + 0644/0755 | P2 跨平台 | 源码已证 |
| WP-R6 | Provider 取 map[0]、死参数、重复函数、CI | P2 整洁 | 源码已证 |

**不要把 WP-R4 写成 P0。** 聊天生产 TLS 在 openai provider 默认 `http.Client` 上，是校验的。

---

## WP-R1　探活全局 `InsecureSkipVerify`　【P0】

### 证据

- `internal/network/pinger.go`：`defaultTransport` 的 `TLSClientConfig.InsecureSkipVerify: true`（全局，**所有 host**）。  
- `PingTarget` 使用该 `defaultClient`。  
- `App.PingChannel`（`app_config.go`）→ `network.PingTarget(ch.Endpoint)`。用户可填任意 HTTPS。  
- UA：`codex_cli_rs/0.101.0 ...`（同文件请求头）。  

**易错点（审查稿曾混在一起）：**

- **TLS 跳过是全局的**，不是「仅本地」。公网 `https://…` 探活同样不验证书。  
- `HasPrefix(lower, "localhost")` 只用在**缺 scheme 时补 `http://` 还是 `https://`**。`localhost.evil.com` 无 scheme 会被补成 `http://localhost.evil.com`，这是**另一处前缀误判**，修 TLS 时一并改成 `url.Parse` 后的 `Hostname()`。  
- `FetchUpstreamModels` 用 `&http.Client{Timeout: 10s}` **默认验 TLS**，不要和 pinger 混为一谈。

### 红测（先写先跑，必须红）

`internal/network/pinger_test.go`：

- `TestPingTarget_UntrustedPublicTLSFails`：对自签/不可信证书的 **非回环 host**（或 `httptest.NewTLSServer` + 改 transport 注入）断言 **必须 error**。当前 skip verify → 应红。  
- `TestPingTarget_LoopbackAllowsHTTP`：`http://127.0.0.1:<port>` 探活成功。  
- `TestClassifyPingHost_RejectsLocalhostPrefixSpoof`：`localhost.evil.com` 不得被当成回环。

可将 client/transport 做成包内可替换变量，便于测，这算本包允许的最小解耦。

### 实现

1. 默认 **strict** transport（校验证书）。  
2. 仅当 hostname 精确为 `localhost` / `127.0.0.1` / `::1` 时允许 insecure 或明文 http。  
3. 用 `url.Parse` + `Hostname()`，禁止 `HasPrefix`/`Contains` 判本地。  
4. **UA**：先保留为命名常量（上游网关可能认这套头）。**是否改成产品 UA 由人类拍板**，AI 不得擅自删导致探活全挂。

### 验收

- 上述测试绿。  
- 手工：不可信公网 HTTPS 探活失败；`http://127.0.0.1:…` 成功。  
- `go test ./internal/network ./internal/config` 绿。

---

## WP-R2　SafetyRail `rm` 变体　【P1 · 先红后绿】

### 证据

`plugins/rail/safety/safety_rail.go` 前两条：

```text
(?i)\brm\s+(-[a-zA-Z]*f[a-zA-Z]*\s+)?-?[rR]f\b
(?i)\brm\s+-rf\b
```

静态推演：`rm -rf /` 命中；`rm -fr /`、`rm -r -f /`、`rm -f -r /`、`rm --recursive --force /` **按语义不命中**。  
`safety_rail_test.go` 只有 `TestSafetyRail_BlocksRmRf`。  
终端非 Windows：`sh -c`（`plugins/tool/terminal/terminal_tool.go`）。发货是 Windows `cmd /c`，变体仍可能从 `exec_command` 进来。

**禁止未跑红测就改正则。**

### 红测（表驱动）

应拦：`rm -rf /`、`rm -fr /`、`rm -r -f /`、`rm -f -r /`、`rm --recursive --force /`。  
不应拦：`rm file.txt`、`go test ./...`、`rm -i x`。

以**跑出来的红/绿**为准，不要假设上表全红。

### 实现（禁止再堆一条特例正则）

`rm` 命中后拆 token：短选项展开字符集（`-fr` → f,r），识别 `--force` / `--recursive`；同时含 force+recursive 则拦。正则可留作兜底。

### 验收

表驱动全绿 + 反例不误伤；`go test ./plugins/rail/safety` 绿。

---

## WP-R3　analyze 按名字硬编码；MCP 无策略闸　【P1】

### 证据

- `ApplyStrategy` analyze 只从可见表去掉 `exec_command` 与 **`write_file`**。  
- 生产工具名：`fs_control`、`git_control`、`search_workspace`、`exec_command`。 **没有注册 `write_file`**（幽灵名）。  
- `DenyByStrategy` analyze：拦 `exec_command`、`write_file`、**仅** `fs_control` 且 `action=="write"`；turn==1 非清单 `read`。  
- **`fs_control` 无 delete**（只有 read/write/list）。不要再写「delete 是否被拦未验证」——动作不存在。  
- **`git_control.Execute` 无视参数，只返回 status**。描述写「查询并管理」是名实不符；analyze 下当前**调了也写不了 Git**。  
- `runTool`：Deny → Rail → Registry，否则 `MCPCall`。**Deny 不看 MCP 名/能力。** 用户若挂带写的 MCP，analyze 可被穿透。未挂 MCP 时是潜在债。

### 红测

- `TestApplyStrategy_AnalyzeHidesMutatingTools`：可见表含 `fs_control`/`git_control`/`search_workspace`；analyze 后不得把「会改工作区」的算子留给模型。当前应红（`fs_control` 仍在表里）。  
- `TestDenyByStrategy_AnalyzeBlocksFsWrite`：回归，现应绿。  
- MCP：无能力声明时 analyze **fail-closed**（见人类拍板）。未拍板前先写测试标明 Skip 或只测内置工具。

### 实现

1. `v1.ToolDefinition` 或 `ToolPlugin` 增加能力：`Mutating bool`（最小字段，勿过度设计）。  
2. `fs_control` write → Mutating；read/list → 否（或按 action 在 Deny 里判，表上可拆成只读描述）。  
3. `git_control` 现状 Mutating=false，**改 Description 为只读 status**。将来加写再改 true 并进 Deny。  
4. `ApplyStrategy`/`DenyByStrategy` 按 Mutating，删幽灵 `write_file` 或映射到 `fs_control` write。  
5. MCP：缺能力声明时 analyze **拦截**（fail-closed）。**若人类否决，本条不做，在 PR 写未做。**  
6. **不得破坏** `strategy_test.go` 首轮地图契约。

### 验收

新测试绿；既有 turn-1 测试绿；analyze 可见表无写盘内置算子。

---

## WP-R4　死路径 `getHTTPClient`　【P2】

### 证据

`internal/llm/client.go`：`Contains("localhost")` 会把 `notlocalhost.evil.com` 当本地 insecure。  
`llm.StreamChat` 仅 `client_test.go` / `client_real_test.go`。生产 `llm_path.go` 用 `prov.StreamChat`。  
`DefaultWorkspaceTools` 在 `client.go`，测试引用。

### 处置

确认无生产引用后删除 `StreamChat` / `getHTTPClient` 死分支 / 第二套工具词表，并改测试。  
**不要修死路径的 TLS 还留两套 client。**

### 验收

`go test ./internal/llm ./internal/core/loop` 绿；生产包不再依赖 `llm.StreamChat`。

---

## WP-R5　非 Windows 密钥　【P2】

### 证据

- `secret.go`：非 Windows `plain:` + 明文。Windows DPAPI。  
- `channel_store.go`：`MkdirAll(..., 0755)`。  
- `extra_stores.go` `atomicWriteConfig`：`0644`。  
当前发货仅 Windows → **P2**。AI 不得把发货范围默默扩到 mac/Linux。

### 处置（仅当人类批准排期）

目录 `0700`、密钥文件 `0600`；非 Windows 单测 `os.Stat` 断言。Keyring 可后续，本包不要上大依赖。

---

## WP-R6　Provider 与整洁　【P2】

| 子项 | 证据 | 处置 |
|------|------|------|
| `provs[0]` | `GetProviders` 遍历 **map**，顺序不定 | 有 `EngineRequest.Provider` 则按 ID；否则按 ID 排序后取首 |
| `toolMap` | `llm_path.go` 调用 `runTool(..., nil, ...)` | 删参数与死分支，只走 registry |
| `trimToolOutput` | `tools.go` 与 `app.go` 各一份 | 留 loop 包一处，app 复用或删除 app 内未调用副本 |
| `maxSteps` | 只赋值不读 | 删除；**禁止**借机加大轮次 |
| CI | `.github/workflows/ci.yml` 仅 ubuntu：`go test ./...` + archcheck + frontend build | 加 `go vet ./...`；可选 Windows `go build -tags desktop,production`；`backend/` 单独 `go test` 或文档写明不测 |

一次 PR 不要把 CI、删死代码、Provider 全揉在一起。优先：删 `toolMap` 死参 + Provider 排序；CI 另 PR。

---

## 3. 每个 WP 交卷清单

- [ ] 新测试先红后绿，红灯可复述。  
- [ ] `go test ./...`；`go run ./tools/archcheck`。  
- [ ] 动到前端才跑 `npm.cmd run build`。  
- [ ] 未开第二套 LLM、未 Demo、未改 `archive/**`。  
- [ ] 未把半成品标 🟢。

## 4. Commit 模板

```
fix(wp-rN): <安全/质量可感知的一句话>

- 改了：<文件>
- 红测：<测试名 + 修前失败>
- 绿证：go test <包>
- 未做：<未覆盖项>
```

禁止「安全已全部加固」。

## 5. 人类拍板（AI 不得擅自）

1. WP-R1：探活 UA `codex_cli_rs` 是否必须保留？默认 **保留常量**。  
2. WP-R3：MCP 无能力声明时 analyze 是否 fail-closed？默认 **是**；否决则跳过 MCP 条。  
3. WP-R5：仅 Windows 发货时是否延后？默认 **延后**。  
4. WP-R4：是否授权删除 `internal/llm.StreamChat` 旧层？默认 **授权删除（改测试）**。

## 6. 建议顺序

```
WP-R1（TLS 探活）→ WP-R2（rm 红测+语义拦）→ WP-R3（能力元数据）
然后 WP-R4（删死层）→ WP-R6 小清理 → WP-R5 仅在拍板后
```

---

## 7. 证据边界

- `file:line` 为快照。  
- WP-R2 **必须**用测试证实漏拦，禁止只信正则推演。  
- 接手第一件事：在本机跑现有 `go test ./plugins/rail/safety ./internal/network ./internal/core/loop`，再开 WP-R1 红测。

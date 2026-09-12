# 知识点 48: 控件错名治理、防脱缰智能熔断器与 TDD 结构化失败提取

## ① 知识点与问题背景 (Context & Problem Statement)

在 Agent 桌面工作台演进过程中，随着外壳冗余图标的收敛，暴露了深层次的工程信任与产品诚实性问题：
1. **错名与预期膨胀 (Inflated Terminology)**：
   - 侧边栏原称为「微内核快照」，但底层逻辑实际为 `git stash` 管道。将 Stash 命名为「快照 / Snapshot」严重误导开发者，让其预期拥有时光机或按步点回溯能力，且 stash pop 会搅乱当前工作区未提交的文件；
   - 顶栏原将单一 Monaco 编辑器与简单文件树命名为「代码工作区」，给用户带来完整 IDE 的虚假预期；
   - 顶栏快速跳转原提示「快速检索分支、文件与算子...」，实际为会话、文件与面板的轻量跳转；
   - 顶栏模型状态原为硬编码绿灯「就绪」，不论上游网关是否配置、网络测速是否离线均常亮绿灯；
   - 渠道卡片原附带 `auth_type: "bearer_token"` 等无实质实现的虚假标签；
   - 策略药丸原采用单次点击盲切循环（`analyze` -> `tdd` -> `implement`），用户极易误触切换至 `implement`（完全放行读写），带来安全风险。
2. **加时不加智与死循环爆 Token (Runaway Loop & Token Exhaustion)**：
   - 上一轮为了给模型自主判断空间，将轮次上限提高至 100 轮。但在实际编码任务中，若模型陷入连续编译报错、或连续发起完全相同的工具调用，100 轮循环将无节制消耗成千上万 Token 并冲破 32k 上下文窗口造成严重截断；
3. **TDD 失败缺乏结构化用例定位**：
   - 原 TDD 验证失败时仅给出模糊报错或截断输出，开发者无法一眼得知具体是哪个单测用例挂了。

---

## ② 核心原理与知识内容 (Knowledge Content & Root Cause)

1. **命名诚实性原则 (The Principle of Honest Naming)**：
   - 工具名必须如实反映其底层机制：`git stash` 就是 `Git 暂存储藏 (Stash)`，操作就是 `＋ 储藏 (Stash)` / `恢复 (Pop)` / `无暂存记录`；
   - 布局就是 `文件与编辑器`，搜索就是 `快捷跳转文件、会话或面板...`；
   - 状态灯必须与真实网络探活结果强绑定：
     - 未配置渠道：灰色（`bg-zinc-400`，未配置渠道）
     - 待测速 / standby：琥珀色（`bg-amber-500`，待测速）
     - 离线 / 失败：红色（`bg-red-500`，离线/异常）
     - 在线 / 有延迟：绿色（`bg-[#10A37F]`，在线 · 42ms）
     - 流式推理中：陶土橙脉冲（`bg-[#D96B27]`，推理中）。
2. **防脱缰双重主动熔断器机制 (Dual Runaway Circuit Breakers)**：
   - 将兜底防爆硬上限恢复至合理的 20 轮。
   - **重复调用主动熔断 (Duplicate Tool Call Circuit Breaker)**：追踪 `lastToolSig` 与 `consecutiveIdenticalCalls`。当模型连续发起 3 次参数完全一致的工具调用时，判定陷入病态自旋死循环，立即拦截执行、切断循环，并引导模型基于已有信息作答；
   - **连续错误主动熔断 (Consecutive Error Circuit Breaker)**：追踪 `consecutiveErrors`。当工具连续报错 3 次时，判定当前执行路径存在环境依赖缺失或不可行假设，立即中止尝试，避免继续空耗上下文。
3. **双栈 TDD 失败用例精准解析器 (Structured Test Failure Parser)**：
   - Go 栈：解析 `--- FAIL: <TestName>` 与 `FAIL:\t<TestName>`；
   - Node / Vitest / Jest 栈：解析 `FAIL <file>`、`✕ <test>`、`✖ <test>`、`● <test>`；
   - 去重提取为 `FailedTests []string`，前置渲染高亮清单：`❌ 【失败测试用例清单】(N 个)`。

---

## ③ 标准解决方案与实操步骤 (Actionable Solutions & Step-by-Step Guide)

### 1. 结构化用例提取实现 (`internal/agent/swarm.go`)
```go
func ExtractFailedTests(goOut, npmOut string) []string {
    failed := make([]string, 0)
    seen := make(map[string]bool)

    addTest := func(name string) {
        name = strings.TrimSpace(name)
        if name != "" && !seen[name] {
            seen[name] = true
            failed = append(failed, name)
        }
    }

    if goOut != "" {
        for _, line := range strings.Split(goOut, "\n") {
            trimmed := strings.TrimSpace(line)
            if strings.HasPrefix(trimmed, "--- FAIL:") {
                parts := strings.Fields(trimmed)
                if len(parts) >= 3 { addTest(parts[2]) }
            } else if strings.HasPrefix(trimmed, "FAIL:\t") || strings.HasPrefix(trimmed, "FAIL: ") {
                parts := strings.Fields(trimmed)
                if len(parts) >= 2 { addTest(parts[1]) }
            }
        }
    }
    // npm/vitest 同样解析
    return failed
}
```

### 2. 双重智能熔断器 (`internal/core/loop/llm_path.go`)
```go
sig := fmt.Sprintf("%s:%s", tc.Function.Name, strings.TrimSpace(tc.Function.Arguments))
if sig == lastToolSig {
    consecutiveIdenticalCalls++
} else {
    consecutiveIdenticalCalls = 1
    lastToolSig = sig
}

if consecutiveIdenticalCalls >= 3 {
    circuitBroken = true
    circuitBreakReason = fmt.Sprintf("⚠️ [防死循环熔断] 工具 [%s] 连续发起 3 次完全相同的调用，已触发防脱缰保护。", tc.Function.Name)
    break
}
```

### 3. 前端原生安全选择器 (`ChatCockpit.vue`)
```html
<select
  v-model="s.executionStrategy"
  class="text-xs font-semibold px-2.5 py-1 rounded-full border cursor-pointer outline-none transition-colors appearance-none pr-5.5 shadow-2xs"
>
  <option value="analyze">🛡️ 只读审查 (拦截写盘)</option>
  <option value="tdd">🧪 TDD 闭环 (强制跑测)</option>
  <option value="implement">⚡ 直接改代码 (全读写)</option>
</select>
```

---

## ④ 避坑指南与最佳实践 (Troubleshooting & Best Practices)

1. **绝对禁止自欺欺人的 UI 铭牌**：
   - 给基础功能套高大上名词（如把 `git stash` 叫微内核快照，把文件树叫 IDE 工作区）不仅无法提升产品力，反而会极大损害开发者群体的信任感；
2. **AI 自主结束必须与物理安全熔断双轨并行**：
   - 赋予 AI 自主判断结束权（0 工具调用即交付完成）不等于取消防爆阀门；20 轮硬上限 + 重复调用熔断 + 连续错误熔断是保证商业化模型调用的经济安全底线；
3. **TDD 必须直接暴露失败用例名**：
   - 不要让开发者在数百行测试日志中人工 grep 失败用例，结构化解析并置顶是提升 TDD 自愈效率的核心。

---
description: 项目级最高前置规则：每次执行任务前必须优先通读并对齐 UI/UX、React Web 与 Rust 三大专业技能规范
always_on: true
---

# 项目级强制前置规则：三大发货专业技能优先对齐法则 (Mandatory Prior Skill Consultation)

在 湉码 / tiancode 项目中，现行唯一发货技术栈为 **Wails v2 + Go 微内核 + Vue 3 / Vite**。任何涉及功能分析、需求设计、交互设计、前端组件编写、微内核业务引擎与系统调用的任务，**每次都必须在开始行动前优先通读并严格对齐以下三大发货专业技能（Skills）规约**：

---

## 📋 每次开发前必查清单 (Mandatory Pre-Flight Check)

### 1. 🎨 `ui-ux` (UI/UX 体验与人机工程设计规约)
- **技能路径**: `.agents/skills/ui-ux/SKILL.md`
- **必查要点**:
  - [ ] 是否遵循 Warm Minimalist 暖色极简调色盘：App Base `#FAF8F5`、Surface `#F4EFEA`、主强调色 `#D96B27`、代码底色 `#1E1C1A`？
  - [ ] 是否符合 16:9 宽屏人机工程学比例？
  - [ ] 弹窗是否严禁使用浏览器原生 `alert/confirm/prompt`，并实现屏幕严格居中、Esc 退出、显式 `[X]` 关闭？
  - [ ] 所有纯图标操作是否配置了鼠标悬停 Tooltip / title 提示？
  - [ ] 对话流中思考过程、工具调用卡片与最终回复是否层级清晰、时序紧密对齐？
  - [ ] 项目宪法是否在对话顶栏常驻展示，改动文件是否带出 Diff 供人工确认？

---

### 2. 🏛️ `tcode-studio-architect` (Wails v2 + Go 微内核架构规约)
- **技能路径**: `.agents/skills/tcode-studio-architect/SKILL.md`
- **必查要点**:
  - [ ] 是否遵循单一执行内核（Direct LLM + Registry Tools，禁止在 `app.go` 另起第二套循环）？
  - [ ] 工具调用唯一路径：必须通过 `registry.GetTool(name).Execute(ctx, args)`，严禁直接在业务层持有 Tool 字段？
  - [ ] Rail 拦截链必须生效：工具执行前必须调用 `rail.OnBeforeAct()`，执行后调用 `rail.OnAfterAct()`？
  - [ ] 任务生命周期完整收敛：触顶、取消、上游 4xx/5xx 与空输出必须人话收尾，绝无「突然没了」？
  - [ ] 「继续」指令必须精准接续 Session 挂载的 `TaskModel`，禁止推翻重勘？

---

### 3. 🧪 `sdd-tdd-workflow` (SDD 规范驱动与 TDD 测试驱动开发规约)
- **技能路径**: `.agents/skills/sdd-tdd-workflow/SKILL.md`
- **必查要点**:
  - [ ] 必须先完成 Spec 接口与数据契约设计，严禁未定义直接编码？
  - [ ] 必须严格遵循 **红 (Red: 编写前置测试验证失败) ➔ 绿 (Green: 最小实现全绿) ➔ 重构 (Refactor: 双向钢人审查)** 三步节拍？
  - [ ] TDD 策略下测试未通过时，任务状态绝对不是完成，阻断虚假宣称完成？
  - [ ] 验证闭环：修改 Go 代码必须执行 `go test ./...`，架构变动执行 `archcheck`，前端改动执行 `npm run build`？

---

## 🚨 历史技术栈归档说明
历史基于 `Tauri v2 / Rust Core` 与 `React 19 / Python Daemon` 的技能与原型材料已全部归档至 `archive/`。**严禁以历史栈作为主路径施工依据**。

## 🚨 违规阻断令
**严禁绕过上述三大发货技能规范直接编写代码或提交变更！若发现任何违背上述三大技能准则的设计或实现，一律打回重构。**

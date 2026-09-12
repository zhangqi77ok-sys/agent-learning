# 历史原型阶段设计沉淀记录 (Historical Prototype Notes)

> ⚠️ **注意**：本文件为项目早期在 `archive/web_prototype.html` / `archive/prototype/` 探索阶段（WP 16~26）的原型交互与功能设计记录，**仅供历史追溯与视觉参考**。
> 湉码现行发货代码已全面升级收敛至 **Wails v2 原生桌面端 + Go 微内核 + Vue 3** 架构，请勿将本文档中的 React `.tsx`、Python 脚本或旧 Store 路径作为施工基准。

---

### 16. 高保真原型架构对齐记录
* **顶层沉浸式标题栏 (`Titlebar.tsx`)**：
  * 高度严格锁定 38px，左侧呈现 `T` Logo、项目与分支徽标；
  * 中间集成单焦点工作区胶囊（`💬 智能对话` / `◫ 双栏协同` / `📝 代码工作区`）；
  * 右侧快捷集成终端抽屉开关、全局设置弹窗与原生窗口控制。
* **左侧极简活动栏与次级多功能抽屉 (`ActivityBar.tsx` & `LeftSidebar.tsx`)**：
  * 活动栏挂载功能图标；
  * 次级抽屉实现平滑切换：会话分支抽屉、工程文件目录树、Git 源代码管理抽屉。
* **智能对话工作台与代码 Diff 联动 (`ChatCockpit.tsx` & `CodeWorkspace.tsx`)**：
  * 顶部多会话 Tab 切换与代码区展开/收起按钮；
  * 消息流呈现思考折叠卡片、算子调用抽屉日志；
  * 改动文件卡片组与 Monaco 行级双栏对比。

### 17. 原型期流式推理与 ReAct 自主算子执行
* 动态多模型网关与 WAF 穿透路由验证；
* 原生深度心智思考流实时推送 (`Thinking Stream`)；
* ReAct 自主物理算子调度闭环 (`run_command`, `read_file`, `write_file`, `git_status`)。

### 18. 原型期 Git 控制模态窗与效能监控大盘
* 分支管理与即时检出模态窗 (`GitBranchModal.tsx`)；
* 影子快照与 Stash 储藏中心 (`SnapshotModal.tsx`)；
* Canvas 吞吐时序走势 (`ThroughputCanvas.tsx`)。

### 19. 原型期 MCP 工具协议服务导入
* MCP 工具协议服务导入模态窗 (`MCPImportModal.tsx`)：JSON 一键粘贴、手动表单配置与精选服务。

### 20. 原型期 Windows 打包脚本探索
* 早期基于 Python 的打包流水线实验 (`build_installer.py`)，现已全面替换为 Go 嵌入式安装程序与 PowerShell 构建脚本 (`scripts/build-windows.ps1`)。

### 21. 原型期大模型 ReAct 物理编程实战验证
* 早期首字延迟 (TTFT) 极速响应测试与算子闭环验证。

### 22. 原型期 Vitest 单元测试矩阵
* 原型验证期基于 Vitest + JSDOM 构建测试环境，涵盖工作区模式切换、分支列表过滤、检出切换、以及默认空 Key Fail-Closed 安全断言等规范。

### 23. 原型期系统原生文件夹选择
* Windows 原生对话框集成探索（现已由 Go Wails 原生对话框承接）。

### 24. 原型期会话分支分叉与时光倒流
* 对话分支分叉 (Session Forking) 与时光倒流回退 (Time-Travel Revert) 设计。

### 25. 原型期工程知识库 RAG 检索增强
* 四段论沉淀动态解析与星系研读视窗设计原型。

### 26. 原型期源码 AST 依赖拓扑探索
* 源码 AST 关系提取与可视化双向依赖探索网络。

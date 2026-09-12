<template>
  <div class="h-full w-full bg-[#FAF8F5] text-[#18181B] flex flex-col font-sans select-none overflow-hidden antialiased">
    <header style="--wails-draggable:drag" class="h-[38px] min-h-[38px] bg-[#FAF8F5] border-b border-black/[0.08] flex items-center justify-between px-3 z-30 select-none">
      <div style="--wails-draggable:no-drag" class="flex items-center gap-2">
        <div class="w-5 h-5 rounded-md bg-[#18181B] text-white flex items-center justify-center font-bold text-xs shadow-xs">湉</div>
        <span class="text-xs font-semibold tracking-tight text-[#18181B]">湉码</span>
        <span class="text-[#A1A1AA] text-xs">/</span>
        <button
          @click="s.chooseWorkspace"
          style="--wails-draggable:no-drag"
          class="flex items-center gap-1.5 px-2 py-0.5 rounded-md hover:bg-black/[0.05] transition-all cursor-pointer group text-xs font-medium text-[#27272A]"
          title="点击切换工作区文件夹 (系统原生文件夹对话框)"
        >
          <span class="text-[#D96B27]">📁</span>
          <span class="group-hover:text-[#D96B27] max-w-[160px] truncate">{{ s.workspaceName }}</span>
          <span class="text-[10px] text-[#71717A] bg-black/[0.04] px-1.5 py-0.2 rounded-full font-mono">{{ s.gitBranchLabel }}</span>
        </button>
        <div class="h-3 w-[1px] bg-black/[0.08] mx-1"></div>
        <div class="flex items-center gap-1.5 text-[11px] text-[#10A37F] font-medium bg-[#10A37F]/10 px-2 py-0.5 rounded-full">
          <span class="w-1.5 h-1.5 rounded-full bg-[#10A37F]" :class="{ 'animate-pulse': s.isStreaming }"></span>
          <span>{{ s.selectedModel }} · {{ s.isStreaming ? '推理中' : '就绪' }}</span>
        </div>
      </div>

      <button
        style="--wails-draggable:no-drag"
        @click="s.openCommandPalette()"
        class="hidden md:flex items-center gap-1.5 px-3 py-1 rounded-xl bg-black/[0.04] hover:bg-black/[0.07] border border-black/[0.06] text-xs text-[#71717A] cursor-pointer shadow-2xs"
        title="全局快速命令 (Ctrl+K)"
      >
        <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
        <span class="text-[11px] font-medium">快速检索分支、文件与算子...</span>
        <kbd class="text-[10px] font-mono px-1.5 py-0.2 rounded bg-white text-[#71717A] border border-black/[0.08]">Ctrl+K</kbd>
      </button>

      <div style="--wails-draggable:no-drag" class="flex items-center p-0.5 bg-black/[0.05] rounded-xl text-xs font-medium">
        <button @click="s.setWorkspaceView('chat')" :class="['px-2.5 py-1 rounded-lg flex items-center gap-1.5 cursor-pointer', s.workspaceView === 'chat' ? 'bg-white text-[#D96B27] shadow-2xs font-semibold' : 'text-[#71717A]']">智能对话</button>
        <button @click="s.setWorkspaceView('split')" :class="['px-2.5 py-1 rounded-lg flex items-center gap-1.5 cursor-pointer', s.workspaceView === 'split' ? 'bg-white text-[#D96B27] shadow-2xs font-semibold' : 'text-[#71717A]']">
          <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><line x1="12" y1="3" x2="12" y2="21"/></svg>
          双栏协同
        </button>
        <button @click="s.setWorkspaceView('editor')" :class="['px-2.5 py-1 rounded-lg flex items-center gap-1.5 cursor-pointer', s.workspaceView === 'editor' ? 'bg-white text-[#D96B27] shadow-2xs font-semibold' : 'text-[#71717A]']">代码工作区</button>
      </div>

      <div style="--wails-draggable:no-drag" class="flex items-center gap-2">
        <button
          @click="s.toggleTerminalDrawer()"
          class="flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium text-[#52525B] bg-white border border-black/[0.08] shadow-2xs hover:bg-black/[0.03] cursor-pointer"
        >
          <span class="font-mono font-bold text-[#18181B]">$_</span><span>终端抽屉</span>
        </button>
        <button
          @click="s.isSettingsOpen = true"
          class="flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium text-[#18181B] bg-white border border-black/[0.08] shadow-2xs hover:bg-black/[0.03] cursor-pointer"
        >
          <svg class="w-3.5 h-3.5 text-[#D96B27]" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0-2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
          <span>模型与设置</span>
        </button>
        <div class="h-3 w-[1px] bg-black/[0.08]"></div>
        <div class="flex items-center gap-1">
          <button @click="wailsBridge.windowMinimise()" class="w-6 h-6 rounded flex items-center justify-center hover:bg-black/[0.05] text-[#71717A] cursor-pointer" title="最小化">
            <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="5" y1="12" x2="19" y2="12"/></svg>
          </button>
          <button @click="wailsBridge.windowToggleMaximise()" class="w-6 h-6 rounded flex items-center justify-center hover:bg-black/[0.05] text-[#71717A] cursor-pointer" title="最大化/还原">
            <svg class="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/></svg>
          </button>
          <button @click="wailsBridge.windowClose()" class="w-6 h-6 rounded flex items-center justify-center hover:bg-red-500 hover:text-white text-[#71717A] cursor-pointer" title="关闭">
            <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>
      </div>
    </header>
    <div class="flex-1 flex overflow-hidden relative">
      <ActivityBar />
      <LeftDrawer />
      <div class="flex-1 flex flex-col overflow-hidden relative">
        <div class="flex-1 flex overflow-hidden relative">
          <ChatCockpit v-show="s.workspaceView !== 'editor'" />
          <DiffWorkspace />
        </div>
        <TerminalDrawer />
      </div>
    </div>
    <div
      v-if="s.isSettingsOpen"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/45 backdrop-blur-xs animate-in fade-in duration-150 font-sans"
    >
      <div class="w-[90vw] max-w-[1050px] h-[82vh] bg-white rounded-2xl shadow-2xl border border-black/[0.1] flex flex-col overflow-hidden relative">
        <header class="h-12 bg-[#FAF8F5] border-b border-black/[0.08] flex items-center justify-between px-5 select-none shrink-0">
          <div class="flex items-center gap-2">
            <span class="text-base">⚙️</span>
            <span class="font-bold text-sm text-[#18181B]">系统设置中枢 (Settings Hub)</span>
          </div>
          <button @click="s.isSettingsOpen = false" class="p-1.5 rounded-lg text-[#71717A] hover:bg-black/[0.05] cursor-pointer">✕</button>
        </header>

        <div class="flex-1 flex overflow-hidden">
          <!-- 左侧菜单 -->
          <aside class="w-48 bg-[#F4EFEA] border-r border-black/[0.08] p-3 space-y-1 select-none shrink-0">
            <button
              v-for="m in [
                { id: 'models', label: '🌐 模型与网关渠道' },
                { id: 'mcp', label: '🧩 MCP 服务协议' },
                { id: 'skills', label: '🛠️ Agent 技能库' },
                { id: 'rules', label: '📜 软件规则与提示词' },
                { id: 'theme', label: '🎨 外观与工作区' },
                { id: 'sandbox', label: '🛡️ 安全沙箱与防线' },
                { id: 'about', label: 'ℹ️ 关于系统' }
              ]"
              :key="m.id"
              @click="s.activeSettingsTab = m.id"
              :class="[
                'w-full text-left px-3 py-2 rounded-xl text-xs transition-all cursor-pointer',
                s.activeSettingsTab === m.id ? 'font-bold bg-white text-[#D96B27] shadow-xs' : 'text-[#71717A] hover:bg-black/[0.04]'
              ]"
            >
              {{ m.label }}
            </button>
          </aside>

          <!-- 右侧选项卡主体 -->
          <main class="flex-1 p-5 overflow-y-auto bg-white space-y-4">
            <!-- 选项卡 1: 模型与网关 -->
            <div v-if="s.activeSettingsTab === 'models'" class="space-y-4">
              <div class="flex items-center justify-between">
                <div>
                  <h3 class="text-xs font-bold text-[#18181B]">活跃网关与模型渠道</h3>
                  <p class="text-[11px] text-[#71717A]">已读写 ~/.tiancode/channels.json · 支持实时在线探活测速</p>
                </div>
                <div class="flex items-center gap-2">
                  <button @click="s.pingAllChannels" class="px-3 py-1.5 rounded-xl border border-black/[0.1] text-xs font-medium hover:bg-black/[0.02] cursor-pointer">
                    ⚡ 探测全部通道
                  </button>
                  <button @click="s.openAddChannelModal" class="px-3 py-1.5 rounded-xl bg-[#D96B27] text-white text-xs font-bold shadow-xs hover:bg-[#B8551B] cursor-pointer">
                    ➕ 新增渠道
                  </button>
                </div>
              </div>

              <!-- 渠道卡片列表 -->
              <div class="space-y-2">
                <div v-if="s.channels.length === 0" class="p-8 text-center bg-[#FAF8F5] rounded-xl border border-black/[0.06] text-[#71717A] text-xs">
                  <span class="text-2xl block mb-2">🌐</span>
                  <span class="font-bold text-[#18181B] block mb-1">当前未配置任何模型渠道</span>
                  <p class="text-[11px] text-[#A1A1AA] mb-3">支持配置 AgentRouter、OpenAI、Claude、DeepSeek 等兼容端点</p>
                  <button @click="s.openAddChannelModal" class="px-3 py-1.5 rounded-lg bg-[#D96B27] text-white text-xs font-semibold shadow-xs hover:bg-[#B8551B] cursor-pointer">➕ 新增渠道</button>
                </div>
                <div
                  v-for="ch in s.channels"
                  :key="ch.id"
                  class="p-3 rounded-xl border border-black/[0.08] bg-[#FAF8F5] flex items-center justify-between shadow-2xs"
                >
                  <div class="flex items-center gap-3">
                    <input type="radio" :checked="ch.primary" @change="s.setPrimaryChannel(ch.id)" class="text-[#D96B27] focus:ring-[#D96B27] cursor-pointer">
                    <div>
                      <div class="flex items-center gap-2">
                        <span class="text-xs font-bold text-[#18181B]">{{ ch.name }}</span>
                        <span class="text-[9px] bg-emerald-50 text-emerald-700 px-1.5 py-0.2 rounded font-mono font-bold">{{ ch.status }}</span>
                        <span class="text-[9px] bg-black/[0.04] text-[#52525B] px-1.5 py-0.2 rounded font-mono">{{ ch.auth_type }}</span>
                      </div>
                      <div class="text-[11px] text-[#71717A] mt-0.5 font-mono">
                        {{ ch.endpoint }} · 延迟: <strong :class="s.pingLoadingMap[ch.id] ? 'text-amber-500 animate-pulse' : 'text-[#10A37F]'">{{ s.pingLoadingMap[ch.id] ? '测速中...' : ch.latency }}</strong>
                      </div>
                    </div>
                  </div>

                  <div class="flex items-center gap-2">
                    <button @click="s.executePing(ch.id)" class="px-2.5 py-1 rounded-lg bg-white border border-black/[0.08] text-xs font-medium hover:bg-black/[0.02] cursor-pointer">⚡ 测速</button>
                    <button @click="s.editChannel(ch)" class="px-2.5 py-1 rounded-lg bg-white border border-black/[0.08] text-xs font-medium hover:bg-black/[0.02] cursor-pointer">✏️ 配置</button>
                    <button @click="s.deleteChannel(ch.id)" class="px-2.5 py-1 rounded-lg bg-white border border-red-200 text-xs font-medium text-red-600 hover:bg-red-50 cursor-pointer">🗑️</button>
                  </div>
                </div>
              </div>
            </div>

            <!-- 选项卡 2: MCP 服务 -->
            <div v-else-if="s.activeSettingsTab === 'mcp'" class="space-y-4">
              <div class="flex items-center justify-between">
                <div>
                  <h3 class="text-xs font-bold text-[#18181B]">Model Context Protocol (MCP) 本地服务</h3>
                  <p class="text-[11px] text-[#71717A]">已读写 ~/.tiancode/mcp_servers.json</p>
                </div>
                <button @click="s.isMcpModalOpen = true" class="px-3 py-1.5 rounded-xl bg-[#D96B27] text-white text-xs font-bold shadow-xs hover:bg-[#B8551B] cursor-pointer">
                  ➕ 导入 MCP 服务
                </button>
              </div>

              <div class="space-y-2">
                <div v-if="s.mcps.length === 0" class="p-8 text-center bg-[#FAF8F5] rounded-xl border border-black/[0.06] text-[#71717A] text-xs">
                  <span class="text-2xl block mb-2">🧩</span>
                  <span class="font-bold text-[#18181B] block mb-1">当前暂无挂载的 MCP 本地服务</span>
                  <p class="text-[11px] text-[#A1A1AA]">可导入并管理基于 Model Context Protocol 的工具算子服务</p>
                </div>
                <div v-for="mcp in s.mcps" :key="mcp.id" class="p-3 rounded-xl border border-black/[0.08] bg-[#FAF8F5] flex items-center justify-between shadow-2xs">
                  <div>
                    <div class="flex items-center gap-2">
                      <span class="text-xs font-bold text-[#18181B]">{{ mcp.name }}</span>
                      <span class="text-[9px] bg-black/[0.04] text-[#52525B] px-1.5 py-0.2 rounded font-mono">{{ mcp.type }}</span>
                    </div>
                    <div class="text-[11px] text-[#71717A] mt-0.5 font-mono">{{ mcp.command }} {{ (mcp.args || []).join(' ') }}</div>
                  </div>
                  <div class="flex items-center gap-2">
                    <button class="px-2 py-1 rounded-lg bg-white border border-black/[0.08] text-[11px] cursor-pointer" @click="s.testMcpAction(mcp.id)">探活</button>
                    <button class="px-2 py-1 rounded-lg bg-white border border-red-200 text-[11px] text-red-600 cursor-pointer" @click="s.deleteMcpAction(mcp.id)">删除</button>
                    <label class="relative inline-flex items-center cursor-pointer">
                      <input type="checkbox" v-model="mcp.enabled" @change="s.toggleMcp(mcp)" class="sr-only peer">
                      <div class="w-9 h-5 bg-gray-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-[#10A37F]"></div>
                    </label>
                  </div>
                </div>
              </div>
            </div>

            <!-- 选项卡 3: Skill 技能库 -->
            <div v-else-if="s.activeSettingsTab === 'skills'" class="space-y-4">
              <div class="flex items-center justify-between">
                <div>
                  <h3 class="text-xs font-bold text-[#18181B]">Agent 技能库 (Skills)</h3>
                  <p class="text-[11px] text-[#71717A]">已读写 ~/.tiancode/skills.json</p>
                </div>
                <button @click="s.isSkillModalOpen = true" class="px-3 py-1.5 rounded-xl bg-[#D96B27] text-white text-xs font-bold shadow-xs hover:bg-[#B8551B] cursor-pointer">
                  ➕ 创建新技能
                </button>
              </div>

              <div class="space-y-2">
                <div v-if="s.skills.length === 0" class="p-8 text-center bg-[#FAF8F5] rounded-xl border border-black/[0.06] text-[#71717A] text-xs">
                  <span class="text-2xl block mb-2">🛠️</span>
                  <span class="font-bold text-[#18181B] block mb-1">当前暂无自定义技能</span>
                  <p class="text-[11px] text-[#A1A1AA]">点击右上角可为智能体扩展专有技术栈提示词与工作流</p>
                </div>
                <div v-for="skill in s.skills" :key="skill.id" class="p-3 rounded-xl border border-black/[0.08] bg-[#FAF8F5] flex items-center justify-between shadow-2xs">
                  <div>
                    <span class="text-xs font-bold text-[#18181B]">{{ skill.name }}</span>
                    <div class="text-[11px] text-[#71717A] mt-0.5">{{ skill.description }}</div>
                  </div>
                  <div class="flex items-center gap-2">
                    <label class="relative inline-flex items-center cursor-pointer">
                      <input type="checkbox" v-model="skill.enabled" @change="s.toggleSkill(skill)" class="sr-only peer">
                      <div class="w-9 h-5 bg-gray-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-[#10A37F]"></div>
                    </label>
                    <button @click="s.deleteSkillAction(skill.id)" class="p-1 rounded text-red-500 hover:bg-red-50 cursor-pointer text-xs" title="删除技能">🗑️</button>
                  </div>
                </div>
              </div>
            </div>

            <!-- 选项卡 4: 软件规则 -->
            <div v-else-if="s.activeSettingsTab === 'rules'" class="space-y-4">
              <div class="flex items-center justify-between">
                <div>
                  <h3 class="text-xs font-bold text-[#18181B]">软件工程规则与提示词规约</h3>
                  <p class="text-[11px] text-[#71717A]">已读写 ~/.tiancode/rules.json · 自动注入大模型 System Prompt</p>
                </div>
                <button @click="s.isRuleModalOpen = true" class="px-3 py-1.5 rounded-xl bg-[#D96B27] text-white text-xs font-bold shadow-xs hover:bg-[#B8551B] cursor-pointer">
                  ➕ 添加规则
                </button>
              </div>

              <div class="space-y-2">
                <div v-if="s.rules.length === 0" class="p-8 text-center bg-[#FAF8F5] rounded-xl border border-black/[0.06] text-[#71717A] text-xs">
                  <span class="text-2xl block mb-2">📜</span>
                  <span class="font-bold text-[#18181B] block mb-1">当前暂无自定义工程规则</span>
                  <p class="text-[11px] text-[#A1A1AA]">点击右上角可配置规范守卫，自动在推理时注入智能体 System Prompt</p>
                </div>
                <div v-for="rule in s.rules" :key="rule.id" class="p-3 rounded-xl border border-black/[0.08] bg-[#FAF8F5] space-y-1 shadow-2xs">
                  <div class="flex items-center justify-between">
                    <span class="text-xs font-bold text-[#18181B]">{{ rule.title }}</span>
                    <div class="flex items-center gap-2">
                      <label class="relative inline-flex items-center cursor-pointer">
                        <input type="checkbox" v-model="rule.enabled" @change="s.toggleRule(rule)" class="sr-only peer">
                        <div class="w-9 h-5 bg-gray-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-[#10A37F]"></div>
                      </label>
                      <button @click="s.deleteRuleAction(rule.id)" class="p-1 rounded text-red-500 hover:bg-red-50 cursor-pointer text-xs" title="删除规则">🗑️</button>
                    </div>
                  </div>
                  <div class="text-[11px] text-[#52525B] font-mono bg-white p-2 rounded border border-black/[0.04]">{{ rule.content }}</div>
                </div>
              </div>
            </div>

            <div v-else-if="s.activeSettingsTab === 'theme'" class="space-y-3">
              <h3 class="text-xs font-bold text-[#18181B]">外观与工作区</h3>
              <div class="p-3 rounded-xl bg-[#FAF8F5] border border-black/[0.06] text-xs text-[#52525B] leading-relaxed space-y-2">
                <p>当前发货主题是陶土暖橙：底色 <span class="font-mono">#FAF8F5</span>，强调色 <span class="font-mono">#D96B27</span>。没有第二套可切换皮肤，因此这里不提供假开关。</p>
                <p>工作区路径：<span class="font-mono text-[#18181B]">{{ s.workspacePath || '尚未打开项目' }}</span></p>
                <p>用户数据目录：<span class="font-mono text-[#18181B]">~/.tiancode</span>（渠道、会话、MCP、技能、规则均落盘于此）。</p>
              </div>
            </div>
            <div v-else-if="s.activeSettingsTab === 'sandbox'" class="space-y-3">
              <h3 class="text-xs font-bold text-[#18181B]">安全沙箱与防线</h3>
              <div class="p-3 rounded-xl bg-[#FAF8F5] border border-black/[0.06] text-xs text-[#52525B] leading-relaxed space-y-2">
                <p>Agent 工具调用走 Go 微内核 <span class="font-mono">SafetyRail</span>：写文件、执行命令、Git 操作在内核侧校验，而不是前端开关。</p>
                <p>发送消息前会选择执行策略：只读分析会拦截写盘与命令；直接改代码 / TDD 才允许工具写文件。</p>
                <p>API Key 在 Windows 上用 DPAPI 加密写入 <span class="font-mono">~/.tiancode/channels.json</span>。</p>
              </div>
            </div>
            <div v-else-if="s.activeSettingsTab === 'about'" class="space-y-3">
              <h3 class="text-xs font-bold text-[#18181B]">关于 湉码</h3>
              <div class="p-3 rounded-xl bg-[#FAF8F5] border border-black/[0.06] text-xs text-[#52525B] leading-relaxed space-y-2">
                <p><strong class="text-[#18181B]">湉码 / tiancode</strong> · Wails v2 + Go 微内核 + Vue 3。</p>
                <p>仓库：<span class="font-mono">github.com/zhangqi77ok-sys/tiancode</span></p>
                <p>本页只陈述真实能力：流式对话、会话树、Git 分支/快照、终端、MCP stdio、技能/规则注入、Go AST 扫描。</p>
                <div class="pt-2 border-t border-black/[0.06] font-mono text-[11px] space-y-1">
                  <p>Token 累计：{{ s.usageMetrics.total_tokens }}</p>
                  <p>调用次数：{{ s.usageMetrics.total_calls }}</p>
                  <p>估算费用（非账单）：{{ s.usageMetrics.estimated_cost }}</p>
                  <p>活跃会话计数：{{ s.usageMetrics.active_sessions }}</p>
                  <p>更新时间：{{ s.usageMetrics.last_updated_time || '尚无记录' }}</p>
                </div>
              </div>
            </div>
          </main>
        </div>
      </div>
    </div>

    <!-- ========================================================================= -->
    <!-- 4. 项目知识图谱模态窗 (Knowledge Graph Modal) -->
    <!-- ========================================================================= -->
    <div
      v-if="s.isKnowledgeGraphOpen"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/45 backdrop-blur-xs animate-in fade-in duration-150 font-sans"
    >
      <div class="w-[92vw] max-w-[1200px] h-[86vh] bg-white rounded-2xl shadow-2xl border border-black/[0.1] flex flex-col overflow-hidden relative">
        <header class="h-12 bg-[#FAF8F5] border-b border-black/[0.08] flex items-center justify-between px-5 select-none shrink-0">
          <div class="flex items-center gap-3">
            <span class="text-base">🕸️</span>
            <span class="font-bold text-sm text-[#18181B]">工作区 Go AST 拓扑</span>
          </div>

          <div class="flex items-center gap-2">
            <button
              @click="s.scanASTGraph"
              class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-[#D96B27] text-white text-xs font-bold shadow-xs hover:bg-[#B8551B] cursor-pointer"
            >
              <span>🔄</span><span>代码扫描与图谱重建</span>
            </button>
            <button @click="s.isKnowledgeGraphOpen = false" class="p-1.5 rounded-lg text-[#71717A] hover:bg-black/[0.05] cursor-pointer">✕</button>
          </div>
        </header>

        <div class="flex-1 flex overflow-hidden">
          <!-- 拓扑节点列表 -->
          <div class="flex-1 p-5 overflow-y-auto bg-[#FAF8F5] space-y-3">
            <div v-if="s.isGraphLoading" class="p-12 text-center text-[#71717A] text-xs flex flex-col items-center justify-center gap-3 mt-12">
              <span class="animate-spin text-3xl">⏳</span>
              <span class="font-bold text-[#18181B] text-sm">正在深度解析工作区 Go AST 语法拓扑树...</span>
              <p class="text-[11px] text-[#A1A1AA]">提取代码包、结构体、接口与依赖实体，请稍候</p>
            </div>
            <div v-else-if="s.astNodes.length === 0" class="p-12 text-center text-[#71717A] text-xs flex flex-col items-center justify-center gap-2 mt-12">
              <span class="text-3xl">🕸️</span>
              <span class="font-bold text-[#18181B]">暂无代码拓扑节点</span>
              <p class="text-[11px] text-[#A1A1AA]">点击右上角【代码扫描与图谱重建】即可扫描当前工作区</p>
            </div>
            <div v-else>
              <div class="text-xs font-bold text-[#71717A] uppercase mb-2">AST 拓扑 ({{ s.astNodes.length }} 节点，图中最多 80)</div>
              <div class="mb-3 overflow-auto rounded-xl border border-black/[0.08] bg-[#18181B] max-h-[48vh]">
                <svg :width="s.astGraph.maxX" :height="s.astGraph.maxY">
                  <line
                    v-for="(e, i) in s.astGraph.edges"
                    :key="'e'+i"
                    :x1="e.x1" :y1="e.y1" :x2="e.x2" :y2="e.y2"
                    stroke="#D96B27" stroke-opacity="0.45"
                  />
                  <g
                    v-for="p in s.astGraph.pos"
                    :key="p.id"
                    @click="s.selectedAstNode = s.astNodes.find(n => n.id === p.id) || s.selectedAstNode"
                    class="cursor-pointer"
                  >
                    <circle :cx="p.x" :cy="p.y" r="10" :fill="s.selectedAstNode?.id === p.id ? '#D96B27' : '#FAF8F5'" />
                    <text :x="p.x + 14" :y="p.y + 4" fill="#F4F4F5" font-size="10">{{ p.name }}</text>
                  </g>
                </svg>
              </div>
              <div class="grid grid-cols-2 gap-3">
                <div
                  v-for="node in s.astNodes"
                  :key="node.id"
                  @click="s.selectedAstNode = node"
                  :class="[
                    'p-3.5 rounded-2xl border bg-white shadow-2xs flex items-center justify-between cursor-pointer transition-all',
                    s.selectedAstNode?.id === node.id ? 'border-2 border-[#D96B27] ring-2 ring-[#D96B27]/20' : 'border-black/[0.08] hover:border-[#D96B27]/40'
                  ]"
                >
                  <div>
                    <div class="flex items-center gap-2">
                      <span class="text-sm">{{ node.type === 'package' ? '📦' : (node.type === 'struct' ? '🏛️' : '📄') }}</span>
                      <span class="text-xs font-bold text-[#18181B] font-mono">{{ node.name }}</span>
                      <span class="text-[9px] bg-[#D96B27]/10 text-[#D96B27] px-1.5 py-0.2 rounded font-mono font-bold">{{ node.type }}</span>
                    </div>
                    <div class="text-[11px] text-[#71717A] mt-1 font-mono">{{ node.file }}</div>
                  </div>
                </div>
              </div>
            </div>
          </div>


          <!-- 实体详情侧板 -->
          <aside class="w-80 border-l border-black/[0.08] bg-white p-5 flex flex-col justify-between overflow-y-auto">
            <div v-if="s.selectedAstNode" class="space-y-4 text-xs">
              <div class="flex items-center gap-2 pb-3 border-b border-black/[0.06]">
                <span class="text-xl">🏛️</span>
                <div>
                  <h4 class="font-bold text-sm text-[#18181B]">{{ s.selectedAstNode.name }}</h4>
                  <span class="text-[10px] text-[#D96B27] bg-[#D96B27]/10 px-1.5 py-0.2 rounded font-mono">{{ s.selectedAstNode.type }}</span>
                </div>
              </div>
              <div>
                <span class="font-bold text-[#71717A]">源文件位置</span>
                <p class="font-mono text-[11px] text-[#18181B] mt-1 bg-[#FAF8F5] p-2 rounded border border-black/[0.04]">{{ s.selectedAstNode.file }}</p>
              </div>
              <div>
                <span class="font-bold text-[#71717A]">拓扑摘要</span>
                <p class="text-[11px] text-[#52525B] leading-relaxed mt-1">{{ s.selectedAstNode.details }}</p>
              </div>
            </div>

            <button
              v-if="s.selectedAstNode"
              @click="s.injectNodeToPrompt"
              class="w-full py-2 rounded-xl bg-[#D96B27] text-white text-xs font-bold shadow-xs hover:bg-[#B8551B] cursor-pointer flex items-center justify-center gap-1.5 mt-4"
            >
              <span>📌</span><span>引用该节点架构约束至对话</span>
            </button>
          </aside>
        </div>
      </div>
    </div>

    <!-- 渠道编辑弹窗 -->
    <div
      v-if="s.isChannelModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/45 backdrop-blur-xs font-sans"
    >
      <div class="w-full max-w-md bg-white rounded-2xl shadow-2xl border border-black/[0.1] p-5 space-y-4">
        <h4 class="text-sm font-bold text-[#18181B]">渠道配置管理</h4>
        <div class="space-y-3 text-xs">
          <div>
            <label class="block font-medium text-[#71717A] mb-1">渠道名称</label>
            <input v-model="s.channelForm.name" type="text" class="w-full px-2.5 py-1.5 rounded-lg border border-black/[0.1] focus:outline-none focus:border-[#D96B27]">
          </div>
          <div>
            <label class="block font-medium text-[#71717A] mb-1">API Base URL</label>
            <input v-model="s.channelForm.endpoint" type="text" class="w-full px-2.5 py-1.5 rounded-lg border border-black/[0.1] focus:outline-none focus:border-[#D96B27]">
          </div>
          <div>
            <label class="block font-medium text-[#71717A] mb-1">API Key</label>
            <input v-model="s.channelForm.api_key" type="password" class="w-full px-2.5 py-1.5 rounded-lg border border-black/[0.1] focus:outline-none focus:border-[#D96B27]">
          </div>
          <button @click="s.fetchModelsAction" class="w-full py-1.5 rounded-lg border border-[#D96B27] text-[#D96B27] text-xs font-bold hover:bg-[#D96B27]/10 cursor-pointer">
            🔄 真实自动获取上游模型 (/v1/models)
          </button>
        </div>

        <div class="flex justify-end gap-2 pt-2 border-t border-black/[0.06]">
          <button @click="s.isChannelModalOpen = false" class="px-3 py-1 rounded-lg border border-black/[0.1] text-xs">取消</button>
          <button @click="s.saveChannelAction" class="px-4 py-1 rounded-lg bg-[#D96B27] text-white text-xs font-semibold hover:bg-[#B8551B]">保存至磁盘</button>
        </div>
      </div>
    </div>

    <!-- MCP 导入配置弹窗 -->
    <div
      v-if="s.isMcpModalOpen"
      @keydown.esc="s.isMcpModalOpen = false"
      tabindex="-1"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/45 backdrop-blur-xs font-sans"
    >
      <div class="w-full max-w-md bg-white rounded-2xl shadow-2xl border border-black/[0.1] p-5 space-y-4">
        <div class="flex items-center justify-between pb-2 border-b border-black/[0.06]">
          <h4 class="text-sm font-bold text-[#18181B] flex items-center gap-1.5">
            <span>🧩</span><span>导入 MCP 服务配置</span>
          </h4>
          <button @click="s.isMcpModalOpen = false" class="text-[#71717A] hover:text-[#18181B] p-1 rounded-md cursor-pointer" title="关闭弹窗 (Esc)">✕</button>
        </div>
        <div class="space-y-3 text-xs">
          <div>
            <label class="block font-medium text-[#71717A] mb-1">服务名称</label>
            <input v-model="s.mcpForm.name" placeholder="如 fetch-mcp 或 filesystem" type="text" class="w-full px-2.5 py-1.5 rounded-lg border border-black/[0.1] focus:outline-none focus:border-[#D96B27]">
          </div>
          <div>
            <label class="block font-medium text-[#71717A] mb-1">通信类型</label>
            <select v-model="s.mcpForm.type" class="w-full px-2.5 py-1.5 rounded-lg border border-black/[0.1] focus:outline-none focus:border-[#D96B27]">
              <option value="stdio">stdio（当前内核已实现）</option>
            </select>
          </div>
          <div>
            <label class="block font-medium text-[#71717A] mb-1">启动命令 (Command)</label>
            <input v-model="s.mcpForm.command" placeholder="如 npx 或 python" type="text" class="w-full px-2.5 py-1.5 rounded-lg border border-black/[0.1] focus:outline-none focus:border-[#D96B27]">
          </div>
          <div>
            <label class="block font-medium text-[#71717A] mb-1">启动参数 (以空格隔开)</label>
            <input v-model="s.mcpArgsInput" placeholder="-y @modelcontextprotocol/server-filesystem D:/workspace" type="text" class="w-full px-2.5 py-1.5 rounded-lg border border-black/[0.1] focus:outline-none focus:border-[#D96B27]">
          </div>
        </div>
        <div class="flex justify-end gap-2 pt-2 border-t border-black/[0.06]">
          <button @click="s.isMcpModalOpen = false" class="px-3 py-1 rounded-lg border border-black/[0.1] text-xs cursor-pointer">取消</button>
          <button @click="s.saveMcpAction" class="px-4 py-1 rounded-lg bg-[#D96B27] text-white text-xs font-semibold hover:bg-[#B8551B] cursor-pointer">保存 MCP 服务</button>
        </div>
      </div>
    </div>

    <!-- Skill 新增弹窗 -->
    <div
      v-if="s.isSkillModalOpen"
      @keydown.esc="s.isSkillModalOpen = false"
      tabindex="-1"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/45 backdrop-blur-xs font-sans"
    >
      <div class="w-full max-w-md bg-white rounded-2xl shadow-2xl border border-black/[0.1] p-5 space-y-4">
        <div class="flex items-center justify-between pb-2 border-b border-black/[0.06]">
          <h4 class="text-sm font-bold text-[#18181B] flex items-center gap-1.5">
            <span>🛠️</span><span>创建 Agent 技能 (Skill)</span>
          </h4>
          <button @click="s.isSkillModalOpen = false" class="text-[#71717A] hover:text-[#18181B] p-1 rounded-md cursor-pointer" title="关闭弹窗 (Esc)">✕</button>
        </div>
        <div class="space-y-3 text-xs">
          <div>
            <label class="block font-medium text-[#71717A] mb-1">技能名称</label>
            <input v-model="s.skillForm.name" placeholder="如 vue3-expert 或 rust-analyzer" type="text" class="w-full px-2.5 py-1.5 rounded-lg border border-black/[0.1] focus:outline-none focus:border-[#D96B27]">
          </div>
          <div>
            <label class="block font-medium text-[#71717A] mb-1">描述与职责说明</label>
            <input v-model="s.skillForm.description" placeholder="专有技术栈模式、规约与实现导向" type="text" class="w-full px-2.5 py-1.5 rounded-lg border border-black/[0.1] focus:outline-none focus:border-[#D96B27]">
          </div>
          <div>
            <label class="block font-medium text-[#71717A] mb-1">提示词与技能正文</label>
            <textarea v-model="s.skillForm.content" rows="4" placeholder="在此输入注入大模型系统指令的专业技能提示词..." class="w-full px-2.5 py-1.5 rounded-lg border border-black/[0.1] focus:outline-none focus:border-[#D96B27] resize-none"></textarea>
          </div>
        </div>
        <div class="flex justify-end gap-2 pt-2 border-t border-black/[0.06]">
          <button @click="s.isSkillModalOpen = false" class="px-3 py-1 rounded-lg border border-black/[0.1] text-xs cursor-pointer">取消</button>
          <button @click="s.saveSkillAction" class="px-4 py-1 rounded-lg bg-[#D96B27] text-white text-xs font-semibold hover:bg-[#B8551B] cursor-pointer">保存技能</button>
        </div>
      </div>
    </div>

    <!-- Rule 新增弹窗 -->
    <div
      v-if="s.isRuleModalOpen"
      @keydown.esc="s.isRuleModalOpen = false"
      tabindex="-1"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/45 backdrop-blur-xs font-sans"
    >
      <div class="w-full max-w-md bg-white rounded-2xl shadow-2xl border border-black/[0.1] p-5 space-y-4">
        <div class="flex items-center justify-between pb-2 border-b border-black/[0.06]">
          <h4 class="text-sm font-bold text-[#18181B] flex items-center gap-1.5">
            <span>📜</span><span>添加工程规约与规则 (Rule)</span>
          </h4>
          <button @click="s.isRuleModalOpen = false" class="text-[#71717A] hover:text-[#18181B] p-1 rounded-md cursor-pointer" title="关闭弹窗 (Esc)">✕</button>
        </div>
        <div class="space-y-3 text-xs">
          <div>
            <label class="block font-medium text-[#71717A] mb-1">规则名称</label>
            <input v-model="s.ruleForm.title" placeholder="如 铁律 0.5 严禁假数据" type="text" class="w-full px-2.5 py-1.5 rounded-lg border border-black/[0.1] focus:outline-none focus:border-[#D96B27]">
          </div>
          <div>
            <label class="block font-medium text-[#71717A] mb-1">规则内容</label>
            <textarea v-model="s.ruleForm.content" rows="4" placeholder="在此输入强制约束与守卫提示词..." class="w-full px-2.5 py-1.5 rounded-lg border border-black/[0.1] focus:outline-none focus:border-[#D96B27] resize-none"></textarea>
          </div>
        </div>
        <div class="flex justify-end gap-2 pt-2 border-t border-black/[0.06]">
          <button @click="s.isRuleModalOpen = false" class="px-3 py-1 rounded-lg border border-black/[0.1] text-xs cursor-pointer">取消</button>
          <button @click="s.saveRuleAction" class="px-4 py-1 rounded-lg bg-[#D96B27] text-white text-xs font-semibold hover:bg-[#B8551B] cursor-pointer">保存规则</button>
        </div>
      </div>
    </div>

    <div
      v-if="s.isCommandPaletteOpen"
      class="fixed inset-0 z-[60] flex items-start justify-center bg-black/40 pt-[12vh]"
      @click.self="s.isCommandPaletteOpen = false"
    >
      <div class="w-[min(640px,90vw)] bg-white rounded-2xl shadow-2xl border border-black/[0.1] overflow-hidden">
        <input
          v-model="s.commandPaletteQuery"
          type="text"
          autofocus
          placeholder="检索会话、文件、设置、终端…"
          class="w-full px-4 py-3 text-sm border-b border-black/[0.08] focus:outline-none"
          @keydown.down.prevent="s.moveCommandPalette(1)"
          @keydown.up.prevent="s.moveCommandPalette(-1)"
          @keydown.enter.prevent="s.confirmCommandPalette()"
        />
        <div class="max-h-[50vh] overflow-y-auto py-1">
          <div v-if="s.commandPaletteItems.length === 0" class="px-4 py-6 text-xs text-[#71717A]">没有匹配项</div>
          <button
            v-for="(item, idx) in s.commandPaletteItems"
            :key="item.id"
            class="w-full text-left px-4 py-2 text-xs flex items-center justify-between cursor-pointer"
            :class="idx === s.commandPaletteIndex ? 'bg-[#D96B27]/10 text-[#18181B]' : 'hover:bg-black/[0.03]'"
            @click="s.runCommandPaletteItem(item)"
          >
            <span class="font-medium truncate">{{ item.label }}</span>
            <span class="text-[10px] text-[#A1A1AA] ml-3 shrink-0">{{ item.kind }}{{ item.hint ? ' · ' + item.hint : '' }}</span>
          </button>
        </div>
      </div>
    </div>

    <div
      v-if="s.isStrategyPickerOpen"
      class="fixed inset-0 z-[65] flex items-center justify-center bg-black/45"
      @click.self="s.isStrategyPickerOpen = false"
    >
      <div class="w-[min(720px,92vw)] bg-white rounded-2xl border border-black/[0.1] shadow-2xl p-4 space-y-3">
        <div class="flex items-center justify-between">
          <div>
            <h4 class="text-sm font-bold text-[#18181B]">架构策略决策确认</h4>
            <p class="text-[11px] text-[#71717A]">选择会改内核工具权限，不是装饰。只读分析会拦截写文件和执行命令。</p>
          </div>
          <span class="text-[9px] text-[#D96B27] bg-[#D96B27]/10 px-1.5 py-0.5 rounded font-mono font-bold">待用户决策</span>
        </div>
        <div class="grid grid-cols-1 md:grid-cols-3 gap-2.5">
          <button
            v-for="opt in s.executionStrategies"
            :key="opt.id"
            type="button"
            class="text-left p-3 rounded-xl cursor-pointer transition-all"
            :class="s.executionStrategy === opt.id ? 'border-2 border-[#D96B27] bg-white shadow-xs' : 'border border-black/[0.08] opacity-85 hover:opacity-100'"
            @click="s.executionStrategy = opt.id"
          >
            <div class="flex items-center justify-between gap-1">
              <div class="flex items-center gap-1.5 font-bold text-xs text-[#18181B]">
                <span class="w-3.5 h-3.5 rounded-full bg-[#D96B27] text-white flex items-center justify-center text-[9px]">{{ opt.letter }}</span>
                <span>{{ opt.title }}</span>
              </div>
              <span class="text-[9px] font-mono font-bold text-[#10A37F] bg-emerald-50 px-1.5 py-0.2 rounded">{{ opt.badge }}</span>
            </div>
            <p class="text-[11px] text-[#71717A] mt-1.5 leading-relaxed">{{ opt.desc }}</p>
          </button>
        </div>
        <input
          v-model="s.strategyNote"
          type="text"
          placeholder="补充约束（可选，会写入系统提示并随策略一起生效）"
          class="w-full h-8 px-2.5 rounded-lg border border-black/[0.08] text-xs"
        />
        <div class="flex justify-end gap-2 pt-1 border-t border-black/[0.06]">
          <button class="px-3 py-1 rounded-lg text-xs text-[#71717A] cursor-pointer" @click="s.skipStrategyChoice">跳过并采用直接改代码</button>
          <button class="px-4 py-1 rounded-lg bg-[#D96B27] text-white text-xs font-semibold cursor-pointer" @click="s.confirmStrategyAndSend">确定提交选择</button>
        </div>
      </div>
    </div>

    <div
      v-if="s.tabContextMenu"
      class="fixed z-[70] bg-white border border-black/[0.1] rounded-lg shadow-lg text-xs py-1 min-w-[140px]"
      :style="{ left: s.tabContextMenu.x + 'px', top: s.tabContextMenu.y + 'px' }"
      @click.self="s.tabContextMenu = null"
    >
      <button class="w-full text-left px-3 py-1.5 hover:bg-black/[0.04] cursor-pointer" @click="s.closeSessionTab(s.tabContextMenu.id)">关闭</button>
      <button class="w-full text-left px-3 py-1.5 hover:bg-black/[0.04] cursor-pointer" @click="s.closeOtherTabs(s.tabContextMenu.id)">关闭其他</button>
      <button class="w-full text-left px-3 py-1.5 hover:bg-black/[0.04] cursor-pointer" @click="s.closeAllTabs()">关闭全部</button>
    </div>

    <!-- 全局 Toast 提示 -->
    <div
      v-if="s.toastMessage"
      class="fixed bottom-5 right-5 z-50 px-3.5 py-2 rounded-xl bg-[#18181B] text-white text-xs font-medium shadow-2xl flex items-center gap-2 border border-white/[0.1] animate-in slide-in-from-bottom-3 duration-200"
    >
      <span>{{ s.toastMessage }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { wailsBridge } from './core/wailsBridge'
import { useWorkbenchStore } from './stores/workbench'
import ActivityBar from './components/ActivityBar.vue'
import LeftDrawer from './components/LeftDrawer.vue'
import ChatCockpit from './components/ChatCockpit.vue'
import DiffWorkspace from './components/DiffWorkspace.vue'
import TerminalDrawer from './components/TerminalDrawer.vue'

const s = useWorkbenchStore()
let stop: (() => void) | undefined
onMounted(() => { stop = s.initWorkbench() })
onUnmounted(() => { stop?.() })
</script>

<style>
button {
  transition: transform 0.08s ease, background-color 0.15s ease, opacity 0.15s ease;
}
button:active {
  transform: scale(0.96);
}

.markdown-body pre {
  background-color: #18181B;
  color: #F4F4F5;
  padding: 0.75rem;
  border-radius: 0.5rem;
  overflow-x: auto;
  font-family: 'Fira Code', monospace;
  margin: 0.5rem 0;
  border: 1px solid rgba(255, 255, 255, 0.1);
}
.markdown-body code {
  font-family: 'Fira Code', monospace;
  background-color: rgba(0, 0, 0, 0.05);
  padding: 0.1rem 0.3rem;
  border-radius: 0.25rem;
}
.markdown-body pre code {
  background-color: transparent;
  padding: 0;
}
.markdown-body p {
  margin-bottom: 0.5rem;
}
.markdown-body ul, .markdown-body ol {
  padding-left: 1.25rem;
  margin-bottom: 0.5rem;
}
.markdown-body ul {
  list-style-type: disc;
}
.markdown-body ol {
  list-style-type: decimal;
}
</style>


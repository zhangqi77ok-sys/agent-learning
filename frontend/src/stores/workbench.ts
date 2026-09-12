import { defineStore } from 'pinia'
import { ref, reactive, computed, nextTick } from 'vue'
import {
  wailsBridge,
  type SessionMeta,
  type ChatSession,
  type FileNode,
  type DiffReport,
  type ChannelConfig,
  type MCPServerConfig,
  type SkillConfig,
  type RuleConfig,
  type GraphNode
} from '../core/wailsBridge'
import { renderMarkdown } from '../core/markdown'

export const useWorkbenchStore = defineStore('workbench', () => {
// 1. 活动栏与工作区状态
const activeActivity = ref('chat')
const isDiffOpen = ref(false)
const activeDiffFile = ref('')
const isSettingsOpen = ref(false)
const isKnowledgeGraphOpen = ref(false)
const isChannelModalOpen = ref(false)
const isMcpModalOpen = ref(false)
const isSkillModalOpen = ref(false)
const isRuleModalOpen = ref(false)
const activeSettingsTab = ref('models')
const isFullAuto = ref(false)
const isStreaming = ref(false)

const isGraphLoading = ref(false)
const isFileTreeLoading = ref(false)
const isGitLoading = ref(false)

const toastMessage = ref('')
function showToast(msg: string) {
  toastMessage.value = msg
  setTimeout(() => {
    toastMessage.value = ''
  }, 2500)
}

// 2. 真实会话管理 (读写 ~/.tcode/sessions/)
const sessions = ref<SessionMeta[]>([])
const activeTag = ref('全部')
const currentSessionId = ref('')
const selectedModel = ref('')

const currentSession = ref<ChatSession>({
  id: '',
  title: '新工程对话',
  model: '',
  tag: '',
  created_at: Date.now(),
  updated_at: Date.now(),
  messages: []
})

const availableTags = computed(() => {
  const set = new Set<string>()
  sessions.value.forEach(s => {
    if (s.tag && s.tag.trim()) set.add(s.tag.trim())
  })
  return ['全部', ...Array.from(set)]
})

const filteredSessions = computed(() => {
  if (activeTag.value === '全部') return sessions.value
  return sessions.value.filter(s => (s.tag || '') === activeTag.value)
})

const historyCap = 80
const showFullHistory = ref(false)
const hiddenHistoryCount = computed(() => {
  const n = (currentSession.value.messages || []).length
  if (showFullHistory.value || n <= historyCap) return 0
  return n - historyCap
})
const visibleMessages = computed(() => {
  const msgs = currentSession.value.messages || []
  if (showFullHistory.value || msgs.length <= historyCap) return msgs
  return msgs.slice(-historyCap)
})
function revealFullHistory() {
  showFullHistory.value = true
}

const upstreamFetchedModels = ref<string[]>([])

const availableModels = computed(() => {
  const set = new Set<string>()
  upstreamFetchedModels.value.forEach(m => {
    if (m && m.trim()) set.add(m.trim())
  })
  channels.value.forEach(c => {
    if (c.model && c.model.trim()) set.add(c.model.trim())
  })
  return Array.from(set)
})

async function loadSessionsList() {
  try {
    const list = await wailsBridge.listSessions()
    sessions.value = list || []
  } catch (err) {
    console.error('Failed to load sessions:', err)
    sessions.value = []
  }
}

async function selectSession(id: string) {
  showFullHistory.value = false
  currentSessionId.value = id
  try {
    const sess = await wailsBridge.getSession(id)
    if (sess) {
      currentSession.value = sess
      if (sess.model) selectedModel.value = sess.model
    }
    showToast(`✓ 已载入会话: ${currentSession.value.title}`)
  } catch (err) {
    console.error('Failed to load session:', err)
  }
}

async function createNewSession() {
  const newId = 'sess_' + Date.now()
  const newSess: ChatSession = {
    id: newId,
    title: '新工程对话',
    model: selectedModel.value,
    tag: '',
    created_at: Date.now(),
    updated_at: Date.now(),
    messages: []
  }
  await wailsBridge.saveSession(newSess)
  await loadSessionsList()
  await selectSession(newId)
  showToast('✓ 已新建会话并持久化')
}

async function deleteSession(id: string) {
  if (currentSessionId.value === id && isStreaming.value) {
    await stopGenerationAction()
  }
  await wailsBridge.deleteSession(id)
  await loadSessionsList()
  if (currentSessionId.value === id) {
    if (sessions.value.length > 0) {
      await selectSession(sessions.value[0].id)
    } else {
      currentSessionId.value = ''
      currentSession.value = {
        id: '',
        title: '新工程对话',
        model: selectedModel.value,
        tag: '',
        created_at: Date.now(),
        updated_at: Date.now(),
        messages: []
      }
    }
  }
  showToast('✓ 会话已从本地磁盘移除')
}

// 3. 真实工作区、文件树与 Git 状态
const workspacePath = ref('')
const workspaceName = computed(() => {
  if (!workspacePath.value) return '湉码'
  const normalized = workspacePath.value.replace(/\\/g, '/')
  const parts = normalized.split('/').filter(Boolean)
  return parts[parts.length - 1] || 'Workspace'
})

async function chooseWorkspace() {
  try {
    const selected = await wailsBridge.openDirectoryDialog()
    if (selected && selected !== workspacePath.value) {
      await wailsBridge.setWorkspace(selected)
      workspacePath.value = selected
      await Promise.all([
        loadFileTree(),
        loadGitStatus(),
        scanASTGraph()
      ])
      showToast(`✓ 已成功切换至工作区: ${workspaceName.value}`)
    }
  } catch (err: any) {
    showToast(`切换工作区失败: ${err}`)
  }
}

const fileTree = ref<FileNode[]>([])
const expandedFolders = reactive<Record<string, boolean>>({ 'frontend': true })
const gitStatus = ref<any>({ branch: 'main', working: [], staged: [] })
const commitMessage = ref('')

const stagedTreeFiles = computed(() => {
  const list: { path: string; type: string; color: string }[] = []
  if (gitStatus.value?.staged && Array.isArray(gitStatus.value.staged)) {
    for (const f of gitStatus.value.staged) {
      if (typeof f === 'string') {
        list.push({ path: f, type: 'M', color: 'text-[#10A37F]' })
      } else if (f && typeof f === 'object') {
        const type = f.staged_code || f.index_code || 'M'
        list.push({
          path: f.path || '',
          type: type,
          color: type === 'D' ? 'text-red-500' : 'text-[#10A37F]'
        })
      }
    }
  }
  return list
})

const workingTreeFiles = computed(() => {
  const list: { path: string; type: string; color: string }[] = []
  if (gitStatus.value?.working && Array.isArray(gitStatus.value.working)) {
    for (const f of gitStatus.value.working) {
      if (typeof f === 'string') {
        list.push({ path: f, type: 'M', color: 'text-amber-600' })
      } else if (f && typeof f === 'object') {
        list.push({
          path: f.path || '',
          type: f.work_code || 'M',
          color: f.work_code === 'D' ? 'text-red-500' : 'text-amber-600'
        })
      }
    }
  }
  if (gitStatus.value?.untracked && Array.isArray(gitStatus.value.untracked)) {
    for (const p of gitStatus.value.untracked) {
      list.push({
        path: typeof p === 'string' ? p : (p as any)?.path || '',
        type: 'U',
        color: 'text-emerald-600'
      })
    }
  }
  return list
})

async function loadFileTree() {
  fileTree.value = await wailsBridge.getFileTree()
}

async function loadGitStatus() {
  gitStatus.value = await wailsBridge.getGitStatus()
}

function switchToFileActivity() {
  activeActivity.value = 'files'
  loadFileTree()
}

function switchToGitActivity() {
  activeActivity.value = 'git'
  loadGitStatus()
}

watch(() => store.gitVersion, () => {
  loadGitStatus()
})

function handleFileClick(node: FileNode) {
  if (node.is_dir) {
    expandedFolders[node.path] = !expandedFolders[node.path]
  } else {
    openFileDiff(node.path)
  }
}

async function handleGitCommit() {
  if (!commitMessage.value.trim()) return
  try {
    await wailsBridge.gitCommit(commitMessage.value.trim())
    commitMessage.value = ''
    await loadGitStatus()
    showToast('✓ Git 变更已成功提交本地仓库！')
  } catch (err) {
    showToast('提交异常: ' + err)
  }
}

// 4. 真实物理代码 Diff
const diffReport = ref<DiffReport | null>(null)

async function openFileDiff(filePath: string) {
  activeDiffFile.value = filePath
  isDiffOpen.value = true
  await loadDiff()
}

async function loadDiff() {
  if (!activeDiffFile.value) {
    diffReport.value = null
    return
  }
  try {
    diffReport.value = await wailsBridge.getStructuredDiff(activeDiffFile.value)
  } catch (err) {
    console.error('Diff error:', err)
  }
}

async function unstageFileAction(filePath: string, event?: MouseEvent) {
  if (event) event.stopPropagation()
  if (!filePath) return
  try {
    await wailsBridge.gitUnstage(filePath)
    await loadGitStatus()
    if (activeDiffFile.value === filePath) {
      await loadDiff()
    }
    showToast(`✓ 已取消暂存: ${filePath}`)
  } catch (err) {
    showToast(`取消暂存异常: ${err}`)
  }
}

async function revertFileAction() {
  if (!activeDiffFile.value) return
  try {
    await wailsBridge.revertFile(activeDiffFile.value)
    await loadDiff()
    await loadGitStatus()
    showToast(`✓ 已物理撤回 ${activeDiffFile.value} 磁盘改动 (Git Checkout)`)
  } catch (err) {
    showToast('撤回异常: ' + err)
  }
}

async function stageFileAction() {
  try {
    if (!activeDiffFile.value) return
    await wailsBridge.gitStage(activeDiffFile.value)
    showToast(`✓ 已成功采纳并暂存变更: ${activeDiffFile.value}`)
    await loadDiff()
    await loadGitStatus()
    isDiffOpen.value = false
  } catch (err) {
    showToast(`采纳文件变更异常: ${err}`)
  }
}

async function applyHunkAction(hunkIndex: number, stageOnly: boolean = true) {
  try {
    showToast(`⏳ 正在采纳 [${activeDiffFile.value}] 第 #${hunkIndex + 1} 个变更块...`)
    await wailsBridge.applyDiffHunk(activeDiffFile.value, hunkIndex, stageOnly)
    showToast(`✓ 已成功采纳该块变更 (git apply --cached)`)
    await loadDiff()
    await loadGitStatus()
  } catch (err) {
    showToast(`采纳变更块异常: ${err}`)
  }
}

async function discardHunkAction(hunkIndex: number) {
  try {
    showToast(`⏳ 正在丢弃 [${activeDiffFile.value}] 第 #${hunkIndex + 1} 个变更块...`)
    await wailsBridge.discardDiffHunk(activeDiffFile.value, hunkIndex)
    showToast(`✓ 已成功丢弃撤销该块变更 (git apply --reverse)`)
    await loadDiff()
    await loadGitStatus()
  } catch (err) {
    showToast(`丢弃变更块异常: ${err}`)
  }
}

// 5. 对话输入与流式大模型推理
const inputPrompt = ref('')
const attachedFiles = ref<string[]>([])
const messagesContainerRef = ref<HTMLDivElement | null>(null)

async function triggerUpload() {
  try {
    const selected = await wailsBridge.openFileDialog()
    if (selected && selected.length > 0) {
      for (const item of selected) {
        if (!attachedFiles.value.includes(item)) attachedFiles.value.push(item)
      }
    }
  } catch (err) {
    console.error('File dialog error:', err)
  }
}

async function handleSend() {
  const prompt = inputPrompt.value.trim()
  if (!prompt || isStreaming.value) return
  const slash = prompt.split(/\s+/)[0]
  if (slash === '/test' || slash === '/tdd') {
    inputPrompt.value = ''
    showToast('正在运行工作区真实测试…')
    try {
      const report = await wailsBridge.runTDDValidation()
      const snippet = (report.output || '').slice(0, 120)
      if (report.status === 'PASS') {
        showToast(`✓ TDD PASS passed=${report.passed} ${snippet}`)
      } else {
        showToast(`TDD ${report.status} failed=${report.failed} ${snippet}`)
      }
    } catch (err) {
      showToast('TDD 无法执行: ' + err)
    }
    return
  }
  if (slash === '/diff') {
    inputPrompt.value = ''
    activeActivity.value = 'git'
    await loadGitStatus()
    isDiffOpen.value = true
    showToast('已打开真实 Git 状态（无改动则为空）')
    return
  }

  if (!selectedModel.value) {
    showToast('请先在设置中添加模型渠道并选择模型，不会使用内置假模型')
    return
  }

  let fullPrompt = prompt
  if (attachedFiles.value.length > 0) {
    fullPrompt = `[附加关联文件]\n${attachedFiles.value.map(f => `- ${f}`).join('\n')}\n\n${prompt}`
    attachedFiles.value = []
  }

  // 保证会话 ID 绝对非空，避免向后端传入空 session_id 生成畸形文件
  if (!currentSessionId.value) {
    const newId = 'sess_' + Date.now()
    currentSessionId.value = newId
    currentSession.value.id = newId
    currentSession.value.title = prompt.slice(0, 15)
    await wailsBridge.saveSession(currentSession.value)
    await loadSessionsList()
  }

  inputPrompt.value = ''
  isStreaming.value = true

  const userMsgId = 'msg_' + Date.now() + '_' + Math.random().toString(36).slice(2, 7)
  currentSession.value.messages.push({
    id: userMsgId,
    role: 'user',
    content: fullPrompt,
    time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  })

  const asstMsgId = 'asst_' + Date.now() + '_' + Math.random().toString(36).slice(2, 7)
  currentSession.value.messages.push({
    id: asstMsgId,
    role: 'assistant',
    content: '',
    thinking: '',
    time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  })

  await nextTick()
  if (messagesContainerRef.value) {
    messagesContainerRef.value.scrollTop = messagesContainerRef.value.scrollHeight
  }

  try {
    await wailsBridge.sendMessage(
      {
        session_id: currentSessionId.value,
        prompt: fullPrompt,
        model: selectedModel.value,
        is_full_auto: isFullAuto.value
      },
      {
        onThinking(thinking) {
          const target = currentSession.value.messages.find(m => m.id === asstMsgId)
          if (target) target.thinking = (target.thinking || '') + thinking
          if (messagesContainerRef.value) messagesContainerRef.value.scrollTop = messagesContainerRef.value.scrollHeight
        },
        onChunk(delta) {
          const target = currentSession.value.messages.find(m => m.id === asstMsgId)
          if (target) target.content += delta
          if (messagesContainerRef.value) messagesContainerRef.value.scrollTop = messagesContainerRef.value.scrollHeight
        },
        onToolStart(tool, args, tcId) {
          const target = currentSession.value.messages.find(m => m.id === asstMsgId)
          let parsedArgs = args
          if (typeof args === 'string') {
            try {
              parsedArgs = JSON.parse(args)
            } catch (_) {
              parsedArgs = args
            }
          }
          if (target) {
            const toolRecord = { id: tcId || 'tool_' + Date.now(), name: tool, args: parsedArgs, output: '正在执行...' }
            target.tool = toolRecord
            if (!target.tools) target.tools = []
            target.tools.push(toolRecord)
          }
        },
        onToolEnd(tool, output, tcId) {
          const target = currentSession.value.messages.find(m => m.id === asstMsgId)
          if (target) {
            if (target.tool && target.tool.name === tool) {
              target.tool.output = output
            }
            if (target.tools && target.tools.length > 0) {
              const matched = tcId ? target.tools.find(t => t.id === tcId) : target.tools[target.tools.length - 1]
              if (matched) matched.output = output
            }
          }
        },
        onDone() {
          isStreaming.value = false
          wailsBridge.saveSession(currentSession.value)
          showToast('✓ 智能体推理与持久化完毕')
        }
      }
    )
  } catch (err) {
    isStreaming.value = false
    showToast('请求异常: ' + err)
  }
}

async function stopGenerationAction() {
  await wailsBridge.cancelAgentStream()
  isStreaming.value = false
  if (currentSession.value && currentSession.value.id) {
    try {
      await wailsBridge.saveSession(currentSession.value)
    } catch (_) {}
  }
  showToast('已中断本次推理并保存当前内容')
}

// 6. 设置中枢 (渠道、MCP、Skill、Rule)
const channels = ref<ChannelConfig[]>([])
const mcps = ref<MCPServerConfig[]>([])
const skills = ref<SkillConfig[]>([])
const rules = ref<RuleConfig[]>([])
const pingLoadingMap = reactive<Record<string, boolean>>({})

const channelForm = reactive({
  name: '',
  endpoint: '',
  api_key: ''
})

async function loadSettingsData() {
  channels.value = await wailsBridge.listChannels()
  mcps.value = await wailsBridge.listMCPs()
  skills.value = await wailsBridge.listSkills()
  rules.value = await wailsBridge.listRules()

  const primary = channels.value.find(c => c.primary)
  if (primary && primary.model) {
    selectedModel.value = primary.model
  }
}

function openSettingsTab(tab: string) {
  activeSettingsTab.value = tab
  isSettingsOpen.value = true
}

async function executePing(id: string) {
  pingLoadingMap[id] = true
  try {
    const latency = await wailsBridge.pingChannel(id)
    const target = channels.value.find(c => c.id === id)
    if (target) target.latency = latency
    showToast(`✓ 渠道真实网络往返延迟: ${latency}`)
  } catch (err) {
    showToast('测速失败: ' + err)
  } finally {
    pingLoadingMap[id] = false
  }
}

async function pingAllChannels() {
  for (const ch of channels.value) {
    await executePing(ch.id)
  }
}

function setPrimaryChannel(id: string) {
  channels.value.forEach(c => c.primary = (c.id === id))
  const cur = channels.value.find(c => c.id === id)
  if (cur) wailsBridge.saveChannel(cur)
}

function openAddChannelModal() {
  channelForm.name = ''
  channelForm.endpoint = ''
  channelForm.api_key = ''
  isChannelModalOpen.value = true
}

function editChannel(ch: ChannelConfig) {
  channelForm.name = ch.name
  channelForm.endpoint = ch.endpoint
  channelForm.api_key = ch.api_key || ''
  isChannelModalOpen.value = true
}

async function deleteChannel(id: string) {
  await wailsBridge.deleteChannel(id)
  channels.value = channels.value.filter(c => c.id !== id)
  showToast('✓ 渠道已移除')
}

async function fetchModelsAction() {
  try {
    let ep = channelForm.endpoint.trim()
    if (ep && !ep.startsWith('http://') && !ep.startsWith('https://')) {
      ep = (ep.includes('localhost') || ep.includes('127.0.0.1')) ? 'http://' + ep : 'https://' + ep
      channelForm.endpoint = ep
    }
    const models = await wailsBridge.fetchUpstreamModels(channelForm.endpoint, channelForm.api_key)
    if (models && models.length > 0) {
      upstreamFetchedModels.value = models
      showToast(`✓ 成功从上游网关探测到 ${models.length} 个真实在线模型！`)
    } else {
      showToast('未探测到可用模型列表')
    }
  } catch (err) {
    showToast('拉取模型异常: ' + err)
  }
}

async function saveChannelAction() {
  await wailsBridge.saveChannel({
    id: 'ch_' + Date.now(),
    name: channelForm.name,
    primary: false,
    status: 'online',
    auth_type: 'bearer_token',
    endpoint: channelForm.endpoint,
    api_key: channelForm.api_key,
    model: selectedModel.value,
    latency: '未测速',
    updated_at: Date.now()
  })
  isChannelModalOpen.value = false
  await loadSettingsData()
  showToast('✓ 渠道配置已真实保存至 ~/.tcode/channels.json')
}

async function toggleMcp(mcp: MCPServerConfig) {
  await wailsBridge.saveMCP(mcp)
}

async function toggleSkill(skill: SkillConfig) {
  await wailsBridge.saveSkill(skill)
}

async function toggleRule(rule: RuleConfig) {
  await wailsBridge.saveRule(rule)
}

const mcpForm = reactive({
  name: '',
  type: 'stdio',
  command: '',
  args: [] as string[]
})
const mcpArgsInput = ref('')

const skillForm = reactive({
  name: '',
  description: '',
  content: ''
})

const ruleForm = reactive({
  name: '',
  content: ''
})

async function saveMcpAction() {
  if (!mcpForm.name.trim() || !mcpForm.command.trim()) {
    showToast('请完整填写 MCP 服务名称与启动命令')
    return
  }
  const args = mcpArgsInput.value.trim() ? mcpArgsInput.value.trim().split(/\s+/) : []
  await wailsBridge.saveMCP({
    id: 'mcp_' + Date.now(),
    name: mcpForm.name.trim(),
    type: mcpForm.type,
    command: mcpForm.command.trim(),
    args: args,
    enabled: true,
    updated_at: Date.now()
  })
  isMcpModalOpen.value = false
  mcpForm.name = ''
  mcpForm.command = ''
  mcpArgsInput.value = ''
  await loadSettingsData()
  showToast('✓ MCP 服务已成功注册并保存')
}

async function saveSkillAction() {
  if (!skillForm.name.trim()) {
    showToast('请填写技能名称')
    return
  }
  await wailsBridge.saveSkill({
    id: 'skill_' + Date.now(),
    name: skillForm.name.trim(),
    description: skillForm.description.trim(),
    prompt: skillForm.content.trim(),
    enabled: true,
    updated_at: Date.now()
  })
  isSkillModalOpen.value = false
  skillForm.name = ''
  skillForm.description = ''
  skillForm.content = ''
  await loadSettingsData()
  showToast('✓ 技能已成功添加至本地技能库')
}

async function deleteSkillAction(id: string) {
  await wailsBridge.deleteSkill(id)
  await loadSettingsData()
  showToast('✓ 技能已从本地技能库移除')
}

async function saveRuleAction() {
  if (!ruleForm.name.trim() || !ruleForm.content.trim()) {
    showToast('请完整填写规则名称与规则内容')
    return
  }
  await wailsBridge.saveRule({
    id: 'rule_' + Date.now(),
    name: ruleForm.name.trim(),
    content: ruleForm.content.trim(),
    enabled: true,
    updated_at: Date.now()
  })
  isRuleModalOpen.value = false
  ruleForm.name = ''
  ruleForm.content = ''
  await loadSettingsData()
  showToast('✓ 工程规约已成功添加')
}

async function deleteRuleAction(id: string) {
  await wailsBridge.deleteRule(id)
  await loadSettingsData()
  showToast('✓ 工程规约已删除')
}

// 7. 真实 AST 代码拓扑知识图谱
const astNodes = ref<GraphNode[]>([])
const selectedAstNode = ref<GraphNode | null>(null)

function openKnowledgeGraphModal() {
  isKnowledgeGraphOpen.value = true
  if (astNodes.value.length === 0 && !isGraphLoading.value) {
    scanASTGraph()
  }
}

async function scanASTGraph() {
  isGraphLoading.value = true
  try {
    const nodes = await wailsBridge.getProjectASTGraph()
    astNodes.value = nodes || []
    if (astNodes.value.length > 0) selectedAstNode.value = astNodes.value[0]
  } catch (err) {
    showToast('AST 扫描失败: ' + err)
  } finally {
    isGraphLoading.value = false
  }
}

function injectNodeToPrompt() {
  if (!selectedAstNode.value) return
  const node = selectedAstNode.value
  const quoteText = `\n> 架构拓扑实体引用: \`${node.name}\` [${node.type}]\n> 声明路径: \`${node.file}\`\n> 关联说明: ${node.details}\n`
  inputPrompt.value = inputPrompt.value ? inputPrompt.value + quoteText : quoteText
  isKnowledgeGraphOpen.value = false
  showToast(`✓ 已引用 AST 节点 [${node.name}] 架构约束至输入框`)
}

// =========================================================================
// 6. 底部集成式可折叠流式终端抽屉 (Terminal Drawer)
// =========================================================================
interface TerminalOutputItem {
  type: 'cmd' | 'output' | 'exit'
  text?: string
  exitCode?: number
  durationMs?: number
}

const isTerminalOpen = ref(false)
const isTerminalMaximized = ref(false)
const terminalHeight = ref(240)
const activeTerminalTab = ref<'shell' | 'logs'>('shell')
const isTerminalRunning = ref(false)
const terminalInputCmd = ref('')
const currentTerminalBuffer = ref('')
const terminalOutputs = ref<TerminalOutputItem[]>([])
const commandHistory = ref<string[]>([])
const historyIndex = ref(-1)
const terminalScrollRef = ref<HTMLDivElement | null>(null)

const agentTraceLogs = ref<{ time: string; phase: string; message: string }[]>([])

function toggleTerminalDrawer(forceState?: boolean) {
  isTerminalOpen.value = forceState !== undefined ? forceState : !isTerminalOpen.value
  if (isTerminalOpen.value) {
    scrollToBottomTerminal()
  }
}

function clearTerminalLogs() {
  terminalOutputs.value = []
  currentTerminalBuffer.value = ''
}

function scrollToBottomTerminal() {
  nextTick(() => {
    if (terminalScrollRef.value) {
      terminalScrollRef.value.scrollTop = terminalScrollRef.value.scrollHeight
    }
  })
}

function navigateCommandHistory(direction: number) {
  if (commandHistory.value.length === 0) return
  if (historyIndex.value === -1) {
    historyIndex.value = commandHistory.value.length
  }
  historyIndex.value += direction
  if (historyIndex.value < 0) {
    historyIndex.value = 0
  } else if (historyIndex.value >= commandHistory.value.length) {
    historyIndex.value = commandHistory.value.length
    terminalInputCmd.value = ''
    return
  }
  terminalInputCmd.value = commandHistory.value[historyIndex.value] || ''
}

async function submitTerminalCommand() {
  const cmd = terminalInputCmd.value.trim()
  if (!cmd || isTerminalRunning.value) return

  if (cmd === 'clear' || cmd === 'cls') {
    clearTerminalLogs()
    terminalInputCmd.value = ''
    return
  }

  if (!commandHistory.value.includes(cmd)) {
    commandHistory.value.push(cmd)
  }
  historyIndex.value = -1

  terminalOutputs.value.push({ type: 'cmd', text: cmd })
  terminalInputCmd.value = ''
  currentTerminalBuffer.value = ''
  isTerminalRunning.value = true
  scrollToBottomTerminal()

  try {
    await wailsBridge.execTerminalStream(cmd, {
      onData: (chunk: string) => {
        currentTerminalBuffer.value += chunk
        scrollToBottomTerminal()
      },
      onExit: (data) => {
        if (currentTerminalBuffer.value) {
          terminalOutputs.value.push({ type: 'output', text: currentTerminalBuffer.value })
          currentTerminalBuffer.value = ''
        }
        terminalOutputs.value.push({
          type: 'exit',
          exitCode: data.exit_code,
          durationMs: data.duration_ms
        })
        isTerminalRunning.value = false
        scrollToBottomTerminal()
      }
    })
  } catch (err) {
    terminalOutputs.value.push({ type: 'output', text: `[Execution Error]: ${err}` })
    isTerminalRunning.value = false
    scrollToBottomTerminal()
  }
}

async function cancelTerminalAction() {
  try {
    await wailsBridge.cancelTerminalCommand()
    isTerminalRunning.value = false
  } catch (err) {
    console.error('Cancel terminal error:', err)
  }
}

function handleGlobalKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    if (isMcpModalOpen.value) { isMcpModalOpen.value = false; return }
    if (isSkillModalOpen.value) { isSkillModalOpen.value = false; return }
    if (isRuleModalOpen.value) { isRuleModalOpen.value = false; return }
    if (isKnowledgeGraphOpen.value) { isKnowledgeGraphOpen.value = false; return }
    if (isSettingsOpen.value) { isSettingsOpen.value = false; return }
    if (isTerminalOpen.value) { isTerminalOpen.value = false; return }
  }

  if (e.ctrlKey && (e.key === '`' || e.key === '~')) {
    e.preventDefault()
    toggleTerminalDrawer()
  }
}

function initWorkbench() {
  window.addEventListener('keydown', handleGlobalKeydown)
  loadSessionsList()
  loadSettingsData()
  wailsBridge.getWorkspace().then(ws => {
    if (ws) {
      workspacePath.value = ws
      loadFileTree()
      loadGitStatus()
    }
  })
  return () => window.removeEventListener('keydown', handleGlobalKeydown)
}


  return {
    activeActivity,
    activeDiffFile,
    activeSettingsTab,
    activeTag,
    activeTerminalTab,
    agentTraceLogs,
    applyHunkAction,
    astNodes,
    attachedFiles,
    availableModels,
    availableTags,
    cancelTerminalAction,
    channelForm,
    channels,
    chooseWorkspace,
    clearTerminalLogs,
    commandHistory,
    commitMessage,
    createNewSession,
    currentSession,
    currentSessionId,
    currentTerminalBuffer,
    deleteChannel,
    deleteRuleAction,
    deleteSession,
    deleteSkillAction,
    diffReport,
    discardHunkAction,
    editChannel,
    executePing,
    expandedFolders,
    fetchModelsAction,
    fileTree,
    filteredSessions,
    gitStatus,
    handleFileClick,
    handleGitCommit,
    handleGlobalKeydown,
    handleSend,
    hiddenHistoryCount,
    revealFullHistory,
    historyIndex,
    initWorkbench,
    injectNodeToPrompt,
    inputPrompt,
    isChannelModalOpen,
    isDiffOpen,
    isFileTreeLoading,
    isFullAuto,
    isGitLoading,
    isGraphLoading,
    isKnowledgeGraphOpen,
    isMcpModalOpen,
    isRuleModalOpen,
    isSettingsOpen,
    isSkillModalOpen,
    isStreaming,
    isTerminalMaximized,
    isTerminalOpen,
    isTerminalRunning,
    loadDiff,
    loadFileTree,
    loadGitStatus,
    loadSessionsList,
    loadSettingsData,
    mcpArgsInput,
    mcpForm,
    mcps,
    messagesContainerRef,
    navigateCommandHistory,
    openAddChannelModal,
    openFileDiff,
    openKnowledgeGraphModal,
    openSettingsTab,
    pingAllChannels,
    pingLoadingMap,
    renderMarkdown,
    revertFileAction,
    ruleForm,
    rules,
    saveChannelAction,
    saveMcpAction,
    saveRuleAction,
    saveSkillAction,
    scanASTGraph,
    scrollToBottomTerminal,
    selectSession,
    selectedAstNode,
    selectedModel,
    sessions,
    setPrimaryChannel,
    showToast,
    skillForm,
    skills,
    stageFileAction,
    stagedTreeFiles,
    stopGenerationAction,
    submitTerminalCommand,
    switchToFileActivity,
    switchToGitActivity,
    terminalHeight,
    terminalInputCmd,
    terminalOutputs,
    terminalScrollRef,
    toastMessage,
    toggleMcp,
    toggleRule,
    toggleSkill,
    toggleTerminalDrawer,
    triggerUpload,
    unstageFileAction,
    visibleMessages,
    upstreamFetchedModels,
    workingTreeFiles,
    workspaceName,
    workspacePath
  }
})

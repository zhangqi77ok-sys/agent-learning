<template>
<main class="flex-1 min-h-0 bg-[#FAF8F5] flex flex-col justify-between overflow-hidden relative font-sans" @dragover.prevent @drop.prevent="s.onChatDrop($event)">
        <!-- 顶栏: 场景标签、多模型切换器与收起代码按钮 -->
        <header class="h-10 min-h-[40px] bg-[#FAF8F5] border-b border-black/[0.08] px-2 flex items-center justify-between text-xs select-none z-10 shrink-0 gap-2">
          <div class="flex items-center gap-1 min-w-0 flex-1 overflow-x-auto no-scrollbar">
            <button
              v-for="tab in s.sessionTabs"
              :key="tab.id"
              draggable="true"
              @dragstart="s.onTabDragStart($event, tab.id)"
              @dragover.prevent
              @drop="s.onTabDrop($event, tab.id)"
              @contextmenu="s.openTabMenu($event, tab.id)"
              @click="s.selectSession(tab.id)"
              :class="[
                'flex items-center gap-1 px-2 py-1 rounded-md shrink-0 max-w-[160px] cursor-pointer',
                s.currentSessionId === tab.id ? 'bg-white border border-black/[0.1] font-semibold text-[#18181B]' : 'text-[#71717A] hover:bg-black/[0.04]'
              ]"
            >
              <span class="truncate">{{ tab.title }}</span>
              <span class="text-[#A1A1AA] hover:text-red-500" @click.stop="s.closeSessionTab(tab.id)">✕</span>
            </button>
            <span v-if="s.sessionTabs.length === 0" class="font-bold text-[#18181B] px-1 truncate">{{ s.currentSession.title }}</span>
            <span class="text-[10px] text-[#71717A] bg-black/[0.04] px-1.5 py-0.2 rounded font-mono">{{ s.selectedModel || '未配置渠道' }}</span>

            <select
              v-if="s.availableModels.length > 0"
              v-model="s.selectedModel"
              class="bg-white border border-black/[0.1] rounded-lg px-2 py-0.8 text-xs font-mono font-medium text-[#10A37F] focus:outline-none focus:border-[#D96B27] cursor-pointer shadow-2xs"
            >
              <option v-for="m in s.availableModels" :key="m" :value="m">{{ m }}</option>
            </select>
            <button
              v-else
              type="button"
              class="text-[10px] text-[#D96B27] underline cursor-pointer"
              @click="s.openSettingsTab('models')"
            >添加渠道</button>
          </div>

          <button
            @click="s.setWorkspaceView(s.isDiffOpen ? 'chat' : 'split')"
            class="flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-white border border-black/[0.08] text-xs text-[#52525B] hover:text-[#18181B] hover:bg-black/[0.02] shadow-2xs transition-all cursor-pointer"
          >
            <span>{{ s.isDiffOpen ? '收起代码面板' : '💻 代码面板' }}</span>
          </button>
        </header>

        <!-- 真实动态对话消息列表：跟最新输出，也可拖拽/滚轮回看 -->
        <div class="flex-1 min-h-0 relative">
        <div
          ref="messagesContainerRef"
          class="h-full overflow-y-scroll overflow-x-hidden p-4 space-y-4 flex flex-col select-text overscroll-contain"
          :class="dragging ? 'cursor-grabbing' : 'cursor-grab'"
          @scroll="s.onMessagesScroll"
          @mousedown="onTranscriptDown"
          @mousemove="onTranscriptMove"
          @mouseup="onTranscriptUp"
          @mouseleave="onTranscriptUp"
        >
          <!-- 干净真实的空会话状态 -->
          <div v-if="!s.currentSession.messages || s.currentSession.messages.length === 0" class="flex-1 flex flex-col items-center justify-center text-center p-8 select-none my-auto">
            <div class="w-14 h-14 rounded-2xl bg-white border border-black/[0.08] shadow-xs flex items-center justify-center text-2xl mb-4">
              💬
            </div>
            <h3 class="text-sm font-bold text-[#18181B] mb-1.5">湉码 / tiancode</h3>
            <p class="text-xs text-[#71717A] max-w-sm mb-5 leading-relaxed">
              这是一条尚未保存的新对话。输入任务并发送后，会按当前项目记入左侧列表。
            </p>
            <div class="flex items-center gap-2">
              <button
                @click="s.createNewSession"
                class="px-3.5 py-1.5 rounded-xl bg-[#D96B27] text-white text-xs font-semibold shadow-xs hover:bg-[#B8551B] transition-all cursor-pointer flex items-center gap-1"
              >
                <span>＋</span><span>新建会话</span>
              </button>
            </div>
          </div>

          <button
            v-if="s.hiddenHistoryCount > 0"
            type="button"
            class="self-center text-[11px] text-[#D96B27] underline cursor-pointer"
            @click="s.revealFullHistory"
          >显示更早的 {{ s.hiddenHistoryCount }} 条（默认只渲染最近 80 条）</button>
          <template v-for="msg in s.visibleMessages" :key="msg.id">
            <!-- 用户提问气泡 -->
            <div v-if="msg.role === 'user'" class="flex justify-end">
              <div class="max-w-[80%] bg-[#F4EFEA] text-[#18181B] px-4 py-3 rounded-2xl rounded-tr-sm border border-black/[0.06] shadow-2xs text-xs leading-relaxed whitespace-pre-line">
                {{ msg.content }}
              </div>
            </div>

            <!-- Agent 回答卡片组 -->
            <div v-else class="flex flex-col items-start space-y-3.5 max-w-3xl w-full">
              <div class="flex items-center gap-2 text-xs font-semibold text-[#18181B]">
                <div class="w-4 h-4 rounded bg-[#D96B27] text-white flex items-center justify-center text-[9px] font-bold">T</div>
                <span>湉码 Agent</span>
                <span class="text-[10px] text-[#10A37F] bg-[#10A37F]/10 px-1.5 py-0.2 rounded font-mono">{{ s.selectedModel }} · 自主算子模式</span>
              </div>

              <!-- 深度思考抽屉 (真实 reasoning_content) -->
              <div v-if="msg.thinking" class="w-full rounded-xl border border-black/[0.08] bg-white/70 shadow-2xs overflow-hidden">
                <div class="p-2.5 flex items-center justify-between bg-black/[0.02] text-xs font-semibold text-[#18181B]">
                  <div class="flex items-center gap-2">
                    <span>🧠</span><span>深度心智思考 (Reasoning Process)</span>
                  </div>
                  <span class="text-[10px] text-[#10A37F] font-mono">Token 流</span>
                </div>
                <div class="px-3 pb-3 text-xs text-[#71717A] leading-relaxed italic border-t border-black/[0.04] pt-2 whitespace-pre-wrap font-mono">
                  {{ msg.thinking }}
                </div>
              </div>

              <!-- Tool Call 算子执行卡片列表 (多轮自主执行时序链路) -->
              <div v-if="(msg.tools && msg.tools.length > 0) || msg.tool" class="w-full space-y-2">
                <div
                  v-for="(tItem, tIdx) in (msg.tools && msg.tools.length > 0 ? msg.tools : [msg.tool!])"
                  :key="tItem.id || tIdx"
                  class="rounded-xl border border-black/[0.08] bg-white shadow-2xs overflow-hidden"
                >
                  <div class="p-2 flex items-center justify-between bg-black/[0.02] text-xs font-mono">
                    <span class="font-bold text-[#18181B]">$_ {{ tItem.name }} {{ typeof tItem.args === 'string' ? tItem.args : JSON.stringify(tItem.args) }}</span>
                    <span class="text-[10px]" :class="(tItem.output || '').startsWith('[') || (tItem.output || '').includes('error') || (tItem.output || '').includes('拦截') ? 'text-red-500' : ((tItem.output === '正在执行...' || !tItem.output) ? 'text-amber-600' : 'text-[#10A37F]')">
                      {{ (tItem.output || '').startsWith('[') || (tItem.output || '').includes('拦截') ? '● 已拦截/失败' : ((tItem.output === '正在执行...' || !tItem.output) ? '● 执行中' : '● 完成') }}
                    </span>
                  </div>
                  <div class="p-2.5 bg-[#18181B] text-emerald-400 font-mono text-[11px] whitespace-pre-wrap">
                    {{ tItem.output }}
                  </div>
                </div>
              </div>


              <!-- 优雅 Markdown 正文 -->
              <div
                class="markdown-body text-xs text-[#27272A] leading-relaxed space-y-2 bg-white/70 p-3.5 rounded-xl border border-black/[0.04] w-full"
                v-html="s.renderMarkdown(msg.content)"
              ></div>
              <div class="flex items-center gap-2 text-[10px] text-[#A1A1AA]">
                <button class="hover:text-[#18181B] cursor-pointer" @click="s.copyMessage(msg.content)">复制</button>
                <button class="hover:text-[#18181B] cursor-pointer" @click="s.regenerateLast">重新生成</button>
              </div>
            </div>
          </template>

        </div>
        <button
          v-if="!s.stickToBottom"
          type="button"
          class="absolute bottom-3 left-1/2 -translate-x-1/2 z-10 px-3 py-1.5 rounded-full bg-[#18181B] text-white text-[11px] font-medium shadow-xs cursor-pointer"
          @click="s.followLatestChat"
        >↓ 回到最新</button>
        </div>

        <!-- 底部输入胶囊舱 (Prompt Capsule) -->
        <div class="p-3 bg-[#FAF8F5] border-t border-black/[0.06] select-none relative">
          <!-- 真实附件预览托盘 -->
          <div v-if="s.attachedFiles.length" class="flex items-center gap-1.5 pb-2 overflow-x-auto no-scrollbar">
            <div
              v-for="(file, idx) in s.attachedFiles"
              :key="idx"
              class="flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-white border border-black/[0.08] text-xs font-mono shadow-2xs text-[#18181B]"
            >
              <span class="text-[#D96B27]">📎</span>
              <span class="truncate max-w-xs">{{ file }}</span>
              <button @click="s.attachedFiles.splice(idx, 1)" class="text-[#71717A] hover:text-red-500 cursor-pointer">✕</button>
            </div>
          </div>

          <!-- 输入卡片 -->
          <div class="rounded-2xl bg-white border border-black/[0.12] shadow-sm focus-within:border-[#D96B27] focus-within:ring-2 focus-within:ring-[#D96B27]/15 transition-all p-2.5 flex flex-col gap-2">
            <textarea
              v-model="s.inputPrompt"
              rows="2"
              placeholder="给 湉码 Agent 发送指令（@ 引用会话/技能/文件，/ 调起指令，Shift+Enter 换行，拖入文件作为附件）"
              class="w-full text-xs text-[#18181B] placeholder-[#A1A1AA] bg-transparent focus:outline-none resize-none leading-relaxed"
              @keydown="s.handleComposerKeydown"
            ></textarea>
            <div
              v-if="s.mentionOpen && s.mentionItems.length > 0"
              class="absolute left-4 right-4 bottom-[7.5rem] z-20 bg-white border border-black/[0.1] rounded-xl shadow-lg max-h-48 overflow-y-auto"
            >
              <button
                v-for="(item, idx) in s.mentionItems"
                :key="item.id"
                class="w-full text-left px-3 py-1.5 text-xs flex justify-between cursor-pointer"
                :class="idx === s.mentionIndex ? 'bg-[#D96B27]/10' : 'hover:bg-black/[0.03]'"
                @mousedown.prevent="s.applyMention(item)"
              >
                <span>{{ item.label }}</span>
                <span class="text-[10px] text-[#A1A1AA]">{{ item.kind }}</span>
              </button>
            </div>

            <div class="flex items-center justify-between border-t border-black/[0.04] pt-2 text-xs">
              <div class="flex items-center gap-1">
                <button
                  @click="s.triggerUpload"
                  class="px-2.5 py-1 rounded-full text-xs text-[#52525B] hover:text-[#18181B] hover:bg-black/[0.04] flex items-center gap-1 cursor-pointer"
                  title="调起系统文件选择框"
                >
                  <span>📎</span><span>上传</span>
                </button>
                <div class="h-3.5 w-px bg-black/[0.1] mx-1"></div>

                <div class="flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-[#D96B27]/10 text-[#D96B27] text-xs font-semibold select-none" :title="'当前策略: ' + s.executionStrategy">
                  <span>⚡</span><span>{{ s.executionStrategy === 'analyze' ? '只读分析' : s.executionStrategy === 'tdd' ? 'TDD 闭环' : '直接改代码' }}</span>
                </div>
              </div>

              <div class="flex items-center gap-2">
                <span class="text-[10px] text-[#A1A1AA] font-mono">{{ s.isStreaming ? '正在流式推理...' : '就绪' }}</span>
                <button
                  v-if="s.isStreaming"
                  @click="s.stopGenerationAction"
                  title="中断本次生成 (Esc)"
                  class="w-7 h-7 rounded-xl flex items-center justify-center font-bold shadow-xs transition-all cursor-pointer bg-red-500 hover:bg-red-600 text-white animate-pulse"
                >
                  ■
                </button>
                <button
                  v-else
                  @click="s.handleSend"
                  title="发送消息 (Enter)"
                  class="w-7 h-7 rounded-xl flex items-center justify-center font-bold shadow-xs transition-all cursor-pointer bg-[#D96B27] hover:bg-[#B8551B] text-white"
                >
                  ↑
                </button>
              </div>
            </div>
          </div>
        </div>
      </main>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useWorkbenchStore } from '../stores/workbench'
const s = useWorkbenchStore()
const { messagesContainerRef } = storeToRefs(s)

const dragging = ref(false)
let dragY = 0
let dragStartScroll = 0
let moved = false

function onTranscriptDown(e: MouseEvent) {
  if (e.button !== 0) return
  const t = e.target as HTMLElement
  if (t.closest('button, a, textarea, input, pre, code')) return
  dragging.value = true
  dragY = e.clientY
  dragStartScroll = messagesContainerRef.value?.scrollTop || 0
  moved = false
}

function onTranscriptMove(e: MouseEvent) {
  if (!dragging.value || !messagesContainerRef.value) return
  const dy = e.clientY - dragY
  if (Math.abs(dy) > 3) moved = true
  if (moved) {
    e.preventDefault()
    messagesContainerRef.value.scrollTop = dragStartScroll - dy
    s.onMessagesScroll()
  }
}

function onTranscriptUp() {
  dragging.value = false
}
</script>


<style>
.markdown-body pre { background-color:#18181B; color:#F4F4F5; padding:0.75rem; border-radius:0.5rem; overflow-x:auto; font-family:'Fira Code',monospace; margin:0.5rem 0; border:1px solid rgba(255,255,255,0.1); }
.markdown-body code { font-family:'Fira Code',monospace; background-color:rgba(0,0,0,0.05); padding:0.1rem 0.3rem; border-radius:0.25rem; }
.markdown-body pre code { background-color:transparent; padding:0; }
.markdown-body p { margin-bottom:0.5rem; }
.markdown-body ul, .markdown-body ol { padding-left:1.25rem; margin-bottom:0.5rem; }
.markdown-body ul { list-style-type:disc; }
.markdown-body ol { list-style-type:decimal; }
</style>

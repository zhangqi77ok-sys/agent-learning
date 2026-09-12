<template>
<main class="flex-1 bg-[#FAF8F5] flex flex-col justify-between overflow-hidden relative font-sans">
        <!-- 顶栏: 场景标签、多模型切换器与收起代码按钮 -->
        <header class="h-10 min-h-[40px] bg-[#FAF8F5] border-b border-black/[0.08] px-3 flex items-center justify-between text-xs select-none z-10 shrink-0">
          <div class="flex items-center gap-2">
            <span class="font-bold text-[#18181B]">{{ s.currentSession.title }}</span>
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
            @click="s.isDiffOpen = !s.isDiffOpen"
            class="flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-white border border-black/[0.08] text-xs text-[#52525B] hover:text-[#18181B] hover:bg-black/[0.02] shadow-2xs transition-all cursor-pointer"
          >
            <span>{{ s.isDiffOpen ? '收起代码' : '展开代码' }}</span>
          </button>
        </header>

        <!-- 真实动态对话消息列表 (从当前 Session 动态读取与渲染) -->
        <div ref="messagesContainerRef" class="flex-1 overflow-y-auto p-4 space-y-4 flex flex-col">
          <!-- 干净真实的空会话状态 -->
          <div v-if="!s.currentSession.messages || s.currentSession.messages.length === 0" class="flex-1 flex flex-col items-center justify-center text-center p-8 select-none my-auto">
            <div class="w-14 h-14 rounded-2xl bg-white border border-black/[0.08] shadow-xs flex items-center justify-center text-2xl mb-4">
              💬
            </div>
            <h3 class="text-sm font-bold text-[#18181B] mb-1.5">湉码 / tiancode</h3>
            <p class="text-xs text-[#71717A] max-w-sm mb-5 leading-relaxed">
              当前暂无活跃对话。请在下方输入框键入编程任务，或点击【＋新建会话】开始。
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
                    <span class="text-[10px] text-[#10A37F]">● 执行成功</span>
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
            </div>
          </template>

          <div v-if="s.currentSession.messages.length === 0" class="h-64 flex flex-col items-center justify-center text-center text-xs text-[#71717A] space-y-2">
            <span class="text-3xl">💬</span>
            <div class="font-semibold text-sm text-[#18181B]">新会话已创建就绪</div>
            <div>请输入编程需求或直接拖拽文件，开始自主智能体编程之旅。</div>
          </div>
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
              placeholder="给 湉码 Agent 发送指令 (支持拖拽文件，输入 @ 引用工程，/ 调起算子)..."
              class="w-full text-xs text-[#18181B] placeholder-[#A1A1AA] bg-transparent focus:outline-none resize-none leading-relaxed"
              @keydown.enter.prevent="s.handleSend"
            ></textarea>

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

                <div class="flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-[#D96B27]/10 text-[#D96B27] text-xs font-semibold select-none">
                  <span>⚡</span><span>Act 极速双环</span>
                </div>

                <button
                  @click="s.isFullAuto = !s.isFullAuto"
                  :class="[
                    'flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium border transition-all cursor-pointer shadow-2xs',
                    s.isFullAuto
                      ? 'bg-emerald-50 text-emerald-800 border-emerald-300 ring-2 ring-emerald-400/20'
                      : 'bg-white text-[#52525B] border-black/[0.08] hover:border-black/[0.18]'
                  ]"
                >
                  <span :class="['w-2 h-2 rounded-full', s.isFullAuto ? 'bg-emerald-500 animate-pulse' : 'bg-amber-500']"></span>
                  <span>{{ s.isFullAuto ? '⚡ 全自动执行 (免审核)' : '需人工审核' }}</span>
                </button>
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
import { storeToRefs } from 'pinia'
import { useWorkbenchStore } from '../stores/workbench'
const s = useWorkbenchStore()
const { messagesContainerRef } = storeToRefs(s)
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

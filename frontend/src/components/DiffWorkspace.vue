<template>
<!-- 右侧 Monaco Diff 审查工作区 (Diff Workspace) -->
      <section
        v-show="s.isDiffOpen"
        class="w-[45vw] min-w-[380px] max-w-[700px] border-l border-black/[0.08] bg-[#FAF8F5] flex flex-col justify-between select-none z-10 shrink-0 font-sans"
      >
        <header class="h-10 min-h-[40px] bg-[#FAF8F5] border-b border-black/[0.08] px-3 flex items-center justify-between text-xs shrink-0">
          <div class="flex items-center gap-2 min-w-0">
            <span class="text-sm">📄</span>
            <span class="font-mono font-bold text-[#18181B] truncate">{{ s.activeDiffFile }}</span>
            <span class="text-[10px] font-mono text-[#10A37F] bg-[#10A37F]/10 px-1.5 py-0.2 rounded font-bold shrink-0">
              {{ s.diffReport?.stats || '0 行修改' }}
            </span>
          </div>

          <div class="flex items-center gap-1.5 shrink-0">
            <button
              @click="s.editorView = 'edit'"
              :class="['px-2 py-0.5 rounded-md text-[10px] cursor-pointer', s.editorView === 'edit' ? 'bg-[#18181B] text-white' : 'bg-white border border-black/[0.08]']"
            >编辑</button>
            <button
              @click="s.editorView = 'diff'"
              :class="['px-2 py-0.5 rounded-md text-[10px] cursor-pointer', s.editorView === 'diff' ? 'bg-[#18181B] text-white' : 'bg-white border border-black/[0.08]']"
            >Diff</button>
            <button
              v-if="s.editorView === 'edit'"
              @click="s.saveEditor"
              :disabled="!s.editorDirty"
              class="px-2 py-0.5 rounded-md bg-[#D96B27] text-white text-[10px] font-semibold disabled:opacity-40 cursor-pointer"
              title="写入磁盘 (Ctrl+S)"
            >保存</button>
            <button @click="s.loadDiff" class="p-1 rounded-md text-[#71717A] hover:bg-black/[0.04] cursor-pointer" title="刷新代码差异">🔄</button>
            <button
              @click="s.revertFileAction"
              class="flex items-center gap-1 px-2 py-0.8 rounded-md bg-white border border-red-200 text-xs font-semibold text-red-600 hover:bg-red-50 shadow-2xs transition-all cursor-pointer"
              title="丢弃本次物理改动 (Git Checkout)"
            >
              <span>✕</span><span>放弃</span>
            </button>
            <button
              @click="s.stageFileAction"
              class="flex items-center gap-1 px-2.5 py-0.8 rounded-md bg-[#10A37F] hover:bg-[#0D8C6D] text-white text-xs font-semibold shadow-xs transition-all cursor-pointer"
              title="确认采纳文件修改并提交暂存区 (Git Stage)"
            >
              <span>✓</span><span>采纳变更</span>
            </button>
            <button @click="s.isDiffOpen = false" class="text-[#71717A] hover:text-[#18181B] p-1 rounded-md hover:bg-black/[0.05] cursor-pointer ml-1">✕</button>
          </div>
        </header>

        <div v-if="s.editorView === 'edit'" class="flex-1 min-h-0 bg-[#1e1e1e]">
          <MonacoEditor
            v-if="s.activeDiffFile"
            v-model="s.editorContent"
            :language="s.activeDiffFile"
            @update:modelValue="s.markEditorDirty"
          />
          <div v-else class="h-full flex items-center justify-center text-xs text-[#A1A1AA]">从左侧文件树打开文件即可编辑</div>
        </div>

        <!-- 真实物理行级 Diff (Red / Green) -->
        <div v-else class="flex-1 overflow-y-auto bg-[#18181B] text-[#F4F4F5] font-mono text-[11px] p-2 space-y-2 select-text flex flex-col">
          <div v-if="!s.activeDiffFile || !s.diffReport?.lines || s.diffReport.lines.length === 0" class="flex-1 flex flex-col items-center justify-center p-8 text-center text-[#71717A] my-auto">
            <span class="text-3xl mb-3">📄</span>
            <p class="text-xs font-semibold text-[#A1A1AA]">暂无代码差异对比</p>
            <p class="text-[11px] text-[#71717A] mt-1.5 max-w-xs leading-relaxed">
              当前工作区干净，或尚未选定对比文件。可从左侧文件树或 Git 状态点击文件审查。
            </p>
          </div>

          <!-- 分块 Hunks 细粒度审查模式 -->
          <template v-if="s.diffReport?.hunks && s.diffReport.hunks.length > 0">
            <div
              v-for="(hunk, hIdx) in s.diffReport.hunks"
              :key="hIdx"
              class="p-2.5 rounded-xl bg-black/40 border border-white/[0.08] select-none"
            >
              <div class="flex items-center justify-between pb-1.5 mb-1.5 border-b border-white/[0.06] text-[10px]">
                <div class="flex items-center gap-1.5 font-mono min-w-0">
                  <span class="text-[#D96B27] font-bold shrink-0">块 #{{ hIdx + 1 }}</span>
                  <span class="text-white/40 truncate">{{ hunk.header }}</span>
                  <span v-if="hunk.add_count > 0" class="text-emerald-400 font-bold shrink-0">+{{ hunk.add_count }}</span>
                  <span v-if="hunk.del_count > 0" class="text-rose-400 font-bold shrink-0">-{{ hunk.del_count }}</span>
                </div>
                <div class="flex items-center gap-1.5 shrink-0">
                  <button
                    @click="s.applyHunkAction(hunk.index, true)"
                    title="将此块代码改动暂存入 Git Index (git apply --cached)"
                    class="px-2 py-0.5 rounded bg-[#10A37F]/20 hover:bg-[#10A37F]/30 text-[#10A37F] font-bold text-[10px] cursor-pointer transition-all active:scale-95"
                  >
                    ✓ 采纳块
                  </button>
                  <button
                    @click="s.discardHunkAction(hunk.index)"
                    title="无损丢弃撤销此块代码改动 (git apply --reverse)"
                    class="px-2 py-0.5 rounded bg-red-500/20 hover:bg-red-500/30 text-red-300 font-bold text-[10px] cursor-pointer transition-all active:scale-95"
                  >
                    ✕ 丢弃块
                  </button>
                </div>
              </div>
              <div class="space-y-0.5 font-mono text-[11px] select-text">
                <div
                  v-for="(line, lIdx) in hunk.lines"
                  :key="lIdx"
                  :class="[
                    'px-2 py-0.5 rounded leading-relaxed flex items-center gap-2 whitespace-pre-wrap font-mono transition-colors',
                    line.type === 'add' ? 'bg-[#10A37F]/15 text-emerald-300 border-l-2 border-emerald-500' : '',
                    line.type === 'del' ? 'bg-red-500/15 text-rose-300 border-l-2 border-rose-500' : '',
                    line.type === 'ctx' ? 'text-zinc-400 hover:bg-white/[0.02]' : ''
                  ]"
                >
                  <span class="flex-1">{{ line.text }}</span>
                </div>
              </div>
            </div>
          </template>

          <!-- 备用平铺模式 (Clean 工作区或无 Hunk 分块) -->
          <template v-else>
            <div v-if="s.diffReport?.header" class="text-white/40 pb-1 mb-1 border-b border-white/[0.06] text-[10px]">
              {{ s.diffReport.header }}
            </div>
            <div
              v-for="(line, idx) in (s.diffReport?.lines || [])"
              :key="idx"
              :class="[
                'px-2 py-0.5 rounded leading-relaxed flex items-center gap-2 whitespace-pre-wrap font-mono transition-colors',
                line.type === 'add' ? 'bg-[#10A37F]/15 text-emerald-300 border-l-2 border-emerald-500' : '',
                line.type === 'del' ? 'bg-red-500/15 text-rose-300 border-l-2 border-rose-500' : '',
                line.type === 'ctx' ? 'text-zinc-400 hover:bg-white/[0.02]' : ''
              ]"
            >
              <span class="w-5 text-[10px] select-none opacity-40 font-mono text-right">{{ idx + 1 }}</span>
              <span class="flex-1">{{ line.text }}</span>
            </div>
          </template>
        </div>

        <footer class="h-6 bg-[#FAF8F5] border-t border-black/[0.08] px-3 flex items-center justify-between text-[10px] text-[#71717A] font-mono select-none shrink-0">
          <span>{{ s.diffReport?.lang || 'Go · UTF-8' }}</span>
          <span class="text-emerald-700 font-bold">● Git 磁盘实时同步</span>
        </footer>
      </section>
</template>

<script setup lang="ts">
import { useWorkbenchStore } from '../stores/workbench'
import MonacoEditor from './MonacoEditor.vue'
const s = useWorkbenchStore()
</script>


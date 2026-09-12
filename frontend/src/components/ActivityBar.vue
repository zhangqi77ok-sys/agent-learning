<template>
<nav class="w-12 bg-[#FAF8F5] border-r border-black/[0.08] flex flex-col justify-between py-3 items-center z-20 shrink-0 select-none">
        <div class="flex flex-col gap-2 items-center w-full">
          <button
            @click="s.activeActivity = 'chat'"
            :class="['w-9 h-9 rounded-xl flex items-center justify-center transition-all cursor-pointer', s.activeActivity === 'chat' ? 'bg-white shadow-2xs text-[#D96B27] border border-black/[0.06]' : 'text-[#71717A] hover:bg-black/[0.04]']"
            title="对话工作台"
          >
            <span class="text-base">💬</span>
          </button>
          <button
            @click="s.switchToFileActivity"
            :class="['w-9 h-9 rounded-xl flex items-center justify-center transition-all cursor-pointer', s.activeActivity === 'files' ? 'bg-white shadow-2xs text-[#D96B27] border border-black/[0.06]' : 'text-[#71717A] hover:bg-black/[0.04]']"
            title="工程文件树"
          >
            <span class="text-base">📁</span>
          </button>
          <button
            @click="s.switchToGitActivity"
            :class="['w-9 h-9 rounded-xl flex items-center justify-center transition-all cursor-pointer', s.activeActivity === 'git' ? 'bg-white shadow-2xs text-[#D96B27] border border-black/[0.06]' : 'text-[#71717A] hover:bg-black/[0.04]']"
            title="Git 版本控制"
          >
            <span class="text-base">🌿</span>
          </button>
          <button
            @click="s.openKnowledgeGraphModal"
            class="w-9 h-9 rounded-xl flex items-center justify-center text-[#71717A] hover:bg-black/[0.04] transition-all cursor-pointer"
            title="项目知识图谱与记忆"
          >
            <span class="text-base">🕸️</span>
          </button>
          <button
            @click="s.openSettingsTab('mcp')"
            class="w-9 h-9 rounded-xl flex items-center justify-center text-[#71717A] hover:bg-black/[0.04] transition-all cursor-pointer"
            title="MCP 与技能扩展"
          >
            <span class="text-base">🧩</span>
          </button>
          <button
            @click="s.toggleTerminalDrawer()"
            :class="['w-9 h-9 rounded-xl flex items-center justify-center transition-all cursor-pointer font-mono font-bold text-xs', s.isTerminalOpen ? 'bg-white shadow-2xs text-[#D96B27] border border-black/[0.06]' : 'text-[#71717A] hover:bg-black/[0.04]']"
            title="唤起/收起集成终端抽屉 (Ctrl+`)"
          >
            <span>$_</span>
          </button>
        </div>

        <div class="flex flex-col gap-2 items-center w-full">
          <button
            @click="s.isSettingsOpen = true"
            class="w-9 h-9 rounded-xl flex items-center justify-center text-[#71717A] hover:bg-black/[0.04] transition-all cursor-pointer"
            title="设置"
          >
            <span class="text-base">⚙️</span>
          </button>
          <div class="w-7 h-7 rounded-full bg-gradient-to-tr from-[#D96B27] to-amber-500 text-white flex items-center justify-center text-xs font-bold shadow-2xs">
            湉
          </div>
        </div>
      </nav>
</template>

<script setup lang="ts">
import { useWorkbenchStore } from '../stores/workbench'
const s = useWorkbenchStore()
</script>


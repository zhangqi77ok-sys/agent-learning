<template>
<aside class="w-64 bg-[#FAF8F5] border-r border-black/[0.08] flex flex-col justify-between select-none z-10 shrink-0 font-sans">
        <!-- 抽屉视图 1: 真实会话列表 (Chat Sessions) -->
        <div v-if="s.activeActivity === 'chat'" class="flex flex-col h-full overflow-hidden">
          <div class="p-3 border-b border-black/[0.06] flex items-center justify-between">
            <div class="flex items-center gap-2 min-w-0">
              <span class="text-xs">📦</span>
              <span class="font-bold text-xs text-[#18181B] truncate">agent-learning</span>
            </div>
            <button
              @click="s.createNewSession"
              class="text-[10px] text-[#D96B27] bg-[#D96B27]/10 hover:bg-[#D96B27] hover:text-white px-2 py-0.5 rounded-full font-bold transition-all cursor-pointer flex items-center gap-0.5"
            >
              <span>＋</span><span>新建会话</span>
            </button>
          </div>

          <!-- 场景标签筛选 -->
          <div class="px-3 pt-2 pb-1 border-b border-black/[0.04]">
            <div class="flex items-center gap-1 overflow-x-auto no-scrollbar py-0.5 text-[10px]">
              <button
                v-for="tag in s.availableTags"
                :key="tag"
                @click="s.activeTag = tag"
                :class="[
                  'px-2 py-0.5 rounded-full font-medium transition-all cursor-pointer',
                  s.activeTag === tag ? 'bg-[#D96B27] text-white' : 'bg-white text-[#71717A] hover:text-[#18181B] border border-black/[0.06]'
                ]"
              >
                {{ tag === '全部' ? '全部 (' + s.sessions.length + ')' : '#' + tag }}
              </button>
            </div>
          </div>

          <!-- 真实会话卡片列表 (从 ~/.tiancode/sessions/ 动态读取) -->
          <div class="flex-1 overflow-y-auto p-2 space-y-1.5">
            <div v-if="s.filteredSessions.length === 0" class="p-6 text-center text-[#A1A1AA] text-xs flex flex-col items-center justify-center gap-2 mt-8">
              <span class="text-2xl">📭</span>
              <span>暂无会话记录</span>
              <button @click="s.createNewSession" class="text-[11px] text-[#D96B27] font-semibold hover:underline cursor-pointer">＋ 新建会话</button>
            </div>

            <div
              v-for="sess in s.filteredSessions"
              :key="sess.id"
              @click="s.selectSession(sess.id)"
              :class="[
                'p-2.5 rounded-xl border shadow-xs flex flex-col gap-1 cursor-pointer transition-all',
                s.currentSessionId === sess.id
                  ? 'bg-white border-[#D96B27]/40 ring-2 ring-[#D96B27]/10'
                  : 'bg-white/60 hover:bg-white border-transparent hover:border-black/[0.06]'
              ]"
            >
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-1.5 min-w-0">
                  <span class="text-xs">💬</span>
                  <span class="text-xs font-semibold text-[#18181B] truncate">{{ sess.title }}</span>
                </div>
                <button @click.stop="s.deleteSession(sess.id)" class="text-[#A1A1AA] hover:text-red-500 text-[11px] p-0.5 cursor-pointer" title="删除会话">🗑️</button>
              </div>
              <div class="flex items-center justify-between text-[10px] text-[#71717A] mt-0.5">
                <span v-if="sess.tag" class="bg-[#D96B27]/10 text-[#D96B27] px-1.5 py-0.2 rounded font-medium">#{{ sess.tag }}</span>
                <span v-else class="text-[#A1A1AA]">未分类</span>
                <span class="font-mono text-[#A1A1AA]">{{ sess.time }}</span>
              </div>
              <div v-if="sess.desc" class="text-[11px] text-[#71717A] truncate mt-0.5">{{ sess.desc }}</div>
            </div>
          </div>
        </div>

        <!-- 抽屉视图 2: 真实工程文件树 (File Explorer) -->
        <div v-else-if="s.activeActivity === 'files'" class="flex flex-col h-full overflow-hidden">
          <div class="p-3 border-b border-black/[0.06] flex items-center justify-between">
            <span class="font-bold text-xs text-[#18181B] flex items-center gap-1.5 truncate mr-2" :title="s.workspacePath">
              <span>📁</span><span class="truncate">{{ s.workspaceName }}</span>
            </span>
            <div class="flex items-center gap-1.5 shrink-0">
              <button @click="s.chooseWorkspace" class="text-xs text-[#71717A] hover:text-[#D96B27] cursor-pointer" title="打开/切换工程文件夹 (原生对话框)">📂</button>
              <button @click="s.loadFileTree" class="text-xs text-[#71717A] hover:text-[#D96B27] cursor-pointer" title="刷新文件树">🔄</button>
            </div>
          </div>

          <div class="flex-1 overflow-y-auto p-2 text-xs space-y-1 font-mono">
            <div v-if="s.isFileTreeLoading" class="p-6 text-center text-[#A1A1AA] text-xs flex flex-col items-center justify-center gap-2">
              <span class="animate-spin text-lg">⏳</span>
              <span>正在读取工程目录...</span>
            </div>
            <div v-else-if="s.fileTree.length === 0" class="p-6 text-center text-[#A1A1AA] text-xs">
              <span>工作区暂无文件</span>
            </div>
            <div v-for="node in s.fileTree" :key="node.path" class="space-y-0.5">
              <div
                @click="s.handleFileClick(node)"
                class="px-2 py-1 rounded hover:bg-black/[0.04] cursor-pointer flex items-center justify-between transition-all"
              >
                <div class="flex items-center gap-1.5 min-w-0">
                  <span>{{ node.is_dir ? (s.expandedFolders[node.path] ? '📂' : '📁') : '📄' }}</span>
                  <span class="truncate" :class="{ 'font-bold': node.is_dir }">{{ node.name }}</span>
                </div>
                <span v-if="node.is_dir" class="text-[10px] text-[#A1A1AA]">{{ s.expandedFolders[node.path] ? '▲' : '▼' }}</span>
              </div>

              <div v-if="node.is_dir && s.expandedFolders[node.path] && node.children" class="pl-4 space-y-0.5 border-l border-black/[0.06] ml-2">
                <div
                  v-for="sub in node.children"
                  :key="sub.path"
                  @click="s.openFileDiff(sub.path)"
                  class="px-2 py-0.5 rounded hover:bg-black/[0.04] cursor-pointer flex items-center gap-1.5 text-[11px] text-[#52525B]"
                >
                  <span>{{ sub.is_dir ? '📁' : '📄' }}</span>
                  <span class="truncate">{{ sub.name }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 抽屉视图 3: 真实 Git 变更管理 (Source Control) -->
        <div v-else-if="s.activeActivity === 'git'" class="flex flex-col h-full overflow-hidden">
          <div class="p-3 border-b border-black/[0.06] flex items-center justify-between">
            <span class="font-bold text-xs text-[#18181B] flex items-center gap-1.5">
              <span>🌿</span><span>源代码管理 (Git)</span>
            </span>
            <span class="text-[10px] font-mono text-[#10A37F] bg-[#10A37F]/10 px-1.5 py-0.2 rounded font-bold">
              {{ s.gitStatus.branch || 'main' }}
            </span>
          </div>

          <div class="flex-1 overflow-y-auto p-3 space-y-3 text-xs">
            <div v-if="s.isGitLoading" class="p-6 text-center text-[#A1A1AA] text-xs flex flex-col items-center justify-center gap-2">
              <span class="animate-spin text-lg">⏳</span>
              <span>正在获取 Git 状态...</span>
            </div>
            <div v-else class="space-y-3">
              <!-- 已暂存变更 -->
              <div v-if="s.stagedTreeFiles.length > 0">
                <div class="text-[10px] font-bold text-[#10A37F] uppercase mb-1 flex items-center justify-between">
                  <span>暂存文件 (STAGED)</span>
                  <span class="text-[10px] font-mono bg-[#10A37F]/10 px-1 rounded">{{ s.stagedTreeFiles.length }}</span>
                </div>
                <div class="space-y-1">
                  <div
                    v-for="file in s.stagedTreeFiles"
                    :key="file.path"
                    @click="s.openFileDiff(file.path)"
                    class="p-1.5 rounded hover:bg-black/[0.04] cursor-pointer flex items-center justify-between font-mono text-[11px] group"
                  >
                    <span class="truncate flex-1">{{ file.path }}</span>
                    <div class="flex items-center gap-1.5 shrink-0">
                      <span :class="file.color" class="font-bold">{{ file.type }}</span>
                      <button
                        @click="s.unstageFileAction(file.path, $event)"
                        class="opacity-0 group-hover:opacity-100 text-[#71717A] hover:text-red-500 px-1 rounded hover:bg-black/[0.05] transition-opacity cursor-pointer text-xs"
                        title="取消暂存此文件"
                      >
                        −
                      </button>
                    </div>
                  </div>
                </div>
              </div>

              <!-- 工作区变更 -->
              <div>
                <div class="text-[10px] font-bold text-[#71717A] uppercase mb-1">变更文件 (WORKING TREE)</div>
                <div v-if="s.workingTreeFiles.length === 0" class="p-3 text-center text-[#A1A1AA] text-xs bg-black/[0.02] rounded-lg">
                  ✓ 工作区干净，无未暂存改动
                </div>
                <div v-else class="space-y-1">
                  <div
                    v-for="file in s.workingTreeFiles"
                    :key="file.path"
                    @click="s.openFileDiff(file.path)"
                    class="p-1.5 rounded hover:bg-black/[0.04] cursor-pointer flex items-center justify-between font-mono text-[11px]"
                  >
                    <span class="truncate">{{ file.path }}</span>
                    <span :class="file.color" class="font-bold">{{ file.type }}</span>
                  </div>
                </div>
              </div>
            </div>

            <div class="pt-2 border-t border-black/[0.06] space-y-2">
              <input v-model="s.commitMessage" @keyup.enter="s.handleGitCommit" type="text" placeholder="提交信息 (Commit message)..." class="w-full px-2.5 py-1.5 rounded-lg border border-black/[0.1] text-xs focus:outline-none focus:border-[#D96B27]">
              <button @click="s.handleGitCommit" class="w-full py-1.5 rounded-lg bg-[#D96B27] text-white text-xs font-semibold shadow-xs hover:bg-[#B8551B] cursor-pointer">
                ✓ 提交变更 (Commit & Push)
              </button>
            </div>
          </div>
        </div>
      </aside>
</template>

<script setup lang="ts">
import { useWorkbenchStore } from '../stores/workbench'
const s = useWorkbenchStore()
</script>


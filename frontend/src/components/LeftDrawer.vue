<template>
<aside class="w-64 bg-[#FAF8F5] border-r border-black/[0.08] flex flex-col justify-between select-none z-10 shrink-0 font-sans">
        <!-- 抽屉视图 1: 真实会话列表 (Chat Sessions) -->
        <div v-if="s.activeActivity === 'chat'" class="flex flex-col h-full overflow-hidden">
          <div class="p-3 border-b border-black/[0.06] flex items-center justify-between gap-2">
            <div class="flex items-center gap-1.5 min-w-0">
              <span class="text-xs font-bold text-[#18181B] tracking-tight">会话分支</span>
              <span class="text-[10px] text-[#71717A] bg-black/[0.04] px-1.5 py-0.2 rounded-full font-mono">{{ s.projects.length }} 个项目</span>
            </div>
            <button
              @click="s.openProjectFolder"
              class="flex items-center gap-1 px-2 py-1 rounded-md text-[11px] font-medium text-[#18181B] bg-white hover:bg-black/[0.04] border border-black/[0.08] shadow-2xs cursor-pointer shrink-0"
            >
              <span>📂</span><span>打开项目</span>
            </button>
          </div>

          <div class="p-2.5 space-y-2 border-b border-black/[0.06]">
            <input
              v-model="s.sessionSearch"
              type="text"
              placeholder="搜索会话标题…"
              class="w-full h-7 px-2 rounded-lg bg-white text-xs border border-black/[0.08] focus:border-[#D96B27] focus:outline-none placeholder:text-[#A1A1AA]"
            >
            <div v-if="s.availableTags.length > 1" class="flex items-center gap-1 overflow-x-auto no-scrollbar py-0.5 text-[10px]">
              <button
                v-for="tag in s.availableTags"
                :key="tag"
                @click="s.activeTag = tag"
                :class="[
                  'px-2 py-0.5 rounded-full font-medium cursor-pointer',
                  s.activeTag === tag ? 'bg-[#D96B27] text-white' : 'bg-white text-[#71717A] border border-black/[0.06]'
                ]"
              >
                {{ tag === '全部' ? '全部' : '#' + tag }}
              </button>
            </div>
          </div>

          <div class="flex-1 overflow-y-auto p-2 space-y-3">
            <div v-if="s.projectTree.length === 0" class="p-6 text-center text-[#A1A1AA] text-xs space-y-2">
              <div>还没有打开的项目</div>
              <button @click="s.openProjectFolder" class="text-[#D96B27] underline cursor-pointer">打开本地文件夹</button>
            </div>

            <div
              v-for="proj in s.projectTree"
              :key="proj.path"
              class="rounded-xl border border-black/[0.08] bg-white/70 shadow-2xs overflow-hidden"
            >
              <div
                class="p-2 flex items-center justify-between hover:bg-black/[0.02] cursor-pointer group"
                @click="s.toggleProjectCollapse(proj.path)"
              >
                <div class="flex items-center gap-1.5 min-w-0">
                  <span class="text-[10px] text-[#71717A]">{{ s.collapsedProjects[proj.path] ? '▶' : '▼' }}</span>
                  <span class="text-xs">📁</span>
                  <span
                    class="text-xs font-semibold truncate"
                    :class="s.workspacePath && proj.path.replace(/\\/g,'/').toLowerCase() === s.workspacePath.replace(/\\/g,'/').toLowerCase() ? 'text-[#D96B27]' : 'text-[#18181B]'"
                  >{{ proj.name }}</span>
                </div>
                <div class="flex items-center gap-0.5 shrink-0" @click.stop>
                  <button
                    @click="s.createSessionInProject(proj.path)"
                    title="在此项目下新建会话"
                    class="p-1 rounded hover:bg-black/[0.06] text-[#71717A] hover:text-[#D96B27] cursor-pointer text-xs"
                  >＋</button>
                  <button
                    @click="s.unpinProject(proj.path)"
                    title="从列表移除项目（不删磁盘和历史会话）"
                    class="p-1 rounded hover:bg-black/[0.06] text-[#A1A1AA] hover:text-red-500 cursor-pointer text-[10px]"
                  >✕</button>
                </div>
              </div>

              <div v-show="!s.collapsedProjects[proj.path]" class="p-1 space-y-1 border-t border-black/[0.04] bg-[#FAF8F5]">
                <div v-if="proj.sessions.length === 0" class="px-2 py-2 text-[10px] text-[#A1A1AA]">
                  此项目还没有已保存的对话，发送第一条消息后会出现在这里
                </div>
                <div
                  v-for="sess in proj.sessions"
                  :key="sess.id"
                  @click="s.activateSession(sess.id, sess.workspace)"
                  :class="[
                    'session-item p-2 rounded-lg flex flex-col gap-1 cursor-pointer transition-all border',
                    s.currentSessionId === sess.id
                      ? 'bg-white border-[#D96B27]/40 shadow-xs'
                      : 'hover:bg-white/80 border-transparent hover:border-black/[0.06]'
                  ]"
                >
                  <div class="flex items-center justify-between gap-1">
                    <span class="text-xs font-semibold text-[#18181B] truncate">{{ sess.title }}</span>
                    <button @click.stop="s.deleteSession(sess.id)" class="text-[#A1A1AA] hover:text-red-500 text-[10px] cursor-pointer" title="删除会话">🗑</button>
                  </div>
                  <div class="flex items-center gap-1 text-[10px] text-[#71717A]">
                    <span v-if="sess.tag" class="bg-[#D96B27]/10 text-[#D96B27] px-1 py-0.2 rounded">#{{ sess.tag }}</span>
                    <span class="truncate flex-1">{{ sess.desc }}</span>
                    <span class="font-mono text-[#A1A1AA] shrink-0">{{ sess.time }}</span>
                  </div>
                </div>
              </div>
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
              {{ s.gitCurrentBranch || s.gitBranchLabel }}
            </span>
          </div>

          <div class="flex-1 overflow-y-auto p-3 space-y-3 text-xs">
            <div class="space-y-1.5">
              <div class="text-[10px] font-bold text-[#71717A] uppercase">分支</div>
              <select
                v-if="s.gitBranches.length > 0"
                class="w-full px-2 py-1.5 rounded-lg border border-black/[0.1] text-xs bg-white"
                :value="s.gitCurrentBranch"
                @change="s.checkoutBranch(($event.target as HTMLSelectElement).value)"
              >
                <option v-for="b in s.gitBranches" :key="b" :value="b">{{ b }}</option>
              </select>
              <div v-else class="text-[10px] text-[#A1A1AA]">当前工作区没有 Git 仓库</div>
              <div class="flex gap-1">
                <input v-model="s.newBranchName" type="text" placeholder="新分支名" class="flex-1 px-2 py-1 rounded-lg border border-black/[0.1] text-xs">
                <button @click="s.createBranchAction" class="px-2 py-1 rounded-lg bg-white border border-black/[0.08] text-[10px] cursor-pointer">创建</button>
              </div>
            </div>
            <div class="space-y-1.5">
              <div class="flex items-center justify-between">
                <span class="text-[10px] font-bold text-[#71717A] uppercase">快照 (stash)</span>
                <button @click="s.createSnapshotAction" class="text-[10px] text-[#D96B27] cursor-pointer">＋ 保存</button>
              </div>
              <div v-if="s.gitSnapshots.length === 0" class="text-[10px] text-[#A1A1AA]">没有快照</div>
              <div v-for="snap in s.gitSnapshots" :key="snap.id" class="flex items-center justify-between p-1.5 rounded bg-black/[0.02]">
                <span class="truncate font-mono text-[10px]">{{ snap.message || snap.id }} · {{ snap.time }}</span>
                <button @click="s.restoreSnapshotAction(snap.id)" class="text-[10px] text-[#D96B27] cursor-pointer shrink-0">还原</button>
              </div>
            </div>
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
                ✓ 提交到本地仓库 (Commit)
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


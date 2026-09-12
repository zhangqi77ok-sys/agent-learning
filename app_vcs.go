package main

import (
	"tiancode/internal/agent"
	"tiancode/internal/gitops"
	"tiancode/internal/telemetry"

	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"tiancode/internal/diff"
)

// --- 代码行级 Diff 审查 ---

func (a *App) GetStructuredDiff(filePath string) (diff.DiffReport, error) {
	if a.sandbox != nil {
		if _, err := a.sandbox.ValidatePath(filePath); err != nil {
			return diff.DiffReport{}, fmt.Errorf("security violation: %w", err)
		}
	}
	return diff.ComputeFileDiff(a.workspace, filePath)
}

func (a *App) RevertFile(filePath string) error {
	var validPath string
	var err error
	if a.sandbox != nil {
		validPath, err = a.sandbox.ValidatePath(filePath)
		if err != nil {
			return fmt.Errorf("security violation: %w", err)
		}
	} else {
		validPath = filepath.Join(a.workspace, filePath)
	}

	// 1. 先检查该文件在 git 中的真实状态，严禁无条件物理删除
	statusCmd := exec.Command("git", "status", "--porcelain", "--", filePath)
	statusCmd.Dir = a.workspace
	if attr := windowsSysProcAttr(); attr != nil {
		statusCmd.SysProcAttr = attr
	}
	statusOut, _ := statusCmd.Output()
	statusStr := strings.TrimSpace(string(statusOut))

	// 若确认为未追踪文件 (??)，撤销即删除该未追踪新文件
	if strings.HasPrefix(statusStr, "??") {
		if fi, statErr := os.Stat(validPath); statErr == nil && !fi.IsDir() {
			return os.Remove(validPath)
		}
		return nil
	}

	// 针对无 HEAD 仓库（刚初始化尚未产生首次 commit）的暂存新增文件 (A / AM)
	// git restore 与 git checkout HEAD 均会因无法解析 HEAD 而失败 (exit status 128)
	// 此时安全解法：执行 git rm --cached -f 取消暂存并删除未提交的新文件
	if !diff.HasGitHead(a.workspace) && strings.HasPrefix(statusStr, "A") {
		rmCmd := exec.Command("git", "rm", "--cached", "-f", "--", filePath)
		rmCmd.Dir = a.workspace
		if attr := windowsSysProcAttr(); attr != nil {
			rmCmd.SysProcAttr = attr
		}
		_ = rmCmd.Run()
		if fi, statErr := os.Stat(validPath); statErr == nil && !fi.IsDir() {
			return os.Remove(validPath)
		}
		return nil
	}

	// 2. 对于已追踪文件，优先使用 git restore 撤销暂存区和工作区修改
	restoreCmd := exec.Command("git", "restore", "--staged", "--worktree", "--", filePath)
	restoreCmd.Dir = a.workspace
	if attr := windowsSysProcAttr(); attr != nil {
		restoreCmd.SysProcAttr = attr
	}
	if err := restoreCmd.Run(); err == nil {
		return nil
	}

	// 3. 降级尝试 git checkout HEAD -- filePath
	checkoutCmd := exec.Command("git", "checkout", "HEAD", "--", filePath)
	checkoutCmd.Dir = a.workspace
	if attr := windowsSysProcAttr(); attr != nil {
		checkoutCmd.SysProcAttr = attr
	}
	return checkoutCmd.Run()
}

// ApplyDiffHunk 采纳指定代码块 (Hunk) 变更
func (a *App) ApplyDiffHunk(filePath string, hunkIndex int, stageOnly bool) error {
	if a.sandbox != nil {
		if _, err := a.sandbox.ValidatePath(filePath); err != nil {
			return fmt.Errorf("security violation: %w", err)
		}
	}
	return diff.ApplyHunkPatch(a.workspace, filePath, hunkIndex, stageOnly)
}

// DiscardDiffHunk 丢弃指定代码块 (Hunk) 变更
func (a *App) DiscardDiffHunk(filePath string, hunkIndex int) error {
	if a.sandbox != nil {
		if _, err := a.sandbox.ValidatePath(filePath); err != nil {
			return fmt.Errorf("security violation: %w", err)
		}
	}
	return diff.DiscardHunkPatch(a.workspace, filePath, hunkIndex)
}

// --- 渠道与插件设置 ---
func (a *App) GitCommit(msg string) (string, error) {
	cleanMsg := strings.TrimSpace(msg)
	if cleanMsg == "" {
		cleanMsg = "feat: update by tcode agent"
	}

	// 1. 先检查暂存区是否已有内容 (git diff --cached --quiet 退出码 1 代表有暂存变更)
	diffCachedCmd := exec.Command("git", "diff", "--cached", "--quiet")
	diffCachedCmd.Dir = a.workspace
	if attr := windowsSysProcAttr(); attr != nil {
		diffCachedCmd.SysProcAttr = attr
	}
	hasStaged := diffCachedCmd.Run() != nil // 退出码非 0 说明已有暂存改动

	// 2. 若暂存区完全为空，自动将工作区变动暂存
	if !hasStaged {
		cmdAdd := exec.Command("git", "add", "-A")
		cmdAdd.Dir = a.workspace
		if attr := windowsSysProcAttr(); attr != nil {
			cmdAdd.SysProcAttr = attr
		}
		if err := cmdAdd.Run(); err != nil {
			return "", err
		}
	}

	cmdCommit := exec.Command("git", "commit", "-m", cleanMsg)
	cmdCommit.Dir = a.workspace
	if attr := windowsSysProcAttr(); attr != nil {
		cmdCommit.SysProcAttr = attr
	}
	out, err := cmdCommit.CombinedOutput()
	outStr := string(out)
	if err != nil {
		// 当工作区没有发生任何变更时，git commit 会返回 exit status 1 伴随 nothing to commit
		if strings.Contains(outStr, "nothing to commit") || strings.Contains(outStr, "无文件要提交") {
			return outStr, nil
		}
		return outStr, err
	}
	return outStr, nil
}

// GitStage 真实暂存单个文件
func (a *App) GitStage(filePath string) error {
	if a.sandbox != nil {
		if _, err := a.sandbox.ValidatePath(filePath); err != nil {
			return fmt.Errorf("security violation: %w", err)
		}
	}
	cmd := exec.Command("git", "add", "--", filePath)
	cmd.Dir = a.workspace
	if attr := windowsSysProcAttr(); attr != nil {
		cmd.SysProcAttr = attr
	}
	return cmd.Run()
}

// GitUnstage 真实取消暂存单个文件
func (a *App) GitUnstage(filePath string) error {
	if a.sandbox != nil {
		if _, err := a.sandbox.ValidatePath(filePath); err != nil {
			return fmt.Errorf("security violation: %w", err)
		}
	}

	// 针对无 HEAD 仓库（刚 git init 尚未提交），git restore --staged 会失败 (could not resolve HEAD)
	// 此时优先使用 git rm --cached -f
	if !diff.HasGitHead(a.workspace) {
		rmCmd := exec.Command("git", "rm", "--cached", "-f", "--", filePath)
		rmCmd.Dir = a.workspace
		if attr := windowsSysProcAttr(); attr != nil {
			rmCmd.SysProcAttr = attr
		}
		if err := rmCmd.Run(); err == nil {
			return nil
		}
	}

	cmd := exec.Command("git", "restore", "--staged", "--", filePath)
	cmd.Dir = a.workspace
	if attr := windowsSysProcAttr(); attr != nil {
		cmd.SysProcAttr = attr
	}
	if err := cmd.Run(); err == nil {
		return nil
	}
	// 降级使用 git reset HEAD -- filePath
	resetCmd := exec.Command("git", "reset", "HEAD", "--", filePath)
	resetCmd.Dir = a.workspace
	if attr := windowsSysProcAttr(); attr != nil {
		resetCmd.SysProcAttr = attr
	}
	return resetCmd.Run()
}

// GitListBranches 列出本地全部分支与当前分支
func (a *App) GitListBranches() ([]string, string, error) {
	return gitops.ListBranches(a.workspace)
}

// GetGitBranches 给前端一个可 JSON 的分支清单（Wails 多返回值只会露出第一项）
func (a *App) GetGitBranches() (map[string]any, error) {
	branches, current, err := gitops.ListBranches(a.workspace)
	if err != nil {
		return map[string]any{"branches": []string{}, "current": ""}, nil
	}
	if branches == nil {
		branches = []string{}
	}
	return map[string]any{"branches": branches, "current": current}, nil
}

// GitCheckoutBranch 切换检出分支
func (a *App) GitCheckoutBranch(name string) error {
	return gitops.CheckoutBranch(a.workspace, name)
}

// GitCreateBranch 创建并检出新分支
func (a *App) GitCreateBranch(name string) error {
	return gitops.CreateBranch(a.workspace, name)
}

// GitListSnapshots 枚举快照与暂存回溯点
func (a *App) GitListSnapshots() ([]gitops.Snapshot, error) {
	return gitops.ListSnapshots(a.workspace)
}

// GitCreateSnapshot 创建当前工作区检查点快照
func (a *App) GitCreateSnapshot(msg string) error {
	return gitops.CreateSnapshot(a.workspace, msg)
}

// GitRestoreSnapshot 还原指定快照
func (a *App) GitRestoreSnapshot(stashID string) error {
	return gitops.RestoreSnapshot(a.workspace, stashID)
}

// RunTDDValidation 触发 Sub-Agent TDD 自动化单测红绿灯自愈检查
func (a *App) RunTDDValidation() (agent.TestReport, error) {
	return agent.RunTDDValidation(a.workspace)
}

// RunSecurityAudit 触发 Sub-Agent 安全沙箱代码与命令审查
func (a *App) RunSecurityAudit() (agent.AuditReport, error) {
	return agent.RunSecurityAudit(a.workspace)
}

func (a *App) GitPull() (string, error) {
	cmd := exec.Command("git", "pull", "--rebase", "--autostash")
	cmd.Dir = a.workspace
	if attr := windowsSysProcAttr(); attr != nil {
		cmd.SysProcAttr = attr
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("%s: %w", strings.TrimSpace(string(out)), err)
	}
	return string(out), nil
}

func (a *App) GitPush() (string, error) {
	cmd := exec.Command("git", "push")
	cmd.Dir = a.workspace
	if attr := windowsSysProcAttr(); attr != nil {
		cmd.SysProcAttr = attr
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("%s: %w", strings.TrimSpace(string(out)), err)
	}
	return string(out), nil
}

// GetUsageMetrics 获取 Token 消耗与网关使用量统计大盘
func (a *App) GetUsageMetrics() telemetry.UsageMetrics {
	activeCount := 0
	if a.sessionStore != nil {
		sessions := a.sessionStore.List(a.workspace)
		activeCount = len(sessions)
	}
	return telemetry.GetTracker().GetMetrics(activeCount)
}

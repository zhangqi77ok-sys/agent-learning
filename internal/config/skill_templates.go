package config

func BuiltinSkillTemplates() []SkillConfig {
	return []SkillConfig{
		{ID: "tpl_tdd", Name: "tdd-test-runner", Description: "写入后跑 go test / npm test，把失败栈带回下一轮自愈。", Prompt: "你必须在修改代码后运行对应测试。失败时读取堆栈并修复，直到测试通过或明确说明阻塞。", Enabled: true},
		{ID: "tpl_diff", Name: "git-diff-reviewer", Description: "写盘前对照 git diff，避免覆盖未保存工作。", Prompt: "在 fs_control 写文件前先查看 git status 与相关 diff。不要丢弃用户未提交改动。", Enabled: true},
		{ID: "tpl_arch", Name: "architecture-analyzer", Description: "扫描依赖与分层，指出循环引用。", Prompt: "审查包依赖方向与分层。发现循环依赖或跨层调用时明确指出文件路径。", Enabled: true},
	}
}

package search

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"tiancode/internal/core/sandbox"
	v1 "tiancode/pkg/plugin/v1"
)

// Tool 工作区高效检索算子插件 (grep & glob find)
type Tool struct {
	id      string
	name    string
	version string
	sandbox *sandbox.Sandbox
}

// NewTool 构造检索插件实例
func NewTool(sb *sandbox.Sandbox) *Tool {
	return &Tool{
		id:      "tool.search",
		name:    "Workspace Search Tool (grep & glob)",
		version: "1.0.0",
		sandbox: sb,
	}
}

func (t *Tool) ID() string                          { return t.id }
func (t *Tool) Name() string                        { return t.name }
func (t *Tool) Version() string                     { return t.version }
func (t *Tool) Type() v1.PluginType                 { return v1.TypeTool }
func (t *Tool) Init(ctx context.Context, cfg json.RawMessage) error { return nil }
func (t *Tool) Start(ctx context.Context) error     { return nil }
func (t *Tool) Stop(ctx context.Context) error      { return nil }
func (t *Tool) Health(ctx context.Context) v1.HealthStatus {
	return v1.HealthStatus{Healthy: true, Message: "Search tool ready"}
}

func (t *Tool) Definition() v1.ToolDefinition {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"action": map[string]any{
				"type":        "string",
				"enum":        []string{"grep", "find"},
				"description": "搜索模式: grep (检索文件内文本/代码片段), find (根据文件名或通配符检索文件路径)",
			},
			"query": map[string]any{
				"type":        "string",
				"description": "搜索关键词或模式: grep 时为搜索文本或正则表达式；find 时为文件名或通配符 (例如 '*.go', 'config')",
			},
			"path": map[string]any{
				"type":        "string",
				"description": "可选。在工作区特定的子目录中检索，留空或 '.' 表示工作区根目录",
			},
			"max_results": map[string]any{
				"type":        "integer",
				"description": "可选。最多返回的结果条数，默认 30，最大 100",
			},
			"case_sensitive": map[string]any{
				"type":        "boolean",
				"description": "可选。是否大小写敏感，默认 false",
			},
		},
		"required": []string{"action", "query"},
	}
	schemaBytes, _ := json.Marshal(schema)

	return v1.ToolDefinition{
		Name:        "search_workspace",
		Description: "在工作区沙箱内快速检索文件内容 (grep) 或按文件名通配搜索 (find)。优先使用本工具快速定位关键代码与符号，避免无脑递归遍历整个目录。",
		Parameters:  schemaBytes,
	}
}

func (t *Tool) Execute(ctx context.Context, rawArgs json.RawMessage) (*v1.ToolResult, error) {
	if t.sandbox == nil {
		return &v1.ToolResult{Content: "error: search sandbox not initialized", IsError: true}, nil
	}

	var args struct {
		Action        string `json:"action"`
		Query         string `json:"query"`
		Pattern       string `json:"pattern"`
		Path          string `json:"path"`
		Dir           string `json:"dir"`
		SubDir        string `json:"sub_dir"`
		MaxResults    int    `json:"max_results"`
		CaseSensitive bool   `json:"case_sensitive"`
	}

	if err := json.Unmarshal(rawArgs, &args); err != nil {
		return &v1.ToolResult{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	query := strings.TrimSpace(args.Query)
	if query == "" {
		query = strings.TrimSpace(args.Pattern)
	}
	if query == "" {
		return &v1.ToolResult{Content: "search error: empty search query", IsError: true}, nil
	}

	subPath := strings.TrimSpace(args.Path)
	if subPath == "" {
		subPath = strings.TrimSpace(args.Dir)
	}
	if subPath == "" {
		subPath = strings.TrimSpace(args.SubDir)
	}
	if subPath == "" || subPath == "." {
		subPath = ""
	}

	maxResults := args.MaxResults
	if maxResults <= 0 {
		maxResults = 30
	}
	if maxResults > 100 {
		maxResults = 100
	}

	targetDir := t.sandbox.Root()
	if subPath != "" {
		validDir, err := t.sandbox.ValidatePath(subPath)
		if err != nil {
			return &v1.ToolResult{Content: fmt.Sprintf("invalid search path: %v", err), IsError: true}, nil
		}
		targetDir = validDir
	}

	action := strings.ToLower(strings.TrimSpace(args.Action))
	if action == "" || action == "search" {
		action = "grep"
	}

	switch action {
	case "grep":
		return t.executeGrep(ctx, targetDir, query, args.CaseSensitive, maxResults)
	case "find":
		return t.executeFind(ctx, targetDir, query, maxResults)
	default:
		return &v1.ToolResult{Content: fmt.Sprintf("unknown action '%s', expected 'grep' or 'find'", action), IsError: true}, nil
	}
}

func shouldSkipDir(name string) bool {
	if strings.HasPrefix(name, ".") {
		return true
	}
	switch name {
	case "node_modules", "vendor", "dist", "bin", "build", "target", "tmp", "release", ".git":
		return true
	}
	return false
}

func isBinaryExtension(ext string) bool {
	ext = strings.ToLower(ext)
	switch ext {
	case ".exe", ".dll", ".so", ".dylib", ".bin", ".iso", ".zip", ".tar", ".gz", ".7z",
		".png", ".jpg", ".jpeg", ".gif", ".webp", ".ico", ".pdf", ".mp4", ".mp3", ".wav",
		".woff", ".woff2", ".ttf", ".eot":
		return true
	}
	return false
}

func (t *Tool) executeGrep(ctx context.Context, targetDir, query string, caseSensitive bool, maxResults int) (*v1.ToolResult, error) {
	var re *regexp.Regexp
	var err error
	pattern := query
	if !caseSensitive {
		pattern = "(?i)" + regexp.QuoteMeta(query)
		re, err = regexp.Compile(pattern)
	} else {
		re, err = regexp.Compile(regexp.QuoteMeta(query))
	}
	if err != nil {
		// 回退为普通包含匹配
		re = nil
	}

	results := make([]string, 0)
	count := 0

	err = filepath.WalkDir(targetDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if d.IsDir() {
			if shouldSkipDir(d.Name()) && path != targetDir {
				return filepath.SkipDir
			}
			return nil
		}

		if isBinaryExtension(filepath.Ext(path)) {
			return nil
		}

		// 忽略超大文件 (> 2MB)
		info, err := d.Info()
		if err == nil && info.Size() > 2*1024*1024 {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		rel, _ := filepath.Rel(t.sandbox.Root(), path)
		rel = filepath.ToSlash(rel)

		scanner := bufio.NewScanner(file)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			matched := false
			if re != nil {
				matched = re.MatchString(line)
			} else {
				if caseSensitive {
					matched = strings.Contains(line, query)
				} else {
					matched = strings.Contains(strings.ToLower(line), strings.ToLower(query))
				}
			}

			if matched {
				trimmed := strings.TrimSpace(line)
				if len(trimmed) > 200 {
					trimmed = trimmed[:200] + "..."
				}
				results = append(results, fmt.Sprintf("%s:%d: %s", rel, lineNum, trimmed))
				count++
				if count >= maxResults {
					return filepath.SkipAll
				}
			}
		}

		return nil
	})

	if err != nil && err != filepath.SkipAll && err != context.Canceled {
		return &v1.ToolResult{Content: fmt.Sprintf("grep execution error: %v", err), IsError: true}, nil
	}

	if len(results) == 0 {
		return &v1.ToolResult{Content: fmt.Sprintf("未在工作区找到匹配 '%s' 的代码行", query), IsError: false}, nil
	}

	summary := fmt.Sprintf("共找到 %d 处匹配:\n%s", len(results), strings.Join(results, "\n"))
	return &v1.ToolResult{Content: summary, IsError: false}, nil
}

func (t *Tool) executeFind(ctx context.Context, targetDir, query string, maxResults int) (*v1.ToolResult, error) {
	results := make([]string, 0)
	pattern := strings.ToLower(query)
	isGlob := strings.ContainsAny(query, "*?[]")

	err := filepath.WalkDir(targetDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		name := d.Name()
		if d.IsDir() {
			if shouldSkipDir(name) && path != targetDir {
				return filepath.SkipDir
			}
		}

		matched := false
		if isGlob {
			m, _ := filepath.Match(pattern, strings.ToLower(name))
			matched = m
		} else {
			matched = strings.Contains(strings.ToLower(name), pattern)
		}

		if matched && path != targetDir {
			rel, _ := filepath.Rel(t.sandbox.Root(), path)
			rel = filepath.ToSlash(rel)
			if d.IsDir() {
				rel += "/"
			}
			results = append(results, rel)
			if len(results) >= maxResults {
				return filepath.SkipAll
			}
		}

		return nil
	})

	if err != nil && err != filepath.SkipAll && err != context.Canceled {
		return &v1.ToolResult{Content: fmt.Sprintf("find execution error: %v", err), IsError: true}, nil
	}

	if len(results) == 0 {
		return &v1.ToolResult{Content: fmt.Sprintf("未找到名称匹配 '%s' 的文件或目录", query), IsError: false}, nil
	}

	summary := fmt.Sprintf("共找到 %d 个文件/目录:\n%s", len(results), strings.Join(results, "\n"))
	return &v1.ToolResult{Content: summary, IsError: false}, nil
}

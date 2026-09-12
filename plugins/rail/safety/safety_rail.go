package safety

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	v1 "tiancode/pkg/plugin/v1"
)

var dangerous = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\brm\s+(-[a-zA-Z]*f[a-zA-Z]*\s+)?-?[rR]f\b`),
	regexp.MustCompile(`(?i)\brm\s+-rf\b`),
	regexp.MustCompile(`(?i)\bdel\s+/[sS]\b`),
	regexp.MustCompile(`(?i)Remove-Item\s+.*-Recurse`),
	regexp.MustCompile(`(?i)\bformat\s+[a-zA-Z]:`),
	regexp.MustCompile(`(?i)\bFormat-Volume\b`),
	regexp.MustCompile(`(?i)\bshutdown\s+/[sStT]`),
	regexp.MustCompile(`(?i)\bmkfs\b`),
	regexp.MustCompile(`(?i)\bdd\s+if=`),
	regexp.MustCompile(`(?i)rd\s+/s\b`),
}

type Rail struct{}

func New() *Rail { return &Rail{} }

func (r *Rail) ID() string             { return "rail.safety" }
func (r *Rail) Name() string           { return "SafetyRail" }
func (r *Rail) Version() string        { return "1.0.0" }
func (r *Rail) Type() v1.PluginType    { return v1.TypeRail }
func (r *Rail) Priority() int          { return 100 }
func (r *Rail) Init(context.Context, json.RawMessage) error { return nil }
func (r *Rail) Start(context.Context) error                 { return nil }
func (r *Rail) Stop(context.Context) error                  { return nil }
func (r *Rail) Health(context.Context) v1.HealthStatus {
	return v1.HealthStatus{Healthy: true, Message: "SafetyRail armed"}
}

func (r *Rail) OnBeforeObserve(context.Context, string) error { return nil }
func (r *Rail) OnBeforeReason(_ context.Context, _ string, prompt *string) error {
	if prompt == nil {
		return nil
	}
	*prompt = StripSecretsFromPrompt(*prompt)
	return nil
}

func (r *Rail) OnBeforeAct(_ context.Context, _ string, toolName string, args []byte) (*v1.RailDecision, error) {
	blob := strings.ToLower(toolName + " " + string(args))
	if strings.Contains(blob, "..") && (strings.Contains(blob, "write") || strings.Contains(blob, "rel_path") || strings.Contains(blob, "path")) {
		if strings.Contains(string(args), "..") {
			return &v1.RailDecision{Allow: false, Intercepted: true, Reason: "path traversal blocked"}, nil
		}
	}
	// 检查 rm 的危险选项
	if strings.Contains(blob, "exec_command") {
		// 提取 command 字段
		var payload struct {
			Command string `json:"command"`
		}
		if err := json.Unmarshal(args, &payload); err == nil {
			if isDangerousRm(payload.Command) {
				return &v1.RailDecision{Allow: false, Intercepted: true, NeedsConfirm: true, Reason: "dangerous rm command blocked by SafetyRail (semantic)"}, nil
			}
		}
	}

	hay := toolName + " " + string(args)
	for _, re := range dangerous {
		if re.MatchString(hay) {
			return &v1.RailDecision{
				Allow:        false,
				Intercepted:  true,
				NeedsConfirm: true,
				Reason:       fmt.Sprintf("dangerous command blocked by SafetyRail: %s", re.String()),
			}, nil
		}
	}
	return &v1.RailDecision{Allow: true}, nil
}

func isDangerousRm(cmd string) bool {
	cmd = strings.TrimSpace(cmd)
	if !strings.HasPrefix(strings.ToLower(cmd), "rm ") && strings.ToLower(cmd) != "rm" {
		return false
	}
	
	tokens := strings.Fields(cmd)
	hasForce := false
	hasRecursive := false
	
	for _, token := range tokens[1:] {
		if token == "--force" {
			hasForce = true
		} else if token == "--recursive" || token == "-R" {
			hasRecursive = true
		} else if strings.HasPrefix(token, "-") && !strings.HasPrefix(token, "--") {
			// 短选项拆分
			for _, ch := range token[1:] {
				if ch == 'f' || ch == 'F' {
					hasForce = true
				} else if ch == 'r' || ch == 'R' {
					hasRecursive = true
				}
			}
		}
	}
	return hasForce && hasRecursive
}

func (r *Rail) OnAfterAct(context.Context, string, string, *v1.ToolResult) error { return nil }
func (r *Rail) OnVerify(context.Context, string) (bool, string, error) {
	return true, "", nil
}

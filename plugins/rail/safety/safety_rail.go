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
	hay := toolName + " " + string(args)
	for _, re := range dangerous {
		if re.MatchString(hay) {
			return &v1.RailDecision{
				Allow:       false,
				Intercepted: true,
				Reason:      fmt.Sprintf("dangerous command blocked by SafetyRail: %s", re.String()),
			}, nil
		}
	}
	return &v1.RailDecision{Allow: true}, nil
}

func (r *Rail) OnAfterAct(context.Context, string, string, *v1.ToolResult) error { return nil }
func (r *Rail) OnVerify(context.Context, string) (bool, string, error) {
	return true, "", nil
}

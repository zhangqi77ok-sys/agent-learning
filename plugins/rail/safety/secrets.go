package safety

import (
	"regexp"
	"strings"
)

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`sk-[A-Za-z0-9_-]{8,}`),
	regexp.MustCompile(`ghp_[A-Za-z0-9]{10,}`),
	regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[=:]\s*\S+`),
	regexp.MustCompile(`(?i)(api[_-]?key|secret|token)\s*[=:]\s*\S+`),
}

// StripSecretsFromPrompt redacts common API keys and password assignments from prompts.
func StripSecretsFromPrompt(s string) string {
	out := s
	for _, re := range secretPatterns {
		out = re.ReplaceAllStringFunc(out, func(m string) string {
			if i := strings.IndexAny(m, "=:"); i >= 0 {
				return m[:i+1] + " [REDACTED]"
			}
			return "[REDACTED]"
		})
	}
	return out
}

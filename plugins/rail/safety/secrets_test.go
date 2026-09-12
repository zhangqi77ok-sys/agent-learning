package safety

import "testing"

func TestStripSecretsFromPrompt(t *testing.T) {
	in := "key=sk-abc123secret and token: ghp_abcdefghijklmnop and password=hunter2"
	out := StripSecretsFromPrompt(in)
	if out == in {
		t.Fatal("expected redaction")
	}
	if containsAny(out, "sk-abc123secret", "ghp_abcdefghijklmnop", "hunter2") {
		t.Fatalf("secret leaked: %s", out)
	}
	clean := StripSecretsFromPrompt("hello world")
	if clean != "hello world" {
		t.Fatalf("clean text changed: %s", clean)
	}
}

func containsAny(s string, parts ...string) bool {
	for _, p := range parts {
		if len(p) > 0 && len(s) > 0 {
			for i := 0; i+len(p) <= len(s); i++ {
				if s[i:i+len(p)] == p {
					return true
				}
			}
		}
	}
	return false
}

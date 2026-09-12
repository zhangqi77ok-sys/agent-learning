package loop

import (
	"errors"
	"strings"
	"testing"
)

func TestFormatHitCapNotice(t *testing.T) {
	notice := FormatHitCapNotice(24)
	if !strings.Contains(notice, "24") {
		t.Errorf("expected turns 24 in notice, got: %s", notice)
	}
	if !strings.Contains(notice, "继续") {
		t.Errorf("expected instruction to continue, got: %s", notice)
	}
}

func TestFormatInterruptedNotice(t *testing.T) {
	notice := FormatInterruptedNotice()
	if !strings.Contains(notice, "手动中止") && !strings.Contains(notice, "已中止") {
		t.Errorf("expected interruption note, got: %s", notice)
	}
	if !strings.Contains(notice, "继续") {
		t.Errorf("expected instruction to continue in interrupted notice, got: %s", notice)
	}
}

func TestFormatUpstreamError_StatusCodes(t *testing.T) {
	cases := []struct {
		errMsg   string
		expected []string
	}{
		{
			"upstream returned HTTP 400: {\"error\": \"invalid request payload\"}",
			[]string{"400", "请求格式或参数不合法", "上下文"},
		},
		{
			"upstream returned HTTP 401: Unauthorized API key",
			[]string{"401", "API Key", "鉴权"},
		},
		{
			"upstream returned HTTP 429: rate limit exceeded or quota exhausted",
			[]string{"429", "频率超限", "配额"},
		},
		{
			"upstream returned HTTP 500: internal server error",
			[]string{"500", "上游模型服务发生内部故障"},
		},
		{
			"dial tcp: connection refused",
			[]string{"网络连接异常"},
		},
	}

	for _, c := range cases {
		out := FormatUpstreamError(errors.New(c.errMsg))
		for _, exp := range c.expected {
			if !strings.Contains(out, exp) {
				t.Errorf("FormatUpstreamError(%q) = %q; missing expected keyword %q", c.errMsg, out, exp)
			}
		}
	}
}

func TestFormatEmptyOutputNotice(t *testing.T) {
	notice := FormatEmptyOutputNotice()
	if !strings.Contains(notice, "响应为空") && !strings.Contains(notice, "空内容") {
		t.Errorf("expected empty output notice to mention empty response, got: %s", notice)
	}
}

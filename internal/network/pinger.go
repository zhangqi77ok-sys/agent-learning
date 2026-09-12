package network

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var defaultTransport = &http.Transport{
	TLSClientConfig:     &tls.Config{InsecureSkipVerify: false}, // Default strict
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 20,
	IdleConnTimeout:     30 * time.Second,
	DisableCompression:  true,
}

var insecureTransport = &http.Transport{
	TLSClientConfig:     &tls.Config{InsecureSkipVerify: true}, // Only for loopback
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 20,
	IdleConnTimeout:     30 * time.Second,
	DisableCompression:  true,
}

var defaultClient = &http.Client{
	Timeout:   4 * time.Second,
	Transport: defaultTransport,
}

var insecureClient = &http.Client{
	Timeout:   4 * time.Second,
	Transport: insecureTransport,
}

// PingTarget 真实发起 HTTP 网络探活并测量往返毫秒延迟
func PingTarget(targetURL string) (string, error) {
	trimmed := strings.TrimSpace(targetURL)
	if trimmed == "" {
		return "", fmt.Errorf("empty url")
	}

	// 自动补齐缺失的 HTTP/HTTPS 协议前缀，防止 unsupported protocol scheme 错误
	if !strings.HasPrefix(trimmed, "http://") && !strings.HasPrefix(trimmed, "https://") {
		tempURL := "http://" + trimmed
		u, err := url.Parse(tempURL)
		if err == nil {
			host := strings.ToLower(u.Hostname())
			if host == "localhost" || host == "127.0.0.1" || host == "0.0.0.0" || host == "::1" {
				trimmed = "http://" + trimmed
			} else {
				trimmed = "https://" + trimmed
			}
		} else {
			trimmed = "https://" + trimmed
		}
	}

	u, err := url.Parse(trimmed)
	if err != nil {
		return "", err
	}
	host := strings.ToLower(u.Hostname())
	isLocal := host == "localhost" || host == "127.0.0.1" || host == "0.0.0.0" || host == "::1"

	client := defaultClient
	if isLocal {
		client = insecureClient
	}

	start := time.Now()
	req, err := http.NewRequest("GET", trimmed, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "codex_cli_rs/0.101.0 (Mac OS 26.0.1; arm64) Apple_Terminal/464")
	req.Header.Set("Originator", "codex_cli_rs")
	req.Header.Set("Version", "0.101.0")

	resp, err := client.Do(req)
	duration := time.Since(start)
	if err != nil {
		return "", fmt.Errorf("ping failed: %w", err)
	}
	defer resp.Body.Close()

	// 浅读排空以复用长连接
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))

	if resp.StatusCode >= 500 {
		return "", fmt.Errorf("ping failed: server returned status %d", resp.StatusCode)
	}

	latencyMs := fmt.Sprintf("%dms", duration.Milliseconds())
	return latencyMs, nil
}

package network

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPingTarget_Empty(t *testing.T) {
	_, err := PingTarget("")
	if err == nil {
		t.Fatalf("expected error for empty target, got nil")
	}
}

func TestPingTarget_AutoScheme(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer ts.Close()

	// 传入带 http:// 的测试地址
	latency, err := PingTarget(ts.URL)
	if err != nil {
		t.Fatalf("ping failed: %v", err)
	}
	if !strings.HasSuffix(latency, "ms") {
		t.Errorf("expected latency to end with ms, got: %s", latency)
	}

	// 传入缺少 scheme 的主机名/端口 (如 127.0.0.1:xxxx)
	hostPort := strings.TrimPrefix(ts.URL, "http://")
	// 针对 127.0.0.1 / localhost，应智能补齐 http://，且能成功连通返回有效延迟
	localLatency, pingErr := PingTarget(hostPort)
	if pingErr != nil {
		t.Fatalf("local hostPort without scheme should be completed with http:// and succeed, got error: %v", pingErr)
	}
	if !strings.HasSuffix(localLatency, "ms") {
		t.Errorf("expected local latency to end with ms, got: %s", localLatency)
	}
}

func TestPingTarget_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("502 Bad Gateway"))
	}))
	defer ts.Close()

	_, err := PingTarget(ts.URL)
	if err == nil {
		t.Fatalf("expected error for 502 Bad Gateway, got nil")
	}
	if !strings.Contains(err.Error(), "502") {
		t.Errorf("expected error message to contain 502, got %v", err)
	}
}

func TestPingTarget_UntrustedPublicTLSFails(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer ts.Close()

	// 强制使用非回环的假域名，并利用自定义 Dial 拦截 DNS 解析到本地，验证 TLS 证书拦截
	urlStr := strings.Replace(ts.URL, "127.0.0.1", "untrusted.test.local", 1)
	
	// Hook the dialer
	originalTransport := defaultClient.Transport
	defer func() { defaultClient.Transport = originalTransport }()
	
	customTransport := defaultTransport.Clone()
	customTransport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		// Ignore the addr and always connect to our local ts
		return net.Dial("tcp", ts.Listener.Addr().String())
	}
	defaultClient.Transport = customTransport

	_, err := PingTarget(urlStr)
	if err == nil {
		t.Fatalf("expected ping to fail due to untrusted certificate on non-loopback host, but it succeeded")
	}
	if !strings.Contains(err.Error(), "certificate") && !strings.Contains(err.Error(), "x509") && !strings.Contains(err.Error(), "authority") {
		t.Errorf("expected TLS certificate error, got: %v", err)
	}
}

func TestPingTarget_LoopbackAllowsHTTP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer ts.Close()

	hostPort := strings.TrimPrefix(ts.URL, "http://")
	latency, err := PingTarget(hostPort)
	if err != nil {
		t.Fatalf("expected loopback to succeed, got %v", err)
	}
	if !strings.HasSuffix(latency, "ms") {
		t.Errorf("expected latency to end with ms, got: %s", latency)
	}
}

func TestClassifyPingHost_RejectsLocalhostPrefixSpoof(t *testing.T) {
	// Should not prepend http:// to localhost.evil.com
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	
	hostPort := strings.TrimPrefix(ts.URL, "http://")
	spoofedHostPort := strings.Replace(hostPort, "127.0.0.1", "localhost.evil.com", 1)
	
	originalTransport := defaultClient.Transport
	defer func() { defaultClient.Transport = originalTransport }()
	
	customTransport := defaultTransport.Clone()
	customTransport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return net.Dial("tcp", ts.Listener.Addr().String())
	}
	defaultClient.Transport = customTransport

	_, err := PingTarget(spoofedHostPort)
	if err == nil {
		t.Fatalf("expected ping to spoofed localhost to fail (should use https and hit an http server), but it succeeded")
	}
	if !strings.Contains(err.Error(), "server gave HTTP response to HTTPS client") && !strings.Contains(err.Error(), "tls") {
		t.Errorf("expected HTTPS failure (HTTP response to HTTPS client), got: %v", err)
	}
}

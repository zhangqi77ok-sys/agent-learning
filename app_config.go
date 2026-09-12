package main

import (
	"tiancode/internal/lsp"
	"tiancode/internal/mcp"

	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"tiancode/internal/config"
	"tiancode/internal/network"
)

func (a *App) ListChannels() []config.ChannelConfig {
	if a.channelStore == nil {
		return nil
	}
	return a.channelStore.ListMasked()
}

func (a *App) SaveChannel(cfg config.ChannelConfig) error {
	if a.channelStore == nil {
		return fmt.Errorf("channel store not initialized")
	}
	return a.channelStore.Save(cfg)
}

func (a *App) DeleteChannel(id string) error {
	if a.channelStore == nil {
		return fmt.Errorf("channel store not initialized")
	}
	return a.channelStore.Delete(id)
}

func (a *App) PingChannel(id string) (string, error) {
	if a.channelStore == nil {
		return "", fmt.Errorf("channel store not initialized")
	}
	ch := a.channelStore.Get(id)
	if ch == nil {
		return "", fmt.Errorf("channel [%s] not found", id)
	}
	latency, err := network.PingTarget(ch.Endpoint)
	if err != nil {
		return "", err
	}
	ch.Latency = latency
	_ = a.channelStore.Save(*ch)
	return latency, nil
}

func (a *App) ListMCPs() []config.MCPServerConfig {
	if a.extraStore == nil {
		return nil
	}
	return a.extraStore.ListMCPs()
}

func (a *App) SaveMCP(cfg config.MCPServerConfig) error {
	if a.extraStore == nil {
		return fmt.Errorf("extra store not initialized")
	}
	if err := a.extraStore.SaveMCP(cfg); err != nil {
		return err
	}
	// 动态联动启停 MCP 进程实例
	if a.mcpManager != nil {
		go func() {
			if cfg.Enabled {
				_ = a.mcpManager.StartServer(context.Background(), cfg)
			} else {
				_ = a.mcpManager.StopServer(context.Background(), cfg.ID)
			}
		}()
	}
	return nil
}

// DeleteMCP 从磁盘删除 MCP 配置并停止运行中的实例
func (a *App) DeleteMCP(id string) error {
	if a.extraStore == nil {
		return fmt.Errorf("extra store not initialized")
	}
	if err := a.extraStore.DeleteMCP(id); err != nil {
		return err
	}
	if a.mcpManager != nil {
		go func() {
			_ = a.mcpManager.StopServer(context.Background(), id)
		}()
	}
	return nil
}

// DiagnoseFile 触发指定文件的毫秒级轻量编译器语法诊断
func (a *App) DiagnoseFile(relPath string) (*lsp.DiagnosticReport, error) {
	return lsp.DiagnoseFile(a.workspace, relPath)
}

// TestMCPServer 对指定 MCP 服务执行标准 JSON-RPC 2.0 握手与工具探活
func (a *App) TestMCPServer(id string) (mcp.MCPTestResult, error) {
	if a.mcpManager == nil {
		return mcp.MCPTestResult{Status: "ERROR", Error: "mcp manager not initialized"}, fmt.Errorf("mcp manager not initialized")
	}
	if a.extraStore != nil {
		mcps := a.extraStore.ListMCPs()
		for _, srv := range mcps {
			if srv.ID == id || strings.Contains(strings.ToLower(srv.Name), strings.ToLower(id)) {
				return a.mcpManager.TestServer(context.Background(), srv)
			}
		}
	}
	return mcp.MCPTestResult{
		ID:     id,
		Status: "ERROR",
		Error:  fmt.Sprintf("未找到指定的 MCP 服务配置: [%s]", id),
	}, fmt.Errorf("mcp server [%s] not found", id)
}

func (a *App) ListSkills() []config.SkillConfig {
	if a.extraStore == nil {
		return nil
	}
	return a.extraStore.ListSkills()
}

func (a *App) SaveSkill(cfg config.SkillConfig) error {
	if a.extraStore == nil {
		return fmt.Errorf("extra store not initialized")
	}
	return a.extraStore.SaveSkill(cfg)
}

func (a *App) DeleteSkill(id string) error {
	if a.extraStore == nil {
		return fmt.Errorf("extra store not initialized")
	}
	return a.extraStore.DeleteSkill(id)
}

func (a *App) ListRules() []config.RuleConfig {
	if a.extraStore == nil {
		return nil
	}
	return a.extraStore.ListRules()
}

func (a *App) SaveRule(cfg config.RuleConfig) error {
	if a.extraStore == nil {
		return fmt.Errorf("extra store not initialized")
	}
	return a.extraStore.SaveRule(cfg)
}

func (a *App) DeleteRule(id string) error {
	if a.extraStore == nil {
		return fmt.Errorf("extra store not initialized")
	}
	return a.extraStore.DeleteRule(id)
}
func (a *App) FetchUpstreamModels(endpoint, apiKey string) ([]string, error) {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		endpoint = "https://agentrouter.org/v1"
	} else {
		if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
			if strings.Contains(endpoint, "localhost") || strings.Contains(endpoint, "127.0.0.1") {
				endpoint = "http://" + endpoint
			} else {
				endpoint = "https://" + endpoint
			}
		}
	}
	if apiKey == "" {
		if primary := a.channelStore.GetPrimary(); primary != nil && primary.APIKey != "" {
			apiKey = primary.APIKey
			if endpoint == "https://agentrouter.org/v1" && primary.Endpoint != "" {
				endpoint = primary.Endpoint
			}
		}
	}
	if apiKey == "" {
		return nil, fmt.Errorf("未配置有效 API Key，请先在渠道配置中填写模型 API Key")
	}
	cleanEndpoint := strings.TrimRight(endpoint, "/")
	if strings.HasSuffix(cleanEndpoint, "/models") {
		cleanEndpoint = strings.TrimSuffix(cleanEndpoint, "/models")
	}
	url := cleanEndpoint + "/models"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("User-Agent", "codex_cli_rs/0.101.0 (Mac OS 26.0.1; arm64) Apple_Terminal/464")
	req.Header.Set("Originator", "codex_cli_rs")
	req.Header.Set("Version", "0.101.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("上游模型网关响应错误 (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}

	var data struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	res := make([]string, 0, len(data.Data))
	for _, m := range data.Data {
		res = append(res, m.ID)
	}
	return res, nil
}

package proxy

import (
	"ant-chrome/backend/internal/config"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func ValidateProxyChainConfig(proxyConfig string, proxies []config.BrowserProxy, proxyId string) (bool, string) {
	chain, err := ResolveProxyChain(proxyConfig, proxies, proxyId)
	if err != nil {
		return false, err.Error()
	}
	if chain == nil || strings.TrimSpace(chain.TargetConfig) == "" || isDirectProxyConfig(chain.TargetConfig) {
		return true, ""
	}
	if err := validateResolvedProxyChain(chain); err != nil {
		return false, err.Error()
	}
	return true, ""
}

func RequiresXrayBridgeForChain(proxyConfig string, proxies []config.BrowserProxy, proxyId string) bool {
	chain, err := ResolveProxyChain(proxyConfig, proxies, proxyId)
	if err != nil || chain == nil {
		return false
	}
	if isDirectProxyConfig(chain.TargetConfig) || IsSingBoxProtocol(chain.TargetConfig) {
		return false
	}
	if chain.HasPreProxy() {
		return true
	}
	return RequiresBridge(chain.TargetConfig, nil, "")
}

func AcquireXrayBridgeChain(
	m *XrayManager,
	singboxMgr *SingBoxManager,
	proxyConfig string,
	proxies []config.BrowserProxy,
	proxyId string,
) (string, string, error) {
	return ensureXrayBridgeChain(m, singboxMgr, proxyConfig, proxies, proxyId, true)
}

func EnsureXrayBridgeChain(
	m *XrayManager,
	singboxMgr *SingBoxManager,
	proxyConfig string,
	proxies []config.BrowserProxy,
	proxyId string,
) (string, error) {
	socksURL, _, err := ensureXrayBridgeChain(m, singboxMgr, proxyConfig, proxies, proxyId, false)
	return socksURL, err
}

func EnsureSingBoxBridgeChain(
	m *SingBoxManager,
	xrayMgr *XrayManager,
	proxyConfig string,
	proxies []config.BrowserProxy,
	proxyId string,
) (string, error) {
	if m == nil {
		return "", fmt.Errorf("sing-box 管理器未初始化")
	}

	chain, err := ResolveProxyChain(proxyConfig, proxies, proxyId)
	if err != nil {
		return "", err
	}
	if chain == nil || strings.TrimSpace(chain.TargetConfig) == "" {
		return "", fmt.Errorf("未找到代理节点")
	}

	directProxy, targetOutbound, extraOutbounds, err := buildSingBoxChainPlan(chain, xrayMgr, proxies, false)
	if err != nil {
		return "", err
	}
	if directProxy != "" {
		return directProxy, nil
	}

	key := computeNodeKey(chain.CacheKey())
	if socksURL, reused := m.tryReuseBridge(key); reused {
		return socksURL, nil
	}

	binaryPath, err := m.resolveBinary()
	if err != nil {
		return "", err
	}

	const maxRetries = 3
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		port, err := nextAvailablePort()
		if err != nil {
			lastErr = err
			continue
		}

		cfgPath, err := buildSingBoxRuntimeConfig(m, key, targetOutbound, extraOutbounds, port)
		if err != nil {
			return "", fmt.Errorf("sing-box 配置生成失败: %w", err)
		}

		cmd := exec.Command(binaryPath, "run", "-c", cfgPath)
		hideWindow(cmd)
		cmd.Dir = filepath.Dir(cfgPath)
		stderrPath := filepath.Join(filepath.Dir(cfgPath), "singbox-stderr.log")
		stderrFile, _ := os.Create(stderrPath)
		if stderrFile != nil {
			cmd.Stderr = stderrFile
		}

		if err := cmd.Start(); err != nil {
			if stderrFile != nil {
				stderrFile.Close()
			}
			lastErr = err
			continue
		}

		bridge := &SingBoxBridge{
			NodeKey: key,
			Port:    port,
			Cmd:     cmd,
			Pid:     cmd.Process.Pid,
			Running: true,
		}

		if err := waitPortReady("127.0.0.1", port, 10*time.Second); err != nil {
			if stderrFile != nil {
				stderrFile.Close()
			}
			bridge.Stopping = true
			m.stopBridgeProcess(bridge)
			bridge.Running = false
			bridge.Pid = 0
			bridge.LastError = err.Error()
			lastErr = err
			time.Sleep(200 * time.Millisecond)
			continue
		}

		if stderrFile != nil {
			stderrFile.Close()
		}

		if socksURL, reused := m.registerBridge(key, bridge); reused {
			bridge.Stopping = true
			m.stopBridgeProcess(bridge)
			return socksURL, nil
		}

		go m.watchBridge(bridge, key)
		return fmt.Sprintf("socks5://127.0.0.1:%d", port), nil
	}

	return "", fmt.Errorf("sing-box 启动失败（已重试 %d 次）: %w", maxRetries, lastErr)
}

func ensureXrayBridgeChain(
	m *XrayManager,
	singboxMgr *SingBoxManager,
	proxyConfig string,
	proxies []config.BrowserProxy,
	proxyId string,
	pin bool,
) (string, string, error) {
	if m == nil {
		return "", "", fmt.Errorf("xray 管理器未初始化")
	}

	chain, err := ResolveProxyChain(proxyConfig, proxies, proxyId)
	if err != nil {
		return "", "", err
	}
	if chain == nil || strings.TrimSpace(chain.TargetConfig) == "" {
		return "", "", fmt.Errorf("未找到代理节点")
	}

	directProxy, targetOutbound, extraOutbounds, err := buildXrayChainPlan(chain, singboxMgr, proxies, false)
	if err != nil {
		return "", "", err
	}
	if directProxy != "" {
		return directProxy, "", nil
	}

	key := computeNodeKey(chain.CacheKey())
	if socksURL, reused := m.tryReuseBridge(key, pin); reused {
		return socksURL, key, nil
	}

	binaryPath, err := m.resolveBinary()
	if err != nil {
		return "", "", err
	}

	const maxLaunchRetries = 3
	var lastErr error
	for attempt := 1; attempt <= maxLaunchRetries; attempt++ {
		port, err := nextAvailablePort()
		if err != nil {
			lastErr = err
			continue
		}

		cfgPath, err := buildXrayRuntimeConfig(m, key, targetOutbound, extraOutbounds, port, chain.DNSServers)
		if err != nil {
			return "", "", err
		}

		cmd := exec.Command(binaryPath, "run", "-c", cfgPath)
		hideWindow(cmd)
		cmd.Dir = filepath.Dir(cfgPath)
		stderrPath := filepath.Join(filepath.Dir(cfgPath), "xray-stderr.log")
		stderrFile, _ := os.Create(stderrPath)
		if stderrFile != nil {
			cmd.Stderr = stderrFile
		}
		if err := cmd.Start(); err != nil {
			if stderrFile != nil {
				stderrFile.Close()
			}
			lastErr = err
			continue
		}

		bridge := &XrayBridge{
			NodeKey:    key,
			Port:       port,
			Cmd:        cmd,
			Pid:        cmd.Process.Pid,
			Running:    true,
			RefCount:   0,
			LastUsedAt: time.Now(),
		}

		if err := waitPortReady("127.0.0.1", port, 10*time.Second); err != nil {
			if stderrFile != nil {
				stderrFile.Close()
			}
			bridge.Stopping = true
			m.stopBridgeProcess(bridge)
			bridge.Running = false
			bridge.Pid = 0
			bridge.LastError = err.Error()
			lastErr = err
			time.Sleep(200 * time.Millisecond)
			continue
		}
		if stderrFile != nil {
			stderrFile.Close()
		}

		if socksURL, reused := m.registerBridge(key, bridge, pin); reused {
			bridge.Stopping = true
			m.stopBridgeProcess(bridge)
			return socksURL, key, nil
		}

		go m.watchBridge(bridge, key)
		return fmt.Sprintf("socks5://127.0.0.1:%d", port), key, nil
	}

	return "", "", fmt.Errorf("xray 启动失败（已重试 %d 次）: %w", maxLaunchRetries, lastErr)
}

func validateResolvedProxyChain(chain *ResolvedProxyChain) error {
	if chain == nil || strings.TrimSpace(chain.TargetConfig) == "" || isDirectProxyConfig(chain.TargetConfig) {
		return nil
	}
	if IsSingBoxProtocol(chain.TargetConfig) {
		_, _, _, err := buildSingBoxChainPlan(chain, nil, nil, true)
		return err
	}
	_, _, _, err := buildXrayChainPlan(chain, nil, nil, true)
	return err
}

func buildXrayChainPlan(
	chain *ResolvedProxyChain,
	singboxMgr *SingBoxManager,
	proxies []config.BrowserProxy,
	validateOnly bool,
) (string, map[string]interface{}, []interface{}, error) {
	standardProxy, outbound, err := ParseProxyNode(chain.TargetConfig)
	if err != nil {
		return "", nil, nil, fmt.Errorf("代理配置解析失败: %w", err)
	}

	var targetOutbound map[string]interface{}
	if standardProxy != "" {
		if !chain.HasPreProxy() {
			return standardProxy, nil, nil, nil
		}
		targetOutbound, err = buildXrayStandardOutbound(standardProxy, "proxy-out")
		if err != nil {
			return "", nil, nil, err
		}
	} else if outbound != nil {
		targetOutbound = cloneMap(outbound)
		targetOutbound["tag"] = "proxy-out"
	} else {
		return "", nil, nil, fmt.Errorf("代理配置无效")
	}

	extraOutbounds := make([]interface{}, 0, 1)
	if chain.HasPreProxy() {
		preOutbound, err := buildXrayPreOutbound(chain, singboxMgr, proxies, validateOnly)
		if err != nil {
			return "", nil, nil, err
		}
		targetOutbound["proxySettings"] = map[string]interface{}{"tag": "pre-out"}
		extraOutbounds = append(extraOutbounds, preOutbound)
	}

	return "", targetOutbound, extraOutbounds, nil
}

func buildSingBoxChainPlan(
	chain *ResolvedProxyChain,
	xrayMgr *XrayManager,
	proxies []config.BrowserProxy,
	validateOnly bool,
) (string, map[string]interface{}, []interface{}, error) {
	outbound, err := BuildSingBoxOutbound(chain.TargetConfig)
	if err != nil {
		return "", nil, nil, fmt.Errorf("代理配置解析失败: %w", err)
	}

	targetOutbound := cloneMap(outbound)
	targetOutbound["tag"] = "proxy-out"
	extraOutbounds := make([]interface{}, 0, 1)

	if chain.HasPreProxy() {
		preOutbound, err := buildSingBoxPreOutbound(chain, xrayMgr, proxies, validateOnly)
		if err != nil {
			return "", nil, nil, err
		}
		targetOutbound["detour"] = "pre-out"
		extraOutbounds = append(extraOutbounds, preOutbound)
	}

	return "", targetOutbound, extraOutbounds, nil
}

func buildXrayPreOutbound(
	chain *ResolvedProxyChain,
	singboxMgr *SingBoxManager,
	proxies []config.BrowserProxy,
	validateOnly bool,
) (map[string]interface{}, error) {
	if !chain.HasPreProxy() {
		return nil, nil
	}
	if isStandardProxyConfig(chain.PreConfig) {
		return buildXrayStandardOutbound(chain.PreConfig, "pre-out")
	}
	if IsSingBoxProtocol(chain.PreConfig) {
		if validateOnly {
			return buildXrayStandardOutbound("socks5://127.0.0.1:1080", "pre-out")
		}
		if singboxMgr == nil {
			return nil, fmt.Errorf("sing-box 管理器未初始化")
		}
		socksURL, err := EnsureSingBoxBridgeChain(singboxMgr, nil, chain.PreConfig, proxies, chain.PreProxy.ProxyId)
		if err != nil {
			return nil, err
		}
		return buildXrayStandardOutbound(socksURL, "pre-out")
	}

	standardProxy, outbound, err := ParseProxyNode(chain.PreConfig)
	if err != nil {
		return nil, fmt.Errorf("前置节点解析失败: %w", err)
	}
	if standardProxy != "" {
		return buildXrayStandardOutbound(standardProxy, "pre-out")
	}
	if outbound == nil {
		return nil, fmt.Errorf("前置节点配置无效")
	}
	preOutbound := cloneMap(outbound)
	preOutbound["tag"] = "pre-out"
	return preOutbound, nil
}

func buildSingBoxPreOutbound(
	chain *ResolvedProxyChain,
	xrayMgr *XrayManager,
	proxies []config.BrowserProxy,
	validateOnly bool,
) (map[string]interface{}, error) {
	if !chain.HasPreProxy() {
		return nil, nil
	}
	if isStandardProxyConfig(chain.PreConfig) {
		return buildSingBoxStandardOutbound(chain.PreConfig, "pre-out")
	}
	if IsSingBoxProtocol(chain.PreConfig) {
		outbound, err := BuildSingBoxOutbound(chain.PreConfig)
		if err != nil {
			return nil, fmt.Errorf("前置节点解析失败: %w", err)
		}
		preOutbound := cloneMap(outbound)
		preOutbound["tag"] = "pre-out"
		return preOutbound, nil
	}
	if validateOnly {
		return buildSingBoxStandardOutbound("socks5://127.0.0.1:1080", "pre-out")
	}
	if xrayMgr == nil {
		return nil, fmt.Errorf("xray 管理器未初始化")
	}
	socksURL, err := EnsureXrayBridgeChain(xrayMgr, nil, chain.PreConfig, proxies, chain.PreProxy.ProxyId)
	if err != nil {
		return nil, err
	}
	return buildSingBoxStandardOutbound(socksURL, "pre-out")
}

func buildXrayRuntimeConfig(
	m *XrayManager,
	key string,
	outbound map[string]interface{},
	extraOutbounds []interface{},
	port int,
	dnsServers string,
) (string, error) {
	baseDir := m.resolveWorkdir(key)
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", err
	}

	outbounds := make([]interface{}, 0, len(extraOutbounds)+3)
	outbounds = append(outbounds, outbound)
	outbounds = append(outbounds, extraOutbounds...)
	outbounds = append(outbounds,
		map[string]interface{}{
			"protocol": "direct",
			"tag":      "direct",
		},
		map[string]interface{}{
			"protocol": "blackhole",
			"tag":      "block",
		},
	)

	cfg := map[string]interface{}{
		"log": map[string]interface{}{
			"loglevel": "info",
			"error":    filepath.Join(baseDir, "xray-error.log"),
		},
		"inbounds": []interface{}{
			map[string]interface{}{
				"tag":      "socks-in",
				"port":     port,
				"listen":   "127.0.0.1",
				"protocol": "socks",
				"settings": map[string]interface{}{
					"udp": true,
				},
				"sniffing": map[string]interface{}{
					"enabled": false,
				},
			},
		},
		"outbounds": outbounds,
		"routing": map[string]interface{}{
			"rules": []interface{}{
				map[string]interface{}{
					"type":        "field",
					"inboundTag":  []string{"socks-in"},
					"outboundTag": "proxy-out",
				},
			},
		},
	}
	if dnsCfg := parseDnsConfig(dnsServers); dnsCfg != nil {
		cfg["dns"] = dnsCfg
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", err
	}

	cfgPath := filepath.Join(baseDir, "xray-config.json")
	if err := os.WriteFile(cfgPath, data, 0644); err != nil {
		return "", err
	}
	return cfgPath, nil
}

func buildSingBoxRuntimeConfig(
	m *SingBoxManager,
	key string,
	outbound map[string]interface{},
	extraOutbounds []interface{},
	port int,
) (string, error) {
	baseDir := m.resolveWorkdir(key)
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", err
	}

	outbounds := make([]interface{}, 0, len(extraOutbounds)+2)
	outbounds = append(outbounds, outbound)
	outbounds = append(outbounds, extraOutbounds...)
	outbounds = append(outbounds, map[string]interface{}{
		"type": "direct",
		"tag":  "direct",
	})

	cfg := map[string]interface{}{
		"log": map[string]interface{}{
			"level":     "info",
			"output":    filepath.Join(baseDir, "singbox.log"),
			"timestamp": true,
		},
		"inbounds": []interface{}{
			map[string]interface{}{
				"type":        "socks",
				"tag":         "socks-in",
				"listen":      "127.0.0.1",
				"listen_port": port,
			},
		},
		"outbounds": outbounds,
		"route": map[string]interface{}{
			"rules": []interface{}{
				map[string]interface{}{
					"inbound":  []string{"socks-in"},
					"outbound": "proxy-out",
				},
			},
		},
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", err
	}
	cfgPath := filepath.Join(baseDir, "singbox-config.json")
	if err := os.WriteFile(cfgPath, data, 0644); err != nil {
		return "", err
	}
	return cfgPath, nil
}

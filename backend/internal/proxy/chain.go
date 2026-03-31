package proxy

import (
	"ant-chrome/backend/internal/config"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type ResolvedProxyChain struct {
	Target       config.BrowserProxy
	TargetConfig string
	DNSServers   string
	PreProxy     *config.BrowserProxy
	PreConfig    string
}

func ResolveProxyChain(proxyConfig string, proxies []config.BrowserProxy, proxyId string) (*ResolvedProxyChain, error) {
	targetConfig := strings.TrimSpace(proxyConfig)
	dnsServers := ""
	target := config.BrowserProxy{
		ProxyId:     strings.TrimSpace(proxyId),
		ProxyConfig: targetConfig,
	}

	if proxyId = strings.TrimSpace(proxyId); proxyId != "" {
		item, ok := findProxyByID(proxies, proxyId)
		if !ok {
			if targetConfig == "" {
				return nil, fmt.Errorf("代理链路不可用：代理池节点已不存在（proxyId=%s）。可能因订阅刷新后节点下线或被删除，请重新选择代理后再启动。", proxyId)
			}
		} else {
			target = item
			targetConfig = strings.TrimSpace(item.ProxyConfig)
			dnsServers = strings.TrimSpace(item.DnsServers)
		}
	}

	if targetConfig == "" {
		return nil, fmt.Errorf("代理配置为空")
	}

	chain := &ResolvedProxyChain{
		Target:       target,
		TargetConfig: normalizeNodeScheme(targetConfig),
		DNSServers:   dnsServers,
	}

	preProxyId := strings.TrimSpace(target.PreProxyId)
	if preProxyId == "" {
		return chain, nil
	}
	if target.ProxyId == "" {
		return nil, fmt.Errorf("当前代理未绑定代理池节点，无法使用前置节点")
	}
	if strings.EqualFold(preProxyId, target.ProxyId) {
		return nil, fmt.Errorf("前置节点不能指向自己")
	}
	if isDirectProxyConfig(chain.TargetConfig) {
		return nil, fmt.Errorf("直连节点不能配置前置节点")
	}

	preProxy, ok := findProxyByID(proxies, preProxyId)
	if !ok {
		return nil, fmt.Errorf("前置节点已不存在，请重新选择")
	}
	if strings.TrimSpace(preProxy.ProxyConfig) == "" {
		return nil, fmt.Errorf("前置节点配置为空")
	}
	if strings.TrimSpace(preProxy.PreProxyId) != "" {
		return nil, fmt.Errorf("前置节点不能再配置前置节点")
	}

	chain.PreProxy = &preProxy
	chain.PreConfig = normalizeNodeScheme(strings.TrimSpace(preProxy.ProxyConfig))
	return chain, nil
}

func (c *ResolvedProxyChain) HasPreProxy() bool {
	return c != nil && c.PreProxy != nil && strings.TrimSpace(c.PreConfig) != ""
}

func (c *ResolvedProxyChain) CacheKey() string {
	if c == nil {
		return ""
	}
	parts := []string{
		strings.TrimSpace(c.Target.ProxyId),
		strings.TrimSpace(c.TargetConfig),
		strings.TrimSpace(c.DNSServers),
	}
	if c.PreProxy != nil {
		parts = append(parts, strings.TrimSpace(c.PreProxy.ProxyId), strings.TrimSpace(c.PreConfig))
	}
	return strings.Join(parts, "\x00")
}

func findProxyByID(proxies []config.BrowserProxy, proxyId string) (config.BrowserProxy, bool) {
	for _, item := range proxies {
		if strings.EqualFold(strings.TrimSpace(item.ProxyId), strings.TrimSpace(proxyId)) {
			return item, true
		}
	}
	return config.BrowserProxy{}, false
}

func isDirectProxyConfig(proxyConfig string) bool {
	return strings.EqualFold(strings.TrimSpace(proxyConfig), "direct://")
}

func isStandardProxyConfig(proxyConfig string) bool {
	lower := strings.ToLower(strings.TrimSpace(proxyConfig))
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "socks5://")
}

type standardProxyConfig struct {
	Scheme   string
	Host     string
	Port     int
	Username string
	Password string
}

func parseStandardProxyConfig(proxyConfig string) (*standardProxyConfig, error) {
	raw := strings.TrimSpace(proxyConfig)
	if !isStandardProxyConfig(raw) {
		return nil, fmt.Errorf("unsupported standard proxy config")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("standard proxy parse failed: %w", err)
	}
	port, _ := strconv.Atoi(parsed.Port())
	if parsed.Hostname() == "" || port <= 0 {
		return nil, fmt.Errorf("standard proxy is incomplete")
	}
	cfg := &standardProxyConfig{
		Scheme: strings.ToLower(parsed.Scheme),
		Host:   parsed.Hostname(),
		Port:   port,
	}
	if parsed.User != nil {
		cfg.Username = parsed.User.Username()
		cfg.Password, _ = parsed.User.Password()
	}
	return cfg, nil
}

func buildXrayStandardOutbound(proxyConfig string, tag string) (map[string]interface{}, error) {
	cfg, err := parseStandardProxyConfig(proxyConfig)
	if err != nil {
		return nil, err
	}

	switch cfg.Scheme {
	case "http", "https":
		settings := map[string]interface{}{
			"address": cfg.Host,
			"port":    cfg.Port,
		}
		if cfg.Username != "" {
			settings["user"] = cfg.Username
			settings["pass"] = cfg.Password
		}
		outbound := map[string]interface{}{
			"protocol": "http",
			"tag":      tag,
			"settings": settings,
		}
		if cfg.Scheme == "https" {
			outbound["streamSettings"] = map[string]interface{}{
				"security":    "tls",
				"tlsSettings": map[string]interface{}{},
			}
		}
		return outbound, nil
	case "socks5":
		server := map[string]interface{}{
			"address": cfg.Host,
			"port":    cfg.Port,
		}
		if cfg.Username != "" {
			server["users"] = []interface{}{
				map[string]interface{}{
					"user": cfg.Username,
					"pass": cfg.Password,
				},
			}
		}
		return map[string]interface{}{
			"protocol": "socks",
			"tag":      tag,
			"settings": map[string]interface{}{
				"servers": []interface{}{server},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported standard proxy scheme: %s", cfg.Scheme)
	}
}

func buildSingBoxStandardOutbound(proxyConfig string, tag string) (map[string]interface{}, error) {
	cfg, err := parseStandardProxyConfig(proxyConfig)
	if err != nil {
		return nil, err
	}

	switch cfg.Scheme {
	case "http", "https":
		outbound := map[string]interface{}{
			"type":        "http",
			"tag":         tag,
			"server":      cfg.Host,
			"server_port": cfg.Port,
		}
		if cfg.Username != "" {
			outbound["username"] = cfg.Username
			outbound["password"] = cfg.Password
		}
		if cfg.Scheme == "https" {
			outbound["tls"] = map[string]interface{}{
				"enabled": true,
			}
		}
		return outbound, nil
	case "socks5":
		outbound := map[string]interface{}{
			"type":        "socks",
			"tag":         tag,
			"server":      cfg.Host,
			"server_port": cfg.Port,
			"version":     "5",
		}
		if cfg.Username != "" {
			outbound["username"] = cfg.Username
			outbound["password"] = cfg.Password
		}
		return outbound, nil
	default:
		return nil, fmt.Errorf("unsupported standard proxy scheme: %s", cfg.Scheme)
	}
}

func cloneMap(in map[string]interface{}) map[string]interface{} {
	if in == nil {
		return nil
	}
	out := make(map[string]interface{}, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}


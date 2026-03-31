package proxy

import (
	"ant-chrome/backend/internal/config"
	"testing"
)

func TestResolveProxyChainWithPreProxy(t *testing.T) {
	proxies := []config.BrowserProxy{
		{ProxyId: "hk", ProxyName: "HK", ProxyConfig: "trojan://password@example.com:443"},
		{ProxyId: "us", ProxyName: "US", ProxyConfig: "ss://YWVzLTI1Ni1nY206cGFzc0BleGFtcGxlLmNvbTo4Mzg4", PreProxyId: "hk"},
	}

	chain, err := ResolveProxyChain("", proxies, "us")
	if err != nil {
		t.Fatalf("ResolveProxyChain returned error: %v", err)
	}
	if chain == nil || !chain.HasPreProxy() {
		t.Fatalf("expected pre proxy to be resolved")
	}
	if chain.PreProxy == nil || chain.PreProxy.ProxyId != "hk" {
		t.Fatalf("unexpected pre proxy: %+v", chain.PreProxy)
	}
}

func TestValidateProxyChainConfigRejectsSelfReference(t *testing.T) {
	proxies := []config.BrowserProxy{
		{ProxyId: "loop", ProxyConfig: "http://127.0.0.1:7890", PreProxyId: "loop"},
	}

	ok, msg := ValidateProxyChainConfig("", proxies, "loop")
	if ok {
		t.Fatalf("expected self reference to fail")
	}
	if msg == "" {
		t.Fatalf("expected validation message")
	}
}

func TestValidateProxyChainConfigRejectsNestedPreProxy(t *testing.T) {
	proxies := []config.BrowserProxy{
		{ProxyId: "local", ProxyConfig: "http://127.0.0.1:7890", PreProxyId: "hk"},
		{ProxyId: "hk", ProxyConfig: "trojan://password@example.com:443", PreProxyId: "backup"},
		{ProxyId: "backup", ProxyConfig: "http://127.0.0.1:8080"},
	}

	ok, msg := ValidateProxyChainConfig("", proxies, "local")
	if ok {
		t.Fatalf("expected nested pre proxy to fail")
	}
	if msg == "" {
		t.Fatalf("expected validation message")
	}
}

func TestRequiresXrayBridgeForStandardProxyWithPreProxy(t *testing.T) {
	proxies := []config.BrowserProxy{
		{ProxyId: "local", ProxyConfig: "http://127.0.0.1:7890"},
		{ProxyId: "target", ProxyConfig: "socks5://127.0.0.1:1080", PreProxyId: "local"},
	}

	if !RequiresXrayBridgeForChain("", proxies, "target") {
		t.Fatalf("expected standard proxy with pre proxy to require xray bridge")
	}
}

func TestValidateProxyChainConfigAllowsCrossEngineSingleHop(t *testing.T) {
	proxies := []config.BrowserProxy{
		{ProxyId: "hk", ProxyConfig: "trojan://password@example.com:443"},
		{ProxyId: "us", ProxyConfig: "anytls://password@example.com:8443?sni=example.com", PreProxyId: "hk"},
	}

	ok, msg := ValidateProxyChainConfig("", proxies, "us")
	if !ok {
		t.Fatalf("expected cross-engine single hop to pass, msg=%s", msg)
	}
}
